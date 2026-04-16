package schedule

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"time"
)

type ScheduleRepo interface {
	ScheduleExistsByRoom(ctx context.Context, roomID string) (bool, error)
	CreateSchedule(ctx context.Context, roomID string, daysOfWeek []int, startTime, endTime string) (*domain.Schedule, error)
	GetRoom(ctx context.Context, roomID string) (*domain.Room, error)
}

type Usecase struct {
	scheduleRepo ScheduleRepo
}

func New(scheduleRepo ScheduleRepo) *Usecase {
	return &Usecase{
		scheduleRepo: scheduleRepo,
	}
}

func (u *Usecase) CreateSchedule(
	ctx context.Context,
	roomID string,
	daysOfWeek []int,
	startTime string,
	endTime string,
) (*domain.Schedule, error) {
	if roomID == "" || len(daysOfWeek) == 0 {
		return nil, customErrors.ErrInvalidData
	}

	seen := make(map[int]struct{}, len(daysOfWeek))
	for _, d := range daysOfWeek {
		if d < 1 || d > 7 {
			return nil, customErrors.ErrInvalidData
		}
		if _, ok := seen[d]; ok {
			return nil, customErrors.ErrInvalidData
		}
		seen[d] = struct{}{}
	}

	st, err := time.Parse("15:04", startTime)
	if err != nil {
		return nil, customErrors.ErrInvalidData
	}
	et, err := time.Parse("15:04", endTime)
	if err != nil {
		return nil, customErrors.ErrInvalidData
	}
	if !et.After(st) {
		return nil, customErrors.ErrInvalidData
	}

	if _, err := u.scheduleRepo.GetRoom(ctx, roomID); err != nil {
		return nil, customErrors.ErrRoomNotFound
	}

	exists, err := u.scheduleRepo.ScheduleExistsByRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, customErrors.ErrScheduleExists
	}

	return u.scheduleRepo.CreateSchedule(ctx, roomID, daysOfWeek, startTime, endTime)
}
