package schedule

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"errors"
	"testing"
)

type scheduleRepoMock struct {
	scheduleExistsByRoomFn func(ctx context.Context, roomID string) (bool, error)
	createScheduleFn       func(ctx context.Context, roomID string, daysOfWeek []int, startTime, endTime string) (*domain.Schedule, error)
	getRoomFn              func(ctx context.Context, roomID string) (*domain.Room, error)
}

func (m *scheduleRepoMock) ScheduleExistsByRoom(ctx context.Context, roomID string) (bool, error) {
	if m.scheduleExistsByRoomFn == nil {
		return false, errors.New("unexpected call ScheduleExistsByRoom")
	}
	return m.scheduleExistsByRoomFn(ctx, roomID)
}

func (m *scheduleRepoMock) CreateSchedule(ctx context.Context, roomID string, daysOfWeek []int, startTime, endTime string) (*domain.Schedule, error) {
	if m.createScheduleFn == nil {
		return nil, errors.New("unexpected call CreateSchedule")
	}
	return m.createScheduleFn(ctx, roomID, daysOfWeek, startTime, endTime)
}

func (m *scheduleRepoMock) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	if m.getRoomFn == nil {
		return nil, errors.New("unexpected call GetRoom")
	}
	return m.getRoomFn(ctx, roomID)
}

func TestCreateSchedule_RejectsInvalidDays(t *testing.T) {
	uc := New(&scheduleRepoMock{})

	_, err := uc.CreateSchedule(context.Background(), "room-1", []int{1, 1}, "09:00", "10:00")
	if !errors.Is(err, customErrors.ErrInvalidData) {
		t.Fatalf("expected ErrInvalidData, got %v", err)
	}
}

func TestCreateSchedule_RejectsInvalidTimeRange(t *testing.T) {
	uc := New(&scheduleRepoMock{})

	_, err := uc.CreateSchedule(context.Background(), "room-1", []int{1}, "10:00", "09:00")
	if !errors.Is(err, customErrors.ErrInvalidData) {
		t.Fatalf("expected ErrInvalidData, got %v", err)
	}
}

func TestCreateSchedule_RoomNotFound(t *testing.T) {
	repo := &scheduleRepoMock{
		getRoomFn: func(_ context.Context, _ string) (*domain.Room, error) {
			return nil, errors.New("db down")
		},
	}

	uc := New(repo)
	_, err := uc.CreateSchedule(context.Background(), "room-1", []int{1}, "09:00", "10:00")
	if !errors.Is(err, customErrors.ErrRoomNotFound) {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestCreateSchedule_AlreadyExists(t *testing.T) {
	repo := &scheduleRepoMock{
		getRoomFn: func(_ context.Context, _ string) (*domain.Room, error) {
			return &domain.Room{RoomID: "room-1"}, nil
		},
		scheduleExistsByRoomFn: func(_ context.Context, _ string) (bool, error) {
			return true, nil
		},
	}

	uc := New(repo)
	_, err := uc.CreateSchedule(context.Background(), "room-1", []int{1}, "09:00", "10:00")
	if !errors.Is(err, customErrors.ErrScheduleExists) {
		t.Fatalf("expected ErrScheduleExists, got %v", err)
	}
}

func TestCreateSchedule_Success(t *testing.T) {
	repo := &scheduleRepoMock{
		getRoomFn: func(_ context.Context, _ string) (*domain.Room, error) {
			return &domain.Room{RoomID: "room-1"}, nil
		},
		scheduleExistsByRoomFn: func(_ context.Context, _ string) (bool, error) {
			return false, nil
		},
		createScheduleFn: func(_ context.Context, roomID string, daysOfWeek []int, startTime, endTime string) (*domain.Schedule, error) {
			if roomID != "room-1" || len(daysOfWeek) != 2 || startTime != "09:00" || endTime != "10:00" {
				return nil, errors.New("unexpected args")
			}
			return &domain.Schedule{ScheduleID: "s1", RoomID: roomID}, nil
		},
	}

	uc := New(repo)
	schedule, err := uc.CreateSchedule(context.Background(), "room-1", []int{1, 2}, "09:00", "10:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schedule.ScheduleID != "s1" {
		t.Fatalf("expected schedule id s1, got %q", schedule.ScheduleID)
	}
}
