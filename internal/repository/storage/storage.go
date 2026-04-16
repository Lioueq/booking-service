package storage

import (
	"booking/internal/repository/transaction"
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	conn      *pgxpool.Pool
	txManager transaction.Manager
}

func NewStorage(ctx context.Context) (*Storage, error) {
	const op = "repository.New"

	connectStr := os.Getenv("POSTGRES_CONN_URL")
	conn, err := pgxpool.New(ctx, connectStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	txManager := transaction.NewPGXManager(conn)

	return &Storage{
		conn:      conn,
		txManager: txManager,
	}, nil
}

func (s *Storage) Close() {
	s.conn.Close()
}
