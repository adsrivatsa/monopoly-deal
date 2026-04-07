package service

import (
	"monopoly-deal/internal/config"
	"monopoly-deal/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Controller struct {
	cfg   config.Config
	store store.Store
}

func NewController(cfg config.Config, pool *pgxpool.Pool) *Controller {
	c := &Controller{
		cfg:   cfg,
		store: store.NewSQLStore(pool, nil),
	}

	return c
}
