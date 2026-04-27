package monopoly_deal

import (
	"context"
	stderrors "errors"
	"fmt"
	monopoly_deal "the-deal/internal/engine/monopoly-deal"
	"the-deal/internal/errors"
	"the-deal/internal/schema"
	"the-deal/internal/schema/monopoly_deal_schema"
	"the-deal/internal/store"
	"the-deal/internal/token"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

func protoMonopolyDealError(err error) *schema.ServerMessage {
	var intErr errors.Error
	if !stderrors.As(err, &intErr) {
		intErr = errors.Internal(err)
	}

	return &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_Error{
					Error: &monopoly_deal_schema.Error{
						Message: intErr.Error(),
						Status:  int32(intErr.Status),
						Code:    intErr.Code,
					},
				},
			},
		},
	}
}

func (c *Controller) GetGameState(ctx context.Context, tp token.Payload) *schema.ServerMessage {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return protoMonopolyDealError(errors.EntityNotFound(errors.EntityGame, err))
		}
		return protoMonopolyDealError(err)
	}

	ps, err := c.store.GetPlayersByGame(ctx, g.GameID)
	if err != nil {
		return protoMonopolyDealError(err)
	}

	game, err := monopoly_deal.DecodeMsgpack(g.GameState)
	if err != nil {
		return protoMonopolyDealError(err)
	}

	playerIDs := make([]uuid.UUID, 0, len(ps))
	playerProtos := make([]*monopoly_deal_schema.Player, 0, len(ps))
	for _, p := range ps {
		playerIDs = append(playerIDs, p.PlayerID)

		money, _ := game.CountMoney(p.PlayerID)
		completedSets, _ := game.CountCompletedSets(p.PlayerID)
		handLen, _ := game.CountHands(p.PlayerID)

		playerProtos = append(playerProtos, &monopoly_deal_schema.Player{
			PlayerId:      p.PlayerID.String(),
			DisplayName:   p.DisplayName,
			AvatarUrl:     p.ImageUrl,
			Money:         int32(money),
			CompletedSets: int32(completedSets),
			HandCards:     int32(handLen),
		})
	}

	gameState := game.Proto(tp.PlayerID, playerIDs)
	gameState.Players = playerProtos

	return &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_GameState{
					GameState: gameState,
				},
			},
		},
	}
}

func (c *Controller) ListGameHistory(ctx context.Context, tp token.Payload, gameID uuid.UUID, callback func(message *schema.ServerMessage)) error {
	limit := 25
	offset := 0

	for {
		ghs, err := c.store.ListGameHistory(ctx, store.ListGameHistoryParams{
			GameID: gameID,
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
			case monopoly_deal.ActionKindDemandCreated:
				action = new(monopoly_deal.ActionDemandsCreated)
			case monopoly_deal.ActionKindPendingRentCreated:
				action = new(monopoly_deal.ActionPendingRentCreated)
			case monopoly_deal.ActionKindDemandComplied:
				action = new(monopoly_deal.ActionDemandComplied)
			case monopoly_deal.ActionKindDiscardCards:
				action = new(monopoly_deal.ActionDiscardCards)
			case monopoly_deal.ActionKindStartTurn:
				action = new(monopoly_deal.ActionStartTurn)
			case monopoly_deal.ActionKindRearrangeCard:
				action = new(monopoly_deal.ActionRearrangeCard)
			default:
				return fmt.Errorf("unsupported action kind: %s", gh.ActionKind)
			}
			err = msgpack.Unmarshal(gh.Action, action)
			if err != nil {
				return err
			}

			msg := c.MaskEvent(tp, &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_ActionHistory{
					ActionHistory: &monopoly_deal_schema.ActionHistory{
						Action: action.Proto(),
					},
				},
			})

			callback(msg)
		}

		if len(ghs) < limit {
			break
		}

		offset += limit
	}

	return nil
}
