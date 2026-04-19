package service

import (
	"fun-kames/internal/config"
	"fun-kames/internal/event"
	"fun-kames/internal/store"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Controller struct {
	cfg   config.Config
	store store.Store
	bus   *event.Bus

	// for sync games
	mu        sync.RWMutex
	gameLocks map[uuid.UUID]*sync.RWMutex
}

func NewController(cfg config.Config, pool *pgxpool.Pool, client *redis.Client) *Controller {
	c := &Controller{
		cfg:       cfg,
		store:     store.NewSQLStore(pool, nil),
		bus:       event.NewBus(client),
		gameLocks: make(map[uuid.UUID]*sync.RWMutex),
	}

	return c
}
