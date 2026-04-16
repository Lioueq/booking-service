package slot

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"errors"
	"testing"
	"time"
)

type slotRepoMock struct {
	ensureSlotsForDateFn func(ctx context.Context, roomID string, dateUTC time.Time) error
	listSlotsByDateFn    func(ctx context.Context, roomID string, dateUTC time.Time) ([]domain.Slot, error)
	getRoomFn            func(ctx context.Context, roomID string) (*domain.Room, error)
}

func (m *slotRepoMock) EnsureSlotsForDate(ctx context.Context, roomID string, dateUTC time.Time) error {
	if m.ensureSlotsForDateFn == nil {
		return errors.New("unexpected call EnsureSlotsForDate")
	}
	return m.ensureSlotsForDateFn(ctx, roomID, dateUTC)
}

func (m *slotRepoMock) ListSlotsByDate(ctx context.Context, roomID string, dateUTC time.Time) ([]domain.Slot, error) {
	if m.listSlotsByDateFn == nil {
		return nil, errors.New("unexpected call ListSlotsByDate")
	}
	return m.listSlotsByDateFn(ctx, roomID, dateUTC)
}

func (m *slotRepoMock) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	if m.getRoomFn == nil {
		return nil, errors.New("unexpected call GetRoom")
	}
	return m.getRoomFn(ctx, roomID)
}

func TestGetAvailableSlots_RejectsInvalidInput(t *testing.T) {
	uc := New(&slotRepoMock{})

	_, err := uc.GetAvailableSlots(context.Background(), "", "2026-01-01")
	if !errors.Is(err, customErrors.ErrInvalidData) {
		t.Fatalf("expected ErrInvalidData, got %v", err)
	}
}

func TestGetAvailableSlots_RejectsInvalidDateFormat(t *testing.T) {
	uc := New(&slotRepoMock{})

	_, err := uc.GetAvailableSlots(context.Background(), "room-1", "01-01-2026")
	if !errors.Is(err, customErrors.ErrInvalidData) {
		t.Fatalf("expected ErrInvalidData, got %v", err)
	}
}

func TestGetAvailableSlots_PropagatesGetRoomError(t *testing.T) {
	repoErr := errors.New("room missing")
	repo := &slotRepoMock{
		getRoomFn: func(_ context.Context, _ string) (*domain.Room, error) {
			return nil, repoErr
		},
	}

	uc := New(repo)
	_, err := uc.GetAvailableSlots(context.Background(), "room-1", "2026-01-01")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected propagated error, got %v", err)
	}
}

func TestGetAvailableSlots_CallsEnsureAndListWithUTCDate(t *testing.T) {
	var ensureDate time.Time
	var listDate time.Time

	repo := &slotRepoMock{
		getRoomFn: func(_ context.Context, _ string) (*domain.Room, error) {
			return &domain.Room{RoomID: "room-1"}, nil
		},
		ensureSlotsForDateFn: func(_ context.Context, _ string, dateUTC time.Time) error {
			ensureDate = dateUTC
			return nil
		},
		listSlotsByDateFn: func(_ context.Context, _ string, dateUTC time.Time) ([]domain.Slot, error) {
			listDate = dateUTC
			return []domain.Slot{{SlotID: "slot-1"}}, nil
		},
	}

	uc := New(repo)
	slots, err := uc.GetAvailableSlots(context.Background(), "room-1", "2026-04-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(slots))
	}

	if !ensureDate.Equal(listDate) {
		t.Fatalf("expected same date passed to ensure/list, got ensure=%v list=%v", ensureDate, listDate)
	}
	if ensureDate.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %v", ensureDate.Location())
	}
	if ensureDate.Hour() != 0 || ensureDate.Minute() != 0 || ensureDate.Second() != 0 {
		t.Fatalf("expected start of day UTC, got %v", ensureDate)
	}
}
