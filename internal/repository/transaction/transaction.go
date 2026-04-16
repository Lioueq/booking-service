package transaction

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Manager interface {
	WithTx(ctx context.Context, fn func(context.Context, pgx.Tx) error) error
}

type pgxManager struct {
	pool *pgxpool.Pool
}

func NewPGXManager(pool *pgxpool.Pool) Manager {
	return &pgxManager{pool: pool}
}

func (m *pgxManager) WithTx(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
