package storage

import (
	"booking/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) ScheduleExistsByRoom(ctx context.Context, roomID string) (bool, error) {
	const op = "repository.storage.ScheduleExistsByRoom"

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM schedules WHERE room_id = $1)`
	if err := s.conn.QueryRow(ctx, query, roomID).Scan(&exists); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return exists, nil
}

func (s *Storage) getScheduleByRoom(ctx context.Context, roomID string) (*scheduleForSlots, error) {
	const op = "repository.storage.getScheduleByRoom"

	query := `SELECT days_of_week, start_time::text, end_time::text
	          FROM schedules
			  WHERE room_id = $1`

	var schedule scheduleForSlots
	err := s.conn.QueryRow(ctx, query, roomID).Scan(&schedule.DaysOfWeek, &schedule.StartTime, &schedule.EndTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &schedule, nil
}

func (s *Storage) CreateSchedule(
	ctx context.Context,
	roomID string,
	daysOfWeek []int,
	startTime string,
	endTime string,
) (*domain.Schedule, error) {
	const op = "repository.storage.CreateSchedule"

	days := make([]int, 0, len(daysOfWeek))
	for _, d := range daysOfWeek {
		days = append(days, int(d))
	}

	var sc domain.Schedule
	query := `INSERT INTO schedules (room_id, days_of_week, start_time, end_time)
	          VALUES ($1, $2::int[], $3::time, $4::time)
			  RETURNING schedule_id, room_id, days_of_week, start_time::text, end_time::text, created_at`
	var dbDays []int
	err := s.conn.QueryRow(ctx, query, roomID, days, startTime, endTime).
		Scan(&sc.ScheduleID, &sc.RoomID, &dbDays, &sc.StartTime, &sc.EndTime, &sc.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	sc.DaysOfWeek = make([]int, 0, len(dbDays))
	for _, d := range dbDays {
		sc.DaysOfWeek = append(sc.DaysOfWeek, int(d))
	}

	return &sc, nil
}
