package monopoly_deal

import (
	"context"
	"fmt"
	monopoly_deal "the-deal/internal/engine/monopoly-deal"
	"the-deal/internal/errors"
	"the-deal/internal/event"
	"the-deal/internal/schema"
	"the-deal/internal/schema/monopoly_deal_schema"
	"the-deal/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
)

type TimeoutKind int

const (
	TimeoutKindMove TimeoutKind = iota
	TimeoutKindDemand
)

type TimeoutKey struct {
	Kind     TimeoutKind
	PlayerID uuid.UUID
	DemandID *string
}

func NewMoveTimeoutKey(playerID uuid.UUID) TimeoutKey {
	return TimeoutKey{
		Kind:     TimeoutKindMove,
		PlayerID: playerID,
	}
}

func NewDemandTimeoutKey(playerID uuid.UUID, demandID string) TimeoutKey {
	return TimeoutKey{
		Kind:     TimeoutKindDemand,
		PlayerID: playerID,
		DemandID: &demandID,
	}
}

func (c *Controller) defaultMove(gameID, playerID, tokenID uuid.UUID) func() {
	return func() {
		ctx := context.Background()

		var actions []monopoly_deal.Action
		var err error
		var deadline time.Time

		err = c.store.ExecTx(ctx, func(q *store.Queries) error {
			gt, err := q.GetGameTimeoutForUpdate(ctx, store.GetGameTimeoutForUpdateParams{
				GameID:   gameID,
				PlayerID: playerID,
				DemandID: nil,
			})
			if err != nil {
				if errors.DBErrorCode(err) == errors.NoDataFound {
					return nil
				}
				return err
			}

			if gt.TokenID != tokenID {
				return nil
			}

			g, err := q.GetGameByPlayerForUpdate(ctx, gt.PlayerID)
			if err != nil {
				return err
			}

			game, err := monopoly_deal.DecodeMsgpack(g.GameState)
			if err != nil {
				return err
			}

			defaultActs, err := game.DefaultMove(playerID)
			if err != nil {
				return err
			}

			completeAct, err := game.CompleteTurn(playerID)
			if err != nil {
				return err
			}

			actions = append(defaultActs, completeAct)

			buf, err := game.EncodeMsgpack()
			if err != nil {
				return err
			}

			g, err = q.UpdateGameState(ctx, store.UpdateGameStateParams{
				GameState: buf,
				GameID:    gameID,
			})
			if err != nil {
				return err
			}

			for _, action := range actions {
				buf, err = msgpack.Marshal(action)
				if err != nil {
					return err
				}

				_, err = q.CreateGameHistory(ctx, store.CreateGameHistoryParams{
					GameID:        gameID,
					SeqNum:        int16(action.GetSeqNum()),
					ActionKind:    string(action.GetKind()),
					ActionVersion: int32(action.GetVersion()),
					Action:        buf,
				})
				if err != nil {
					return err
				}
			}

			_, err = q.DeleteGameTimeout(ctx, store.DeleteGameTimeoutParams{
				GameID:   gameID,
				PlayerID: playerID,
				DemandID: nil,
			})
			if err != nil {
				return err
			}

			currPlayerID := game.Players[game.CurrPlayerIdx]
			deadline = time.Now().Add(game.Config.MoveTimeout)
			err = c.scheduleDefaultMove(ctx, q, gameID, currPlayerID, deadline)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, action := range actions {
			actionProto := action.Proto()
			actionProto.TurnDeadlineMs = deadline.UnixMilli()

			e := &schema.ServerMessage{
				Payload: &schema.ServerMessage_MonopolyDealMessage{
					MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
						Payload: &monopoly_deal_schema.ServerMessage_Action{
							Action: actionProto,
						},
					},
				},
			}

			buf, err := proto.Marshal(e)
			if err != nil {
				fmt.Println(err)
				return
			}

			err = c.bus.Publish(ctx, event.GameChannelPre+gameID.String(), event.NewMonopolyDealEvent(buf))
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func (c *Controller) scheduleDefaultMove(ctx context.Context, q *store.Queries, gameID, playerID uuid.UUID, deadline time.Time) error {
	gt, err := q.UpsertGameMoveTimeout(ctx, store.UpsertGameMoveTimeoutParams{
		GameID:   gameID,
		PlayerID: playerID,
		TokenID:  uuid.New(),
		Deadline: deadline,
	})
	if err != nil {
		return err
	}

	fn := c.defaultMove(gameID, playerID, gt.TokenID)

	k := NewMoveTimeoutKey(playerID)
	c.timeoutMan.Add(k, deadline, fn)
	return nil
}

func (c *Controller) defaultDemand(gameID, playerID, tokenID uuid.UUID, demandID string) func() {
	return func() {
		ctx := context.Background()

		var action monopoly_deal.Action
		var err error
		var deadline time.Time
		applied := false

		err = c.store.ExecTx(ctx, func(q *store.Queries) error {
			gt, err := q.GetGameTimeoutForUpdate(ctx, store.GetGameTimeoutForUpdateParams{
				GameID:   gameID,
				PlayerID: playerID,
				DemandID: &demandID,
			})
			if err != nil {
				if errors.DBErrorCode(err) == errors.NoDataFound {
					return nil
				}
				return err
			}

			if gt.TokenID != tokenID {
				return nil
			}

			applied = true

			g, err := q.GetGameByPlayerForUpdate(ctx, gt.PlayerID)
			if err != nil {
				return err
			}

			game, err := monopoly_deal.DecodeMsgpack(g.GameState)
			if err != nil {
				return err
			}

			action, err = game.DefaultDemand(playerID, monopoly_deal.Identifier(demandID))
			if err != nil {
				return err
			}

			buf, err := game.EncodeMsgpack()
			if err != nil {
				return err
			}

			g, err = q.UpdateGameState(ctx, store.UpdateGameStateParams{
				GameState: buf,
				GameID:    gameID,
			})
			if err != nil {
				return err
			}

			buf, err = msgpack.Marshal(action)
			if err != nil {
				return err
			}

			_, err = q.CreateGameHistory(ctx, store.CreateGameHistoryParams{
				GameID:        gameID,
				SeqNum:        int16(action.GetSeqNum()),
				ActionKind:    string(action.GetKind()),
				ActionVersion: int32(action.GetVersion()),
				Action:        buf,
			})
			if err != nil {
				return err
			}

			_, err = q.DeleteGameTimeout(ctx, store.DeleteGameTimeoutParams{
				GameID:   gameID,
				PlayerID: playerID,
				DemandID: &demandID,
			})
			if err != nil {
				return err
			}

			if len(game.Demands) == 0 {
				currPlayerID := game.Players[game.CurrPlayerIdx]
				deadline = time.Now().Add(game.Config.MoveTimeout)
				err = c.scheduleDefaultMove(ctx, q, gameID, currPlayerID, deadline)
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		if !applied || action == nil {
			return
		}

		actionProto := action.Proto()
		if deadline.After(time.Now()) {
			actionProto.TurnDeadlineMs = deadline.UnixMilli()
		}

		e := &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
					Payload: &monopoly_deal_schema.ServerMessage_Action{
						Action: actionProto,
					},
				},
			},
		}

		buf, err := proto.Marshal(e)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = c.bus.Publish(ctx, event.GameChannelPre+gameID.String(), event.NewMonopolyDealEvent(buf))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (c *Controller) scheduleDefaultDemand(ctx context.Context, q *store.Queries, gameID, playerID uuid.UUID, demandID string, deadline time.Time) error {
	gt, err := q.UpsertGameDemandTimeout(ctx, store.UpsertGameDemandTimeoutParams{
		GameID:   gameID,
		PlayerID: playerID,
		DemandID: &demandID,
		TokenID:  uuid.New(),
		Deadline: deadline,
	})
	if err != nil {
		return err
	}

	fn := c.defaultDemand(gameID, playerID, gt.TokenID, demandID)

	k := NewDemandTimeoutKey(playerID, demandID)
	c.timeoutMan.Add(k, deadline, fn)
	return nil
}
