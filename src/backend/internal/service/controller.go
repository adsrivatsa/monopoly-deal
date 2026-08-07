package service

import (
	"the-deal/internal/config"
	"the-deal/internal/event"
	deal_no_mercy "the-deal/internal/service/deal-no-mercy"
	monopoly_deal "the-deal/internal/service/monopoly-deal"
	"the-deal/internal/store"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Controller struct {
	cfg                    config.Config
	store                  store.Store
	bus                    *event.Bus
	MonopolyDealController *monopoly_deal.Controller
	DealNoMercyController  *deal_no_mercy.Controller
}

func NewController(cfg config.Config, pool *pgxpool.Pool, client *redis.Client) *Controller {
	c := &Controller{
		cfg:                    cfg,
		store:                  store.NewSQLStore(pool, nil),
		bus:                    event.NewBus(client),
		MonopolyDealController: monopoly_deal.NewController(cfg, pool, client),
		DealNoMercyController:  deal_no_mercy.NewController(cfg, pool, client),
	}

	return c
}
