package deal_no_mercy

import (
	"context"
	"the-deal/internal/config"
	"the-deal/internal/event"
	"the-deal/internal/store"
	"the-deal/internal/timex"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Controller struct {
	cfg        config.Config
	store      store.Store
	bus        *event.Bus
	timeoutMan *timex.TimeoutManager[TimeoutKey]
}

func NewController(cfg config.Config, pool *pgxpool.Pool, client *redis.Client) *Controller {
	c := &Controller{
		cfg:        cfg,
		store:      store.NewSQLStore(pool, nil),
		bus:        event.NewBus(client),
		timeoutMan: timex.NewTimeoutManager[TimeoutKey](),
	}

	err := c.rehydrateTimeouts()
	if err != nil {
		panic(err)
	}

	return c
}

func (c *Controller) rehydrateTimeouts() error {
	ctx := context.Background()

	limit := 25

	for {
		n := 0

		err := c.store.ExecTx(ctx, func(q *store.Queries) error {
			gts, err := q.ClaimGameTimeoutsByGame(ctx, store.ClaimGameTimeoutsByGameParams{
				Limit: int32(limit),
				Game:  store.GameTypeDealNoMercy,
			})
			if err != nil {
				return err
			}

			n = len(gts)

			for _, gt := range gts {
				delay := time.Duration(gt.DurationMs) * time.Millisecond
				deadline := time.Now().Add(delay)

				if gt.DemandID == nil {
					err = c.scheduleDefaultMove(ctx, q, gt.GameID, gt.PlayerID, delay, deadline)
				} else {
					err = c.scheduleDefaultDemand(ctx, q, gt.GameID, gt.PlayerID, *gt.DemandID, delay, deadline)
				}
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return err
		}

		if n < limit {
			break
		}
	}

	return nil
}
