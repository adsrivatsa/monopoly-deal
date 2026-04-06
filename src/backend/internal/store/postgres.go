package store

import (
	"context"
	"errors"
	"fmt"
	"monopoly-deal/internal/config"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(cfg config.Config, waitTime time.Duration) *pgxpool.Pool {
	MigrateUp(cfg, waitTime)

	count := 0

	for {
		pool, err := newPostgresPool(cfg)
		if err == nil {
			return pool
		}

		count++
		if count >= 5 {
			fmt.Println("unable to connect: ", err)
			fmt.Printf("retrying in %d ms...", waitTime.Milliseconds())
			time.Sleep(waitTime)
			count = 0
		}
	}
}

func newPostgresPool(opts config.Config) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(opts.DatabaseURL)
	if err != nil {
		return nil, err
	}

	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		t, err := conn.LoadType(ctx, "notification_link_type")
		if err != nil {
			return err
		}
		conn.TypeMap().RegisterType(t)

		t, err = conn.LoadType(ctx, "_notification_link_type")
		if err != nil {
			return err
		}
		conn.TypeMap().RegisterType(t)

		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func MigrateUp(cfg config.Config, waitTime time.Duration) {
	count := 0
	for {
		err := migrateUp(cfg)
		if err == nil || errors.Is(err, migrate.ErrNoChange) {
			break
		}

		count++
		if count >= 5 {
			fmt.Println("unable to connect: ", err)
			fmt.Printf("retrying in %d ms...", waitTime.Milliseconds())
			time.Sleep(waitTime)
			count = 0
		}
	}
}

func migrateUp(cfg config.Config) error {
	migration, err := migrate.New(cfg.MigrationURL, cfg.DatabaseURL)
	if err != nil {
		return err
	}

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
