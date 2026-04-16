package storage

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateRoom(ctx context.Context, name, description string, capacity int) (*domain.Room, error) {
	const op = "repository.storage.CreateRoom"

	var room domain.Room
	query := `INSERT INTO rooms (room_name, room_description, capacity) 
			  VALUES ($1, $2, $3) 
			  RETURNING room_id, room_name, room_description, capacity, created_at`
	err := s.conn.QueryRow(ctx, query, name, description, capacity).
		Scan(&room.RoomID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &room, nil
}

func (s *Storage) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	const op = "repository.storage.GetRoom"

	var room domain.Room
	query := `SELECT room_id, room_name, room_description, capacity, created_at FROM rooms WHERE room_id = $1`
	if err := s.conn.QueryRow(ctx, query, roomID).Scan(
		&room.RoomID,
		&room.Name,
		&room.Description,
		&room.Capacity,
		&room.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customErrors.ErrRoomNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &room, nil
}

func (s *Storage) GetRooms(ctx context.Context) ([]domain.Room, error) {
	const op = "repository.storage.GetRooms"

	query := `SELECT room_id, room_name, room_description, capacity, created_at
    		  FROM rooms
			  ORDER BY created_at DESC`
	rows, err := s.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	rooms := make([]domain.Room, 0)
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(
			&room.RoomID,
			&room.Name,
			&room.Description,
			&room.Capacity,
			&room.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return rooms, nil
}
