package booking

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"time"
)

type BookingRepo interface {
	GetSlotByID(ctx context.Context, slotID string) (*domain.Slot, error)
	CreateBooking(ctx context.Context, slotID, userID string, conferenceLink string) (*domain.Booking, error)
	GetBookingByID(ctx context.Context, bookingID string) (*domain.Booking, error)
	CancelBooking(ctx context.Context, bookingID string) (*domain.Booking, error)
	ListBookings(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
	MyBookings(ctx context.Context, userID string) ([]domain.Booking, error)
}

type Usecase struct {
	repo BookingRepo
}

func New(repo BookingRepo) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) CreateBooking(ctx context.Context, userID, slotID string, createConferenceLink bool) (*domain.Booking, error) {
	if userID == "" || slotID == "" {
		return nil, customErrors.ErrInvalidData
	}

	slot, err := u.repo.GetSlotByID(ctx, slotID)
	if err != nil {
		return nil, err
	}

	if slot.Start.Before(time.Now().UTC()) {
		return nil, customErrors.ErrInvalidData
	}

	var conferenceLink string
	if createConferenceLink {
		link := "www.zoom.com"
		conferenceLink = link
	}

	booking, err := u.repo.CreateBooking(ctx, slotID, userID, conferenceLink)
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (u *Usecase) CancelBooking(ctx context.Context, userID, bookingID string) (*domain.Booking, error) {
	if userID == "" || bookingID == "" {
		return nil, customErrors.ErrInvalidData
	}

	booking, err := u.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, customErrors.ErrForbidden
	}

	if booking.Status == "cancelled" {
		return booking, nil
	}

	return u.repo.CancelBooking(ctx, bookingID)
}

func (u *Usecase) ListBookings(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return u.repo.ListBookings(ctx, page, pageSize)
}

func (u *Usecase) MyBookings(ctx context.Context, userID string) ([]domain.Booking, error) {
	if userID == "" {
		return nil, customErrors.ErrInvalidRole
	}

	return u.repo.MyBookings(ctx, userID)
}
