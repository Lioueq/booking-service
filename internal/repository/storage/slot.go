package storage

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type scheduleForSlots struct {
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

func (s *Storage) EnsureSlotsForDate(ctx context.Context, roomID string, dateUTC time.Time) error {
	const op = "repository.storage.EnsureSlotsForDate"

	schedule, err := s.getScheduleByRoom(ctx, roomID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if schedule == nil {
		return nil
	}

	weekday := isoWeekday(dateUTC)
	if !containsWeekday(schedule.DaysOfWeek, weekday) {
		return nil
	}

	startOfDay := time.Date(dateUTC.Year(), dateUTC.Month(), dateUTC.Day(), 0, 0, 0, 0, time.UTC)
	exists, err := s.availableSlotsForDate(ctx, roomID, startOfDay)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	startTOD, err := parseTimeOfDay(schedule.StartTime)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	endTOD, err := parseTimeOfDay(schedule.EndTime)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	startAt := time.Date(
		dateUTC.Year(),
		dateUTC.Month(),
		dateUTC.Day(),
		startTOD.Hour(),
		startTOD.Minute(),
		0,
		0,
		time.UTC,
	)
	endAt := time.Date(
		dateUTC.Year(),
		dateUTC.Month(),
		dateUTC.Day(),
		endTOD.Hour(),
		endTOD.Minute(),
		0,
		0,
		time.UTC,
	)

	for current := startAt; current.Before(endAt); current = current.Add(30 * time.Minute) {
		next := current.Add(30 * time.Minute)
		if next.After(endAt) {
			break
		}
		query := `INSERT INTO slots (room_id, start_at, end_at)
				  VALUES ($1, $2, $3)
				  ON CONFLICT (room_id, start_at, end_at) DO NOTHING`
		if _, err := s.conn.Exec(ctx, query, roomID, current, next); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Storage) ListSlotsByDate(ctx context.Context, roomID string, dateUTC time.Time) ([]domain.Slot, error) {
	const op = "repository.storage.ListSlotsByDate"

	startOfDay := time.Date(dateUTC.Year(), dateUTC.Month(), dateUTC.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `SELECT slot_id, room_id, start_at, end_at
			  FROM slots s
			  WHERE room_id = $1
			    AND start_at >= $2
				AND start_at < $3
				AND NOT EXISTS (
					SELECT 1
					FROM bookings b
					WHERE b.slot_id = s.slot_id
					  AND b.status = 'active'
				)
			  ORDER BY start_at`

	rows, err := s.conn.Query(ctx, query, roomID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	slots := make([]domain.Slot, 0)
	for rows.Next() {
		var slot domain.Slot
		if err := rows.Scan(&slot.SlotID, &slot.RoomID, &slot.Start, &slot.End); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return slots, nil
}

func (s *Storage) GetSlotByID(ctx context.Context, slotID string) (*domain.Slot, error) {
	const op = "repository.storage.GetSlotByID"

	query := `SELECT slot_id, room_id, start_at, end_at
			  FROM slots
			  WHERE slot_id = $1`

	var slot domain.Slot
	err := s.conn.QueryRow(ctx, query, slotID).Scan(&slot.SlotID, &slot.RoomID, &slot.Start, &slot.End)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customErrors.ErrSlotNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &slot, nil
}

func (s *Storage) availableSlotsForDate(ctx context.Context, roomID string, startOfDay time.Time) (bool, error) {
	const op = "repository.storage.availableSlotsForDate"

	endOfDay := startOfDay.Add(24 * time.Hour)

	var exists bool
	query := `SELECT EXISTS(
				SELECT slot_id, room_id, start_at, end_at
				FROM slots
				WHERE room_id = $1
				  AND start_at >= $2
				  AND start_at < $3
			)`

	if err := s.conn.QueryRow(ctx, query, roomID, startOfDay, endOfDay).Scan(&exists); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return exists, nil
}

func containsWeekday(days []int, weekday int) bool {
	for _, d := range days {
		if int(d) == weekday {
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

func parseTimeOfDay(value string) (time.Time, error) {
	if parsed, err := time.Parse("15:04:05", value); err == nil {
		return parsed, nil
	}
	return time.Parse("15:04", value)
}
