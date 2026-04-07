package store

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	ExecTx(ctx context.Context, fn func(queries *Queries) error) error
}

type SQLStore struct {
	logger *slog.Logger
	*Queries
	*pgxpool.Pool
}

func NewSQLStore(pool *pgxpool.Pool, logger *slog.Logger) Store {
	return &SQLStore{
		Pool:    pool,
		Queries: New(pool),
		logger:  logger,
	}
}

func (store *SQLStore) ExecTx(ctx context.Context, fn func(queries *Queries) error) error {
	tx, err := store.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}
