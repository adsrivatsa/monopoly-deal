package service

import (
	"context"
	stderrors "errors"
	monopoly_deal "fun-kames/internal/engine/monopoly-deal"
	"fun-kames/internal/errors"
	"fun-kames/internal/event"
	"fun-kames/internal/schema"
	"fun-kames/internal/schema/monopoly_deal_schema"
	"fun-kames/internal/store"
	"fun-kames/internal/token"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
)

func (c *Controller) maskMonopolyDealPrivateEvents(tp token.Payload, msg *monopoly_deal_schema.ServerMessage) *monopoly_deal_schema.ServerMessage {
	switch p := msg.Payload.(type) {
	case *monopoly_deal_schema.ServerMessage_StartTurnRes:
		if p.StartTurnRes.PlayerId == tp.PlayerID.String() {
			return msg
		}

		return &monopoly_deal_schema.ServerMessage{
			Payload: &monopoly_deal_schema.ServerMessage_StartTurnMaskedRes{
				StartTurnMaskedRes: &monopoly_deal_schema.StartTurnMaskedRes{
					SeqNum:   msg.GetStartTurnRes().GetSeqNum(),
					PlayerId: msg.GetStartTurnRes().GetPlayerId(),
					NumCards: int32(len(msg.GetStartTurnRes().GetCards())),
				},
			},
		}
	case *monopoly_deal_schema.ServerMessage_PlayPassGoRes:
		if p.PlayPassGoRes.PlayerId == tp.PlayerID.String() {
			return msg
		}

		return &monopoly_deal_schema.ServerMessage{
			Payload: &monopoly_deal_schema.ServerMessage_PlayPassGoMaskedRes{
				PlayPassGoMaskedRes: &monopoly_deal_schema.PlayPassGoMaskedRes{
					SeqNum:         msg.GetPlayPassGoRes().GetSeqNum(),
					PlayerId:       msg.GetPlayPassGoRes().GetPlayerId(),
					NumCards:       int32(len(msg.GetPlayPassGoRes().GetCards())),
					LastPlayedCard: msg.GetPlayPassGoRes().GetLastPlayedCard(),
				},
			},
		}
	default:
		return msg
	}
}

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
	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
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

func (c *Controller) HandleMonopolyDealEvent(ctx context.Context, tp token.Payload, msg *schema.ClientMessage_MonopolyDealMessage) error {
	switch p := msg.MonopolyDealMessage.GetPayload().(type) {
	case *monopoly_deal_schema.ClientMessage_Chat:
		return c.handleMonopolyDealChat(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayMoney:
		return c.handleMonopolyDealPlayMoney(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayProperty:
		return c.handleMonopolyDealPlayProperty(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_CompleteTurn:
		return c.handleMonopolyDealCompleteTurn(ctx, tp)
	case *monopoly_deal_schema.ClientMessage_PlayPassGo:
		return c.handleMonopolyDealPlayPassGo(ctx, tp, p)
	default:
		return nil
	}
}

func (c *Controller) handleMonopolyDealChat(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_Chat) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	e := &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_ChatReceived{
					ChatReceived: &monopoly_deal_schema.ChatReceived{
						PlayerId: tp.PlayerID.String(),
						Payload:  msg.Chat.Payload,
					},
				},
			},
		},
	}

	buf, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	err = c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
	return err
}

func (c *Controller) handleMonopolyDealPlayMoney(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayMoney) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.PlayMoney.CardId)
	card, err := game.PlayMoney(tp.PlayerID, cardID)
	if err != nil {
		return err
	}

	gameState, err := game.EncodeMsgpack()
	if err != nil {
		return err
	}

	g, err = c.store.UpdateGameState(ctx, store.UpdateGameStateParams{
		GameState: gameState,
		GameID:    g.GameID,
	})
	if err != nil {
		return err
	}

	e := &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_PlayMoneyRes{
					PlayMoneyRes: &monopoly_deal_schema.PlayMoneyRes{
						SeqNum:   int32(game.SequenceNum),
						PlayerId: tp.PlayerID.String(),
						Card:     card.Proto(),
					},
				},
			},
		},
	}

	buf, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
}

func (c *Controller) handleMonopolyDealPlayProperty(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayProperty) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.PlayProperty.CardId)
	var propSetID *monopoly_deal.Identifier
	if msg.PlayProperty.PropertySetId != nil {
		id := monopoly_deal.Identifier(*msg.PlayProperty.PropertySetId)
		propSetID = &id
	}
	var propSetColor *monopoly_deal.Color
	if msg.PlayProperty.ActiveColor != nil {
		id := monopoly_deal.ColorFromProto(*msg.PlayProperty.ActiveColor)
		propSetColor = &id
	}
	propSet, err := game.PlayProperty(tp.PlayerID, cardID, propSetID, propSetColor)
	if err != nil {
		return err
	}

	gameState, err := game.EncodeMsgpack()
	if err != nil {
		return err
	}

	g, err = c.store.UpdateGameState(ctx, store.UpdateGameStateParams{
		GameState: gameState,
		GameID:    g.GameID,
	})
	if err != nil {
		return err
	}

	e := &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_PlayPropertyRes{
					PlayPropertyRes: &monopoly_deal_schema.PlayPropertyRes{
						SeqNum:      int32(game.SequenceNum),
						PlayerId:    tp.PlayerID.String(),
						PropertySet: propSet.Proto(tp.PlayerID),
					},
				},
			},
		},
	}

	buf, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
}

func (c *Controller) handleMonopolyDealCompleteTurn(ctx context.Context, tp token.Payload) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	drawn, nextPlayerID, err := game.CompleteTurn(tp.PlayerID)
	if err != nil {
		return err
	}

	gameState, err := game.EncodeMsgpack()
	if err != nil {
		return err
	}

	g, err = c.store.UpdateGameState(ctx, store.UpdateGameStateParams{
		GameState: gameState,
		GameID:    g.GameID,
	})
	if err != nil {
		return err
	}

	e := &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_StartTurnRes{
					StartTurnRes: &monopoly_deal_schema.StartTurnRes{
						SeqNum:    int32(game.SequenceNum),
						PlayerId:  nextPlayerID.String(),
						Cards:     drawn.Proto(),
						MovesLeft: int32(game.Config.MovesPerTurn),
					},
				},
			},
		},
	}

	buf, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
}

func (c *Controller) handleMonopolyDealPlayPassGo(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayPassGo) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.PlayPassGo.CardId)
	drawn, err := game.PlayPassGo(tp.PlayerID, cardID)
	if err != nil {
		return err
	}

	gameState, err := game.EncodeMsgpack()
	if err != nil {
		return err
	}

	g, err = c.store.UpdateGameState(ctx, store.UpdateGameStateParams{
		GameState: gameState,
		GameID:    g.GameID,
	})
	if err != nil {
		return err
	}

	e := &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_PlayPassGoRes{
					PlayPassGoRes: &monopoly_deal_schema.PlayPassGoRes{
						SeqNum:         int32(game.SequenceNum),
						PlayerId:       tp.PlayerID.String(),
						Cards:          drawn.Proto(),
						LastPlayedCard: game.LastAction.Proto(),
					},
				},
			},
		},
	}

	buf, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
}
