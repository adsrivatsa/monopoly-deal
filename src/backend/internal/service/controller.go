package service

import (
	"fun-kames/internal/config"
	"fun-kames/internal/event"
	monopoly_deal "fun-kames/internal/service/monopoly-deal"
	"fun-kames/internal/store"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Controller struct {
	cfg                    config.Config
	store                  store.Store
	bus                    *event.Bus
	MonopolyDealController *monopoly_deal.Controller
}

func NewController(cfg config.Config, pool *pgxpool.Pool, client *redis.Client) *Controller {
	c := &Controller{
		cfg:                    cfg,
		store:                  store.NewSQLStore(pool, nil),
		bus:                    event.NewBus(client),
		MonopolyDealController: monopoly_deal.NewController(cfg, pool, client),
	}

	return c
}
