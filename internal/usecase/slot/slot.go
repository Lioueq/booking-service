package slot

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"time"
)

type SlotRepo interface {
	EnsureSlotsForDate(ctx context.Context, roomID string, dateUTC time.Time) error
	ListSlotsByDate(ctx context.Context, roomID string, dateUTC time.Time) ([]domain.Slot, error)
	GetRoom(ctx context.Context, roomID string) (*domain.Room, error)
}

type Usecase struct {
	slotRepo SlotRepo
}

func New(slotRepo SlotRepo) *Usecase {
	return &Usecase{slotRepo: slotRepo}
}

func (u *Usecase) GetAvailableSlots(ctx context.Context, roomID, date string) ([]domain.Slot, error) {
	if roomID == "" || date == "" {
		return nil, customErrors.ErrInvalidData
	}

	dateUTC, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, customErrors.ErrInvalidData
	}
	dateUTC = time.Date(dateUTC.Year(), dateUTC.Month(), dateUTC.Day(), 0, 0, 0, 0, time.UTC)

	if _, err := u.slotRepo.GetRoom(ctx, roomID); err != nil {
		return nil, err
	}

	if err := u.slotRepo.EnsureSlotsForDate(ctx, roomID, dateUTC); err != nil {
		return nil, err
	}

	return u.slotRepo.ListSlotsByDate(ctx, roomID, dateUTC)
}
