package service

import (
	"context"
	monopoly_deal "fun-kames/internal/engine/monopoly-deal"
	"fun-kames/internal/errors"
	"fun-kames/internal/event"
	"fun-kames/internal/schema"
	"fun-kames/internal/schema/room_schema"
	"fun-kames/internal/store"
	"fun-kames/internal/token"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (c *Controller) ListenGameEvents(ctx context.Context, tp token.Payload, callback func(message *schema.ServerMessage)) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame, err)
		}
		return err
	}

	ch, err := c.bus.Subscribe(ctx, event.GameChannelPre+g.GameID.String())
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-ch:
			var msg schema.ServerMessage
			switch e.Kind {
			case event.KindServerMessage:
				err = proto.Unmarshal(e.Message, &msg)
				if err != nil {
					return err
				}

			case event.KindMonopolyDealEvent:
				err = proto.Unmarshal(e.Message, &msg)
				if err != nil {
					return err
				}

				mdMsg := c.maskMonopolyDealPrivateEvents(tp, msg.GetMonopolyDealMessage())
				msg.Payload = &schema.ServerMessage_MonopolyDealMessage{
					MonopolyDealMessage: mdMsg,
				}
			default:
			}

			callback(&msg)
		}
	}
}

func (c *Controller) GetGame(ctx context.Context, tp token.Payload) (Game, error) {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return Game{}, errors.EntityNotFound(errors.EntityGame, err)
		}
		return Game{}, err
	}

	return g, nil
}

func (c *Controller) CreateGame(ctx context.Context, tp token.Payload) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return err
	}

	if !rp.IsHost {
		return errors.PlayerIsNotHost
	}

	ps, err := c.store.GetPlayersByRoom(ctx, rp.RoomID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return err
	}

	allReady := true
	playerIDs := make([]uuid.UUID, 0, len(ps))
	for _, p := range ps {
		allReady = allReady && (p.IsReady || p.IsHost)
		playerIDs = append(playerIDs, p.PlayerID)
	}
	if !allReady {
		return errors.AllPlayersNotReady
	}

	r, err := c.store.GetRoomByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return err
	}

	var buf []byte
	switch r.Game {
	case store.GameTypeMonopolyDeal:
		var settings monopoly_deal.Settings
		err = settings.Decode(r.Settings)
		if err != nil {
			return err
		}

		game := monopoly_deal.NewGame(settings, playerIDs)
		buf, err = game.EncodeMsgpack()
		if err != nil {
			return err
		}
	default:
		return errors.GameNotSupported
	}

	g, err := c.store.CreateGame(ctx, store.CreateGameParams{
		DisplayName: r.DisplayName,
		Game:        r.Game,
		GameState:   buf,
	})
	if err != nil {
		return err
	}

	err = c.store.CreateGamePlayersFromRoom(ctx, store.CreateGamePlayersFromRoomParams{
		GameID: g.GameID,
		RoomID: r.RoomID,
	})
	if err != nil {
		return err
	}

	err = c.store.DeleteRoom(ctx, r.RoomID)
	if err != nil {
		return err
	}

	err = c.store.DeleteRoomPlayersByRoom(ctx, r.RoomID)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.gameLocks[g.GameID] = &sync.RWMutex{}
	c.mu.Unlock()

	e := &schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: &room_schema.ServerMessage{
				Payload: &room_schema.ServerMessage_GameStarted{
					GameStarted: &room_schema.GameStarted{
						GameId: g.GameID.String(),
					},
				},
			},
		},
	}

	buf, err = proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.RoomChannelPre+rp.RoomID.String(), event.NewServerMessageEvent(buf))
}
