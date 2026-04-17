package service

import (
	"context"
	"fun-kames/internal/errors"
	"fun-kames/internal/event"
	"fun-kames/internal/schema"
	"fun-kames/internal/store"
	"fun-kames/internal/token"
)

func (c *Controller) CreateGame(ctx context.Context, tp token.Payload) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return errors.Internal(err)
	}

	if !rp.IsHost {
		return errors.PlayerIsNotHost
	}

	ps, err := c.store.GetPlayersByRoom(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return errors.Internal(err)
	}

	allReady := true
	for _, p := range ps {
		allReady = allReady && p.IsReady
	}
	if !allReady {
		return errors.AllPlayersNotReady
	}

	r, err := c.store.GetRoomByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return errors.Internal(err)
	}

	_, err = c.store.CreateGame(ctx, store.CreateGameParams{
		DisplayName: r.DisplayName,
		Game:        r.Game,
		Settings:    r.Settings,
		GameState:   nil, // TODO - instantiate new game instance
	})
	if err != nil {
		return errors.Internal(err)
	}

	return c.bus.Publish(ctx, event.RoomChannelPre+rp.RoomID.String(), &schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: &schema.ServerRoomMessage{
				Payload: &schema.ServerRoomMessage_GameStarted{
					GameStarted: &schema.GameStarted{},
				},
			},
		},
	})
}
