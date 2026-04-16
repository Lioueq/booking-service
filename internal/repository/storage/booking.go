package storage

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateBooking(ctx context.Context, slotID, userID string, conferenceLink string) (*domain.Booking, error) {
	const op = "repository.storage.CreateBooking"

	query := `INSERT INTO bookings (slot_id, user_id, status, conference_link)
			  VALUES ($1, $2, 'active', $3)
			  RETURNING booking_id, slot_id, user_id, status, COALESCE(conference_link, ''), created_at`

	var booking domain.Booking
	err := s.conn.QueryRow(ctx, query, slotID, userID, conferenceLink).Scan(
		&booking.BookingID,
		&booking.SlotID,
		&booking.UserID,
		&booking.Status,
		&booking.ConferenceLink,
		&booking.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, customErrors.ErrSlotAlreadyBooked
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &booking, nil
}

func (s *Storage) GetBookingByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	const op = "repository.storage.GetBookingByID"

	query := `SELECT booking_id, slot_id, user_id, status, COALESCE(conference_link, ''), created_at
	          FROM bookings
			  WHERE booking_id = $1`

	var booking domain.Booking
	err := s.conn.QueryRow(ctx, query, bookingID).Scan(
		&booking.BookingID,
		&booking.SlotID,
		&booking.UserID,
		&booking.Status,
		&booking.ConferenceLink,
		&booking.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customErrors.ErrBookingNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &booking, nil
}

func (s *Storage) CancelBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	const op = "repository.storage.CancelBooking"

	query := `UPDATE bookings
			  SET status = 'cancelled'
			  WHERE booking_id = $1
			  RETURNING booking_id, slot_id, user_id, status, COALESCE(conference_link, ''), created_at`

	var booking domain.Booking
	err := s.conn.QueryRow(ctx, query, bookingID).Scan(
		&booking.BookingID,
		&booking.SlotID,
		&booking.UserID,
		&booking.Status,
		&booking.ConferenceLink,
		&booking.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customErrors.ErrBookingNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &booking, nil
}

func (s *Storage) ListBookings(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	const op = "repository.storage.ListBookings"

	var total int
	if err := s.conn.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT booking_id, slot_id, user_id, status, COALESCE(conference_link, ''), created_at
			  FROM bookings
			  ORDER BY created_at DESC
			  LIMIT $1 OFFSET $2`

	rows, err := s.conn.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		var booking domain.Booking
		if err := rows.Scan(
			&booking.BookingID,
			&booking.SlotID,
			&booking.UserID,
			&booking.Status,
			&booking.ConferenceLink,
			&booking.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("%s: %w", op, err)
		}
		bookings = append(bookings, booking)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return bookings, total, nil
}

func (s *Storage) MyBookings(ctx context.Context, userID string) ([]domain.Booking, error) {
	const op = "repository.storage.MyBookings"

	query := `SELECT b.booking_id, b.slot_id, b.user_id, b.status, COALESCE(b.conference_link, ''), b.created_at
			  FROM bookings b
			  JOIN slots s ON s.slot_id = b.slot_id
			  WHERE b.user_id = $1
			    AND s.start_at >= NOW()
			  ORDER BY s.start_at`

	rows, err := s.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		var booking domain.Booking
		if err := rows.Scan(
			&booking.BookingID,
			&booking.SlotID,
			&booking.UserID,
			&booking.Status,
			&booking.ConferenceLink,
			&booking.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		bookings = append(bookings, booking)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return bookings, nil
}
