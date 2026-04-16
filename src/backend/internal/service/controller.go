package service

import (
	"fun-kames/internal/config"
	"fun-kames/internal/event"
	"fun-kames/internal/store"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Controller struct {
	cfg   config.Config
	store store.Store
	bus   *event.Bus
}

func NewController(cfg config.Config, pool *pgxpool.Pool, client *redis.Client) *Controller {
	c := &Controller{
		cfg:   cfg,
		store: store.NewSQLStore(pool, nil),
		bus:   event.NewBus(client),
	}

	return c
}
