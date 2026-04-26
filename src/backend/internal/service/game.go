package service

import (
	"context"
	"fmt"
	monopoly_deal "the-deal/internal/engine/monopoly-deal"
	"the-deal/internal/errors"
	"the-deal/internal/event"
	"the-deal/internal/schema"
	"the-deal/internal/schema/monopoly_deal_schema"
	"the-deal/internal/schema/room_schema"
	"the-deal/internal/store"
	"the-deal/internal/token"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
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
			msg := new(schema.ServerMessage)
			switch e.Kind {
			case event.KindServerMessage:
				err = proto.Unmarshal(e.Message, msg)
				if err != nil {
					return err
				}

			case event.KindMonopolyDealEvent:
				err = proto.Unmarshal(e.Message, msg)
				if err != nil {
					return err
				}

				msg = c.MonopolyDealController.MaskEvents(tp, msg.GetMonopolyDealMessage())
			default:
			}

			if msg != nil {
				callback(msg)
			}
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

	gameID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	g, err := c.store.CreateGame(ctx, store.CreateGameParams{
		GameID:      gameID,
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

func (c *Controller) ListGameHistory(ctx context.Context, tp token.Payload, callback func(message *schema.ServerMessage)) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	limit := 25
	offset := 0

	for {
		ghs, err := c.store.ListGameHistory(ctx, store.ListGameHistoryParams{
			GameID: g.GameID,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return err
		}

		for _, gh := range ghs {
			var action monopoly_deal.Action
			switch monopoly_deal.ActionKind(gh.ActionKind) {
			case monopoly_deal.ActionKindPlayMoney:
				action = new(monopoly_deal.ActionPlayMoney)
			case monopoly_deal.ActionKindPlayProperty:
				action = new(monopoly_deal.ActionPlayProperty)
			case monopoly_deal.ActionKindPlayHouse:
				action = new(monopoly_deal.ActionPlayHouse)
			case monopoly_deal.ActionKindPlayHotel:
				action = new(monopoly_deal.ActionPlayHotel)
			case monopoly_deal.ActionKindPlayPassGo:
				action = new(monopoly_deal.ActionPlayPassGo)
			default:
				return fmt.Errorf("unsupported action kind: %s", gh.ActionKind)
			}
			err = msgpack.Unmarshal(gh.Action, action)
			if err != nil {
				return err
			}

			callback(&schema.ServerMessage{
				Payload: &schema.ServerMessage_MonopolyDealMessage{
					MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
						Payload: &monopoly_deal_schema.ServerMessage_ActionHistory{
							ActionHistory: &monopoly_deal_schema.ActionHistory{
								Action: action.Proto(),
							},
						},
					},
				},
			})
		}

		if len(ghs) < limit {
			break
		}

		offset += limit
	}

	return nil
}
