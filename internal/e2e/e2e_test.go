package e2e

import (
	"booking/internal/domain"
	middleware "booking/internal/handlers"
	authhandler "booking/internal/handlers/auth"
	bookinghandler "booking/internal/handlers/booking"
	roomhandler "booking/internal/handlers/room"
	schedulehandler "booking/internal/handlers/schedule"
	slothandler "booking/internal/handlers/slot"
	authusecase "booking/internal/usecase/auth"
	bookingusecase "booking/internal/usecase/booking"
	roomusecase "booking/internal/usecase/room"
	scheduleusecase "booking/internal/usecase/schedule"
	slotusecase "booking/internal/usecase/slot"
	"booking/internal/utils"
	"booking/internal/utils/customErrors"
	jwtservice "booking/internal/utils/jwtservice"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type memoryStore struct {
	users     map[string]domain.User
	rooms     map[string]domain.Room
	schedules map[string]domain.Schedule
	slots     map[string]domain.Slot
	bookings  map[string]domain.Booking

	nextRoomID     int
	nextScheduleID int
	nextSlotID     int
	nextBookingID  int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		users:     make(map[string]domain.User),
		rooms:     make(map[string]domain.Room),
		schedules: make(map[string]domain.Schedule),
		slots:     make(map[string]domain.Slot),
		bookings:  make(map[string]domain.Booking),
	}
}

func (m *memoryStore) CreateUser(_ context.Context, email, passwordHash, role string) (*domain.User, error) {
	user := domain.User{
		UserID:       fmt.Sprintf("user-%d", len(m.users)+1),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}
	m.users[email] = user
	return &user, nil
}

func (m *memoryStore) GetUserByEmail(_ context.Context, email string) (*domain.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, customErrors.ErrUserNotFound
	}
	return &user, nil
}

func (m *memoryStore) CreateRoom(_ context.Context, name, description string, capacity int) (*domain.Room, error) {
	m.nextRoomID++
	room := domain.Room{
		RoomID:      fmt.Sprintf("room-%d", m.nextRoomID),
		Name:        name,
		Description: description,
		Capacity:    capacity,
		CreatedAt:   time.Now().UTC(),
	}
	m.rooms[room.RoomID] = room
	return &room, nil
}

func (m *memoryStore) GetRooms(_ context.Context) ([]domain.Room, error) {
	rooms := make([]domain.Room, 0, len(m.rooms))
	for _, room := range m.rooms {
		rooms = append(rooms, room)
	}
	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].CreatedAt.After(rooms[j].CreatedAt)
	})
	return rooms, nil
}

func (m *memoryStore) GetRoom(_ context.Context, roomID string) (*domain.Room, error) {
	room, ok := m.rooms[roomID]
	if !ok {
		return nil, customErrors.ErrRoomNotFound
	}
	return &room, nil
}

func (m *memoryStore) ScheduleExistsByRoom(_ context.Context, roomID string) (bool, error) {
	_, ok := m.schedules[roomID]
	return ok, nil
}

func (m *memoryStore) CreateSchedule(_ context.Context, roomID string, daysOfWeek []int, startTime, endTime string) (*domain.Schedule, error) {
	m.nextScheduleID++
	s := domain.Schedule{
		ScheduleID: fmt.Sprintf("schedule-%d", m.nextScheduleID),
		RoomID:     roomID,
		DaysOfWeek: append([]int(nil), daysOfWeek...),
		StartTime:  startTime,
		EndTime:    endTime,
		CreatedAt:  time.Now().UTC(),
	}
	m.schedules[roomID] = s
	return &s, nil
}

func (m *memoryStore) EnsureSlotsForDate(_ context.Context, roomID string, dateUTC time.Time) error {
	schedule, ok := m.schedules[roomID]
	if !ok {
		return nil
	}

	if !containsDay(schedule.DaysOfWeek, isoWeekday(dateUTC)) {
		return nil
	}

	startTOD, err := parseTOD(schedule.StartTime)
	if err != nil {
		return err
	}
	endTOD, err := parseTOD(schedule.EndTime)
	if err != nil {
		return err
	}

	startAt := time.Date(dateUTC.Year(), dateUTC.Month(), dateUTC.Day(), startTOD.Hour(), startTOD.Minute(), 0, 0, time.UTC)
	endAt := time.Date(dateUTC.Year(), dateUTC.Month(), dateUTC.Day(), endTOD.Hour(), endTOD.Minute(), 0, 0, time.UTC)

	for current := startAt; current.Before(endAt); current = current.Add(30 * time.Minute) {
		next := current.Add(30 * time.Minute)
		if next.After(endAt) {
			break
		}

		if m.findSlotID(roomID, current, next) != "" {
			continue
		}

		m.nextSlotID++
		slot := domain.Slot{
			SlotID: fmt.Sprintf("slot-%d", m.nextSlotID),
			RoomID: roomID,
			Start:  current,
			End:    next,
		}
		m.slots[slot.SlotID] = slot
	}

	return nil
}

func (m *memoryStore) ListSlotsByDate(_ context.Context, roomID string, dateUTC time.Time) ([]domain.Slot, error) {
	startOfDay := time.Date(dateUTC.Year(), dateUTC.Month(), dateUTC.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	slots := make([]domain.Slot, 0)
	for _, slot := range m.slots {
		if slot.RoomID != roomID {
			continue
		}
		if slot.Start.Before(startOfDay) || !slot.Start.Before(endOfDay) {
			continue
		}
		if m.hasActiveBooking(slot.SlotID) {
			continue
		}
		slots = append(slots, slot)
	}

	sort.Slice(slots, func(i, j int) bool {
		return slots[i].Start.Before(slots[j].Start)
	})

	return slots, nil
}

func (m *memoryStore) GetSlotByID(_ context.Context, slotID string) (*domain.Slot, error) {
	slot, ok := m.slots[slotID]
	if !ok {
		return nil, customErrors.ErrSlotNotFound
	}
	return &slot, nil
}

func (m *memoryStore) CreateBooking(_ context.Context, slotID, userID string, conferenceLink string) (*domain.Booking, error) {
	if m.hasActiveBooking(slotID) {
		return nil, customErrors.ErrSlotAlreadyBooked
	}

	m.nextBookingID++
	b := domain.Booking{
		BookingID:      fmt.Sprintf("booking-%d", m.nextBookingID),
		SlotID:         slotID,
		UserID:         userID,
		Status:         "active",
		ConferenceLink: conferenceLink,
		CreatedAt:      time.Now().UTC(),
	}
	m.bookings[b.BookingID] = b
	return &b, nil
}

func (m *memoryStore) GetBookingByID(_ context.Context, bookingID string) (*domain.Booking, error) {
	b, ok := m.bookings[bookingID]
	if !ok {
		return nil, customErrors.ErrBookingNotFound
	}
	return &b, nil
}

func (m *memoryStore) CancelBooking(_ context.Context, bookingID string) (*domain.Booking, error) {
	b, ok := m.bookings[bookingID]
	if !ok {
		return nil, customErrors.ErrBookingNotFound
	}
	b.Status = "cancelled"
	m.bookings[bookingID] = b
	return &b, nil
}

func (m *memoryStore) ListBookings(_ context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	list := make([]domain.Booking, 0, len(m.bookings))
	for _, b := range m.bookings {
		list = append(list, b)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	total := len(list)
	offset := (page - 1) * pageSize
	if offset >= total {
		return []domain.Booking{}, total, nil
	}
	end := offset + pageSize
	if end > total {
		end = total
	}

	return list[offset:end], total, nil
}

func (m *memoryStore) MyBookings(_ context.Context, userID string) ([]domain.Booking, error) {
	now := time.Now().UTC()
	bookings := make([]domain.Booking, 0)
	for _, b := range m.bookings {
		if b.UserID != userID {
			continue
		}
		slot, ok := m.slots[b.SlotID]
		if !ok {
			continue
		}
		if slot.Start.Before(now) {
			continue
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

func (m *memoryStore) hasActiveBooking(slotID string) bool {
	for _, b := range m.bookings {
		if b.SlotID == slotID && b.Status == "active" {
			return true
		}
	}
	return false
}

func (m *memoryStore) findSlotID(roomID string, startAt, endAt time.Time) string {
	for id, slot := range m.slots {
		if slot.RoomID == roomID && slot.Start.Equal(startAt) && slot.End.Equal(endAt) {
			return id
		}
	}
	return ""
}

func containsDay(days []int, want int) bool {
	for _, d := range days {
		if d == want {
			return true
		}
	}
	return false
}

func isoWeekday(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func parseTOD(v string) (time.Time, error) {
	if t, err := time.Parse("15:04:05", v); err == nil {
		return t, nil
	}
	return time.Parse("15:04", v)
}

func buildRouterForE2E() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	store := newMemoryStore()
	jwtSvc := jwtservice.New("e2e-secret", time.Hour)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	authUC := authusecase.New(store, jwtSvc)
	authH := authhandler.New(log, authUC)
	roomUC := roomusecase.New(store)
	roomH := roomhandler.New(log, roomUC)
	scheduleUC := scheduleusecase.New(store)
	scheduleH := schedulehandler.New(log, scheduleUC)
	slotUC := slotusecase.New(store)
	slotH := slothandler.New(log, slotUC)
	bookingUC := bookingusecase.New(store)
	bookingH := bookinghandler.New(log, bookingUC)

	r.POST("/dummyLogin", authH.DummyLogin())

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth(jwtSvc))

	roomsGroup := protected.Group("/rooms")
	roomsGroup.POST("/create", middleware.RequireRole(utils.RoleAdmin), roomH.CreateRoom())
	roomsGroup.POST("/:roomId/schedule/create", middleware.RequireRole(utils.RoleAdmin), scheduleH.CreateSchedule())
	roomsGroup.GET("/:roomId/slots/list", slotH.ListSlots())

	bookingsGroup := protected.Group("/bookings")
	bookingsGroup.POST("/create", middleware.RequireRole(utils.RoleUser), bookingH.CreateBooking())
	bookingsGroup.POST("/:bookingId/cancel", middleware.RequireRole(utils.RoleUser), bookingH.CancelBooking())

	return r
}

func TestE2E_CreateRoomScheduleBooking(t *testing.T) {
	r := buildRouterForE2E()

	adminToken := mustDummyToken(t, r, utils.RoleAdmin)
	date := time.Now().UTC().Add(24 * time.Hour)
	dateStr := date.Format("2006-01-02")
	weekday := isoWeekday(date)

	roomResp := doJSON(t, r, http.MethodPost, "/rooms/create", adminToken, map[string]interface{}{
		"name":        "Room A",
		"description": "Main room",
		"capacity":    6,
	})
	if roomResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 create room, got %d body=%s", roomResp.Code, roomResp.Body.String())
	}

	var roomWrap struct {
		Room domain.Room `json:"room"`
	}
	mustDecode(t, roomResp.Body.Bytes(), &roomWrap)
	if roomWrap.Room.RoomID == "" {
		t.Fatal("expected non-empty room id")
	}

	scheduleResp := doJSON(t, r, http.MethodPost, "/rooms/"+roomWrap.Room.RoomID+"/schedule/create", adminToken, map[string]interface{}{
		"daysOfWeek": []int{weekday},
		"startTime":  "09:00",
		"endTime":    "10:00",
	})
	if scheduleResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 create schedule, got %d body=%s", scheduleResp.Code, scheduleResp.Body.String())
	}

	slotsReq := httptest.NewRequest(http.MethodGet, "/rooms/"+roomWrap.Room.RoomID+"/slots/list?date="+dateStr, nil)
	slotsReq.Header.Set("Authorization", "Bearer "+adminToken)
	slotsResp := httptest.NewRecorder()
	r.ServeHTTP(slotsResp, slotsReq)
	if slotsResp.Code != http.StatusOK {
		t.Fatalf("expected 200 list slots, got %d body=%s", slotsResp.Code, slotsResp.Body.String())
	}

	var slotsWrap struct {
		Slots []domain.Slot `json:"slots"`
	}
	mustDecode(t, slotsResp.Body.Bytes(), &slotsWrap)
	if len(slotsWrap.Slots) == 0 {
		t.Fatal("expected at least one generated slot")
	}

	userToken := mustDummyToken(t, r, utils.RoleUser)
	bookingResp := doJSON(t, r, http.MethodPost, "/bookings/create", userToken, map[string]interface{}{
		"slotId":               slotsWrap.Slots[0].SlotID,
		"createConferenceLink": true,
	})
	if bookingResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 create booking, got %d body=%s", bookingResp.Code, bookingResp.Body.String())
	}

	var bookingWrap struct {
		Booking domain.Booking `json:"booking"`
	}
	mustDecode(t, bookingResp.Body.Bytes(), &bookingWrap)
	if bookingWrap.Booking.Status != "active" {
		t.Fatalf("expected booking status active, got %q", bookingWrap.Booking.Status)
	}
	if bookingWrap.Booking.ConferenceLink == "" {
		t.Fatal("expected conference link to be filled")
	}
}

func TestE2E_CancelBooking(t *testing.T) {
	r := buildRouterForE2E()

	bookingID, userToken := mustCreateBookingForCancelScenario(t, r)

	cancelPath := "/bookings/" + bookingID + "/cancel"
	first := doJSON(t, r, http.MethodPost, cancelPath, userToken, nil)
	if first.Code != http.StatusOK {
		t.Fatalf("expected first cancel 200, got %d body=%s", first.Code, first.Body.String())
	}

	var firstWrap struct {
		Booking domain.Booking `json:"booking"`
	}
	mustDecode(t, first.Body.Bytes(), &firstWrap)
	if firstWrap.Booking.Status != "cancelled" {
		t.Fatalf("expected cancelled after first cancel, got %q", firstWrap.Booking.Status)
	}

	second := doJSON(t, r, http.MethodPost, cancelPath, userToken, nil)
	if second.Code != http.StatusOK {
		t.Fatalf("expected second cancel 200, got %d body=%s", second.Code, second.Body.String())
	}

	var secondWrap struct {
		Booking domain.Booking `json:"booking"`
	}
	mustDecode(t, second.Body.Bytes(), &secondWrap)
	if secondWrap.Booking.Status != "cancelled" {
		t.Fatalf("expected cancelled after second cancel, got %q", secondWrap.Booking.Status)
	}
}

func mustCreateBookingForCancelScenario(t *testing.T, r *gin.Engine) (string, string) {
	t.Helper()

	adminToken := mustDummyToken(t, r, utils.RoleAdmin)
	userToken := mustDummyToken(t, r, utils.RoleUser)

	date := time.Now().UTC().Add(24 * time.Hour)
	dateStr := date.Format("2006-01-02")
	weekday := isoWeekday(date)

	roomResp := doJSON(t, r, http.MethodPost, "/rooms/create", adminToken, map[string]interface{}{
		"name":        "Room B",
		"description": "Room for cancel",
		"capacity":    4,
	})
	if roomResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 create room, got %d body=%s", roomResp.Code, roomResp.Body.String())
	}

	var roomWrap struct {
		Room domain.Room `json:"room"`
	}
	mustDecode(t, roomResp.Body.Bytes(), &roomWrap)

	scheduleResp := doJSON(t, r, http.MethodPost, "/rooms/"+roomWrap.Room.RoomID+"/schedule/create", adminToken, map[string]interface{}{
		"daysOfWeek": []int{weekday},
		"startTime":  "09:00",
		"endTime":    "10:00",
	})
	if scheduleResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 create schedule, got %d body=%s", scheduleResp.Code, scheduleResp.Body.String())
	}

	slotsReq := httptest.NewRequest(http.MethodGet, "/rooms/"+roomWrap.Room.RoomID+"/slots/list?date="+dateStr, nil)
	slotsReq.Header.Set("Authorization", "Bearer "+adminToken)
	slotsResp := httptest.NewRecorder()
	r.ServeHTTP(slotsResp, slotsReq)
	if slotsResp.Code != http.StatusOK {
		t.Fatalf("expected 200 list slots, got %d body=%s", slotsResp.Code, slotsResp.Body.String())
	}

	var slotsWrap struct {
		Slots []domain.Slot `json:"slots"`
	}
	mustDecode(t, slotsResp.Body.Bytes(), &slotsWrap)
	if len(slotsWrap.Slots) == 0 {
		t.Fatal("expected generated slots for cancel scenario")
	}

	bookingResp := doJSON(t, r, http.MethodPost, "/bookings/create", userToken, map[string]interface{}{
		"slotId": slotsWrap.Slots[0].SlotID,
	})
	if bookingResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 create booking, got %d body=%s", bookingResp.Code, bookingResp.Body.String())
	}

	var bookingWrap struct {
		Booking domain.Booking `json:"booking"`
	}
	mustDecode(t, bookingResp.Body.Bytes(), &bookingWrap)

	return bookingWrap.Booking.BookingID, userToken
}

func mustDummyToken(t *testing.T, r *gin.Engine, role string) string {
	t.Helper()

	resp := doJSON(t, r, http.MethodPost, "/dummyLogin", "", map[string]interface{}{"role": role})
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 dummy login, got %d body=%s", resp.Code, resp.Body.String())
	}

	var tokenWrap struct {
		Token string `json:"token"`
	}
	mustDecode(t, resp.Body.Bytes(), &tokenWrap)
	if strings.TrimSpace(tokenWrap.Token) == "" {
		t.Fatal("expected non-empty token")
	}

	return tokenWrap.Token
}

func doJSON(t *testing.T, r *gin.Engine, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		payload = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, payload)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func mustDecode(t *testing.T, data []byte, out interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, string(data))
	}
}
