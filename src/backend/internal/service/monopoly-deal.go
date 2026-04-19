package service

import (
	"context"
	stderrors "errors"
	monopoly_deal "fun-kames/internal/engine/monopoly-deal"
	"fun-kames/internal/errors"
	"fun-kames/internal/schema"
	"fun-kames/internal/schema/monopoly_deal_schema"
	"fun-kames/internal/token"

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
						Message: intErr.Message,
						Status:  int32(intErr.Status),
						Code:    intErr.Code,
					},
				},
			},
		},
	}
}

func (c *Controller) GetMonopolyDealGame(ctx context.Context, tp token.Payload) *schema.ServerMessage {
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

	playerIDs := make([]uuid.UUID, 0, len(ps))
	playerProtos := make([]*monopoly_deal_schema.Player, 0, len(ps))
	for _, p := range ps {
		playerIDs = append(playerIDs, p.PlayerID)
		playerProtos = append(playerProtos, &monopoly_deal_schema.Player{
			PlayerId:    p.PlayerID.String(),
			DisplayName: p.DisplayName,
			AvatarUrl:   p.ImageUrl,
		})
	}

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return protoMonopolyDealError(err)
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
