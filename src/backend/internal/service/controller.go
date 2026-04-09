package service

import (
	"monopoly-deal/internal/config"
	"monopoly-deal/internal/event"
	"monopoly-deal/internal/store"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Controller struct {
	cfg           config.Config
	store         store.Store
	bus           *event.Bus
	mu            sync.Mutex
	playerRoomMap map[uuid.UUID]uuid.UUID
}

func NewController(cfg config.Config, pool *pgxpool.Pool, client *redis.Client) *Controller {
	c := &Controller{
		cfg:           cfg,
		store:         store.NewSQLStore(pool, nil),
		bus:           event.NewBus(client),
		playerRoomMap: make(map[uuid.UUID]uuid.UUID),
	}

	return c
}
