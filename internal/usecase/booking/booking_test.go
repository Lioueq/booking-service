package booking

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"errors"
	"testing"
	"time"
)

type bookingRepoMock struct {
	getSlotByIDFn   func(ctx context.Context, slotID string) (*domain.Slot, error)
	createBookingFn func(ctx context.Context, slotID, userID string, conferenceLink string) (*domain.Booking, error)
	getBookingByID  func(ctx context.Context, bookingID string) (*domain.Booking, error)
	cancelBookingFn func(ctx context.Context, bookingID string) (*domain.Booking, error)
	listBookingsFn  func(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
	myBookingsFn    func(ctx context.Context, userID string) ([]domain.Booking, error)
}

func (m *bookingRepoMock) GetSlotByID(ctx context.Context, slotID string) (*domain.Slot, error) {
	if m.getSlotByIDFn == nil {
		return nil, errors.New("unexpected call GetSlotByID")
	}
	return m.getSlotByIDFn(ctx, slotID)
}

func (m *bookingRepoMock) CreateBooking(ctx context.Context, slotID, userID string, conferenceLink string) (*domain.Booking, error) {
	if m.createBookingFn == nil {
		return nil, errors.New("unexpected call CreateBooking")
	}
	return m.createBookingFn(ctx, slotID, userID, conferenceLink)
}

func (m *bookingRepoMock) GetBookingByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	if m.getBookingByID == nil {
		return nil, errors.New("unexpected call GetBookingByID")
	}
	return m.getBookingByID(ctx, bookingID)
}

func (m *bookingRepoMock) CancelBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	if m.cancelBookingFn == nil {
		return nil, errors.New("unexpected call CancelBooking")
	}
	return m.cancelBookingFn(ctx, bookingID)
}

func (m *bookingRepoMock) ListBookings(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	if m.listBookingsFn == nil {
		return nil, 0, errors.New("unexpected call ListBookings")
	}
	return m.listBookingsFn(ctx, page, pageSize)
}

func (m *bookingRepoMock) MyBookings(ctx context.Context, userID string) ([]domain.Booking, error) {
	if m.myBookingsFn == nil {
		return nil, errors.New("unexpected call MyBookings")
	}
	return m.myBookingsFn(ctx, userID)
}

func TestCreateBooking_RejectsInvalidInput(t *testing.T) {
	uc := New(&bookingRepoMock{})

	_, err := uc.CreateBooking(context.Background(), "", "slot-id", false)
	if !errors.Is(err, customErrors.ErrInvalidData) {
		t.Fatalf("expected ErrInvalidData, got %v", err)
	}
}

func TestCreateBooking_RejectsPastSlot(t *testing.T) {
	repo := &bookingRepoMock{
		getSlotByIDFn: func(_ context.Context, _ string) (*domain.Slot, error) {
			return &domain.Slot{Start: time.Now().UTC().Add(-time.Minute)}, nil
		},
	}

	uc := New(repo)
	_, err := uc.CreateBooking(context.Background(), "user-1", "slot-1", false)
	if !errors.Is(err, customErrors.ErrInvalidData) {
		t.Fatalf("expected ErrInvalidData, got %v", err)
	}
}

func TestCreateBooking_PassesConferenceLinkFlag(t *testing.T) {
	var gotConferenceLink string

	repo := &bookingRepoMock{
		getSlotByIDFn: func(_ context.Context, _ string) (*domain.Slot, error) {
			return &domain.Slot{Start: time.Now().UTC().Add(time.Hour)}, nil
		},
		createBookingFn: func(_ context.Context, _ string, _ string, conferenceLink string) (*domain.Booking, error) {
			gotConferenceLink = conferenceLink
			return &domain.Booking{BookingID: "b1"}, nil
		},
	}

	uc := New(repo)
	_, err := uc.CreateBooking(context.Background(), "user-1", "slot-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotConferenceLink != "www.zoom.com" {
		t.Fatalf("expected conference link to be set, got %q", gotConferenceLink)
	}
}

func TestCancelBooking_ForbiddenForForeignBooking(t *testing.T) {
	repo := &bookingRepoMock{
		getBookingByID: func(_ context.Context, _ string) (*domain.Booking, error) {
			return &domain.Booking{BookingID: "b1", UserID: "other-user", Status: "active"}, nil
		},
	}

	uc := New(repo)
	_, err := uc.CancelBooking(context.Background(), "user-1", "b1")
	if !errors.Is(err, customErrors.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCancelBooking_IdempotentForCancelled(t *testing.T) {
	cancelCalled := false
	repo := &bookingRepoMock{
		getBookingByID: func(_ context.Context, _ string) (*domain.Booking, error) {
			return &domain.Booking{BookingID: "b1", UserID: "user-1", Status: "cancelled"}, nil
		},
		cancelBookingFn: func(_ context.Context, _ string) (*domain.Booking, error) {
			cancelCalled = true
			return nil, nil
		},
	}

	uc := New(repo)
	b, err := uc.CancelBooking(context.Background(), "user-1", "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Status != "cancelled" {
		t.Fatalf("expected cancelled status, got %q", b.Status)
	}
	if cancelCalled {
		t.Fatal("cancel repository method must not be called for already cancelled booking")
	}
}

func TestListBookings_NormalizesPagination(t *testing.T) {
	var gotPage, gotPageSize int
	repo := &bookingRepoMock{
		listBookingsFn: func(_ context.Context, page, pageSize int) ([]domain.Booking, int, error) {
			gotPage = page
			gotPageSize = pageSize
			return nil, 0, nil
		},
	}

	uc := New(repo)
	_, _, err := uc.ListBookings(context.Background(), 0, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPage != 1 || gotPageSize != 100 {
		t.Fatalf("expected normalized pagination page=1 pageSize=100, got page=%d pageSize=%d", gotPage, gotPageSize)
	}
}

func TestMyBookings_RejectsEmptyUserID(t *testing.T) {
	uc := New(&bookingRepoMock{})

	_, err := uc.MyBookings(context.Background(), "")
	if !errors.Is(err, customErrors.ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}
