package storage

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateUser(ctx context.Context, email, passwordHash, role string) (*domain.User, error) {
	const op = "repository.storage.CreateUser"

	var user domain.User
	query := `INSERT INTO users (user_email, password_hash, user_role) 
			  VALUES ($1, $2, $3) 
			  RETURNING user_id, user_email, password_hash, user_role, created_at`
	err := s.conn.QueryRow(ctx, query, email, passwordHash, role).
		Scan(&user.UserID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const op = "repository.storage.GetUserByEmail"

	var user domain.User
	query := `SELECT user_id, user_email, password_hash, user_role, created_at 
		 	  FROM users 
			  WHERE user_email = $1`
	err := s.conn.QueryRow(ctx, query, email).
		Scan(&user.UserID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	const op = "repository.storage.GetUser"

	var user domain.User
	query := `SELECT user_id, user_email, user_role 
		  	  FROM users 
			  WHERE user_id = $1`
	err := s.conn.QueryRow(ctx, query, userID).
		Scan(&user.UserID, &user.Email, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
