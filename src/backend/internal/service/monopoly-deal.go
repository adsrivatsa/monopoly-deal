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

func (c *Controller) maskMonopolyDealPrivateEvents(tp token.Payload, msg *monopoly_deal_schema.ServerMessage) *schema.ServerMessage {
	switch p := msg.Payload.(type) {
	case *monopoly_deal_schema.ServerMessage_StartTurnRes:
		if p.StartTurnRes.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_StartTurnMaskedRes{
					StartTurnMaskedRes: &monopoly_deal_schema.StartTurnMaskedRes{
						SeqNum:   msg.GetStartTurnRes().GetSeqNum(),
						PlayerId: msg.GetStartTurnRes().GetPlayerId(),
						NumCards: int32(len(msg.GetStartTurnRes().GetCards())),
					},
				},
			}
		}
	case *monopoly_deal_schema.ServerMessage_PlayPassGoRes:
		if p.PlayPassGoRes.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_PlayPassGoMaskedRes{
					PlayPassGoMaskedRes: &monopoly_deal_schema.PlayPassGoMaskedRes{
						SeqNum:         msg.GetPlayPassGoRes().GetSeqNum(),
						PlayerId:       msg.GetPlayPassGoRes().GetPlayerId(),
						NumCards:       int32(len(msg.GetPlayPassGoRes().GetCards())),
						LastPlayedCard: msg.GetPlayPassGoRes().GetLastPlayedCard(),
					},
				},
			}
		}

	case *monopoly_deal_schema.ServerMessage_DemandCreated:
		if p.DemandCreated.Demand.PlayerId != tp.PlayerID.String() {
			return nil
		}

	case *monopoly_deal_schema.ServerMessage_PendingRentCreated:
		if p.PendingRentCreated.PendingRent.PlayerId != tp.PlayerID.String() {
			return nil
		}

	case *monopoly_deal_schema.ServerMessage_PendingRentResolved:
		if p.PendingRentResolved.PlayerId != tp.PlayerID.String() {
			return nil
		}

	case *monopoly_deal_schema.ServerMessage_DiscardCardsRes:
		if p.DiscardCardsRes.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_DiscardCardsMaskedRes{
					DiscardCardsMaskedRes: &monopoly_deal_schema.DiscardCardsMaskedRes{
						SeqNum:   msg.GetDiscardCardsRes().GetSeqNum(),
						PlayerId: msg.GetDiscardCardsRes().GetPlayerId(),
						NumCards: int32(len(msg.GetDiscardCardsRes().GetCards())),
					},
				},
			}
		}

	}

	return &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: msg,
		},
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
	case *monopoly_deal_schema.ClientMessage_PlayItsMyBirthday:
		return c.handleMonopolyDealPlayItsMyBirthday(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_ComplyPaymentDemand:
		return c.handleMonopolyDealComplyPaymentDemand(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayDebtCollector:
		return c.handleMonopolyDealPlayDebtCollector(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayRent:
		return c.handleMonopolyDealPlayRent(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayDoubleTheRent:
		return c.handleMonopolyDealDoubleTheRent(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_ResolvePendingRent:
		return c.handleMonopolyDealResolveRent(ctx, tp)
	case *monopoly_deal_schema.ClientMessage_RearrangeCard:
		return c.handleMonopolyDealRearrangeCard(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_DiscardCards:
		return c.handleMonopolyDealDiscardCards(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayWildRent:
		return c.handleMonopolyDealPlayWildRent(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_DenyDemand:
		return c.handleMonopolyDealDenyDemand(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlaySlyDeal:
		return c.handlerMonopolyDealPlaySlyDeal(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_ComplyPropertyDemand:
		return c.handleMonopolyDealComplyPropertyDemand(ctx, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayForcedDeal:
		return c.handleMonopolyDealPlayForcedDeal(ctx, tp, p)
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

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

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

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

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

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

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

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

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

func (c *Controller) handleMonopolyDealPlayItsMyBirthday(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayItsMyBirthday) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.PlayItsMyBirthday.CardId)
	demands, card, err := game.PlayItsMyBirthday(tp.PlayerID, cardID)
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
				Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
					PlayActionRes: &monopoly_deal_schema.PlayActionRes{
						SeqNum:         int32(game.SequenceNum),
						PlayerId:       tp.PlayerID.String(),
						LastPlayedCard: card.Proto(),
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
	if err != nil {
		return err
	}

	for _, d := range demands {
		targetUUID, _ := game.IDTranslator.GetUUID(d.TargetID)

		e := &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
					Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
						DemandCreated: &monopoly_deal_schema.DemandCreated{
							Demand: d.Proto(tp.PlayerID, targetUUID),
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
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) handleMonopolyDealPlayDebtCollector(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayDebtCollector) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	targetUUID, err := uuid.Parse(msg.PlayDebtCollector.TargetId)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.PlayDebtCollector.CardId)
	demands, card, err := game.PlayDebtCollector(tp.PlayerID, targetUUID, cardID)
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
				Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
					PlayActionRes: &monopoly_deal_schema.PlayActionRes{
						SeqNum:         int32(game.SequenceNum),
						PlayerId:       tp.PlayerID.String(),
						LastPlayedCard: card.Proto(),
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
	if err != nil {
		return err
	}

	for _, d := range demands {
		targetUUID, _ := game.IDTranslator.GetUUID(d.TargetID)

		e := &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
					Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
						DemandCreated: &monopoly_deal_schema.DemandCreated{
							Demand: d.Proto(tp.PlayerID, targetUUID),
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
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) handleMonopolyDealPlayRent(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayRent) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.PlayRent.CardId)
	pendingRent, card, err := game.PlayRent(tp.PlayerID, cardID)
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
				Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
					PlayActionRes: &monopoly_deal_schema.PlayActionRes{
						SeqNum:         int32(game.SequenceNum),
						PlayerId:       tp.PlayerID.String(),
						LastPlayedCard: card.Proto(),
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
	if err != nil {
		return err
	}

	targetUUIDs := make([]uuid.UUID, 0, len(pendingRent.TargetIDs))
	for _, targetID := range pendingRent.TargetIDs {
		targetUUID, _ := game.IDTranslator.GetUUID(targetID)
		targetUUIDs = append(targetUUIDs, targetUUID)
	}

	e = &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_PendingRentCreated{
					PendingRentCreated: &monopoly_deal_schema.PendingRentCreated{
						PendingRent: pendingRent.Proto(tp.PlayerID, targetUUIDs),
					},
				},
			},
		},
	}

	buf, err = proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
}

func (c *Controller) handleMonopolyDealDoubleTheRent(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayDoubleTheRent) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.PlayDoubleTheRent.CardId)
	pendingRent, card, err := game.PlayDoubleTheRent(tp.PlayerID, cardID)
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
				Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
					PlayActionRes: &monopoly_deal_schema.PlayActionRes{
						SeqNum:         int32(game.SequenceNum),
						PlayerId:       tp.PlayerID.String(),
						LastPlayedCard: card.Proto(),
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
	if err != nil {
		return err
	}

	targetUUIDs := make([]uuid.UUID, 0, len(pendingRent.TargetIDs))
	for _, targetID := range pendingRent.TargetIDs {
		targetUUID, _ := game.IDTranslator.GetUUID(targetID)
		targetUUIDs = append(targetUUIDs, targetUUID)
	}

	e = &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_PendingRentCreated{
					PendingRentCreated: &monopoly_deal_schema.PendingRentCreated{
						PendingRent: pendingRent.Proto(tp.PlayerID, targetUUIDs),
					},
				},
			},
		},
	}

	buf, err = proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
}

func (c *Controller) handleMonopolyDealResolveRent(ctx context.Context, tp token.Payload) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	demands, err := game.ResolvePendingRent(tp.PlayerID)
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
				Payload: &monopoly_deal_schema.ServerMessage_PendingRentResolved{
					PendingRentResolved: &monopoly_deal_schema.PendingRentResolved{
						SeqNum:   int32(game.SequenceNum),
						PlayerId: tp.PlayerID.String(),
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
	if err != nil {
		return err
	}

	for _, d := range demands {
		targetUUID, _ := game.IDTranslator.GetUUID(d.TargetID)

		e := &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
					Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
						DemandCreated: &monopoly_deal_schema.DemandCreated{
							Demand: d.Proto(tp.PlayerID, targetUUID),
						},
					},
				},
			},
		}

		buf, err = proto.Marshal(e)
		if err != nil {
			return err
		}

		err = c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) handleMonopolyDealComplyPaymentDemand(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_ComplyPaymentDemand) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	demandID := monopoly_deal.Identifier(msg.ComplyPaymentDemand.DemandId)
	cardIDs := make([]monopoly_deal.Identifier, 0, len(game.Cards))
	for _, cardID := range msg.ComplyPaymentDemand.CardIds {
		cardIDs = append(cardIDs, monopoly_deal.Identifier(cardID))
	}
	demandSourceID, cards, propSets, err := game.ComplyPaymentDemand(tp.PlayerID, demandID, cardIDs...)
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
				Payload: &monopoly_deal_schema.ServerMessage_CompliedDemand{
					CompliedDemand: &monopoly_deal_schema.CompliedDemand{
						SeqNum:   int32(game.SequenceNum),
						DemandId: string(demandID),
						PlayerId: tp.PlayerID.String(),
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
	if err != nil {
		return err
	}

	e = &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_TransferCards{
					TransferCards: &monopoly_deal_schema.TransferCards{
						SourceId:     tp.PlayerID.String(),
						TargetId:     demandSourceID.String(),
						Cards:        cards.Proto(),
						PropertySets: propSets.Proto(demandSourceID),
					},
				},
			},
		},
	}

	buf, err = proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
}

func (c *Controller) handleMonopolyDealRearrangeCard(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_RearrangeCard) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.RearrangeCard.CardId)
	var targetSetID *monopoly_deal.Identifier
	if msg.RearrangeCard.PropertySetId != nil {
		id := monopoly_deal.Identifier(*msg.RearrangeCard.PropertySetId)
		targetSetID = &id
	}
	var color *monopoly_deal.Color
	if msg.RearrangeCard.Color != nil {
		id := monopoly_deal.ColorFromProto(*msg.RearrangeCard.Color)
		color = &id
	}
	propertySet, card, err := game.RearrangeProperty(tp.PlayerID, cardID, targetSetID, color)
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
				Payload: &monopoly_deal_schema.ServerMessage_RearrangeCardRes{
					RearrangeCardRes: &monopoly_deal_schema.RearrangeCardRes{
						SeqNum:      int32(game.SequenceNum),
						PlayerId:    tp.PlayerID.String(),
						Card:        card.Proto(),
						PropertySet: propertySet.Proto(tp.PlayerID),
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

func (c *Controller) handleMonopolyDealDiscardCards(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_DiscardCards) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	cardIDs := make([]monopoly_deal.Identifier, 0, len(game.Cards))
	for _, cardID := range msg.DiscardCards.CardIds {
		cardIDs = append(cardIDs, monopoly_deal.Identifier(cardID))
	}
	cards, err := game.DiscardCards(tp.PlayerID, cardIDs...)
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
				Payload: &monopoly_deal_schema.ServerMessage_DiscardCardsRes{
					DiscardCardsRes: &monopoly_deal_schema.DiscardCardsRes{
						SeqNum:   int32(game.SequenceNum),
						PlayerId: tp.PlayerID.String(),
						Cards:    cards.Proto(),
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

func (c *Controller) handleMonopolyDealPlayWildRent(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayWildRent) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	targetUUID, err := uuid.Parse(msg.PlayWildRent.TargetId)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.PlayWildRent.CardId)
	pendingRent, card, err := game.PlayWildRent(tp.PlayerID, targetUUID, cardID)
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
				Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
					PlayActionRes: &monopoly_deal_schema.PlayActionRes{
						SeqNum:         int32(game.SequenceNum),
						PlayerId:       tp.PlayerID.String(),
						LastPlayedCard: card.Proto(),
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
	if err != nil {
		return err
	}

	e = &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_PendingRentCreated{
					PendingRentCreated: &monopoly_deal_schema.PendingRentCreated{
						PendingRent: pendingRent.Proto(tp.PlayerID, []uuid.UUID{targetUUID}),
					},
				},
			},
		},
	}

	buf, err = proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
}

func (c *Controller) handleMonopolyDealDenyDemand(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_DenyDemand) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	demandID := monopoly_deal.Identifier(msg.DenyDemand.DemandId)
	cardID := monopoly_deal.Identifier(msg.DenyDemand.CardId)
	demand, card, err := game.DenyDemand(tp.PlayerID, demandID, cardID)
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
				Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
					PlayActionRes: &monopoly_deal_schema.PlayActionRes{
						SeqNum:         int32(game.SequenceNum),
						PlayerId:       tp.PlayerID.String(),
						LastPlayedCard: card.Proto(),
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
	if err != nil {
		return err
	}

	targetUUID, _ := game.IDTranslator.GetUUID(demand.TargetID)

	e = &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
					DemandCreated: &monopoly_deal_schema.DemandCreated{
						Demand: demand.Proto(tp.PlayerID, targetUUID),
					},
				},
			},
		},
	}

	buf, err = proto.Marshal(e)
	if err != nil {
		return err
	}

	err = c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
	if err != nil {
		return err
	}

	e = &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_DemandDenied{
					DemandDenied: &monopoly_deal_schema.DemandDenied{
						SeqNum:   int32(game.SequenceNum),
						DemandId: string(demandID),
					},
				},
			},
		},
	}

	buf, err = proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
}

func (c *Controller) handlerMonopolyDealPlaySlyDeal(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlaySlyDeal) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.PlaySlyDeal.CardId)
	targetID, err := uuid.Parse(msg.PlaySlyDeal.TargetId)
	if err != nil {
		return err
	}
	targetCardID := monopoly_deal.Identifier(msg.PlaySlyDeal.TargetCardId)

	demands, card, err := game.PlaySlyDeal(tp.PlayerID, targetID, cardID, targetCardID)
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
				Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
					PlayActionRes: &monopoly_deal_schema.PlayActionRes{
						SeqNum:         int32(game.SequenceNum),
						PlayerId:       tp.PlayerID.String(),
						LastPlayedCard: card.Proto(),
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
	if err != nil {
		return err
	}

	for _, d := range demands {
		targetUUID, _ := game.IDTranslator.GetUUID(d.TargetID)

		e := &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
					Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
						DemandCreated: &monopoly_deal_schema.DemandCreated{
							Demand: d.Proto(tp.PlayerID, targetUUID),
						},
					},
				},
			},
		}

		buf, err = proto.Marshal(e)
		if err != nil {
			return err
		}

		err = c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) handleMonopolyDealComplyPropertyDemand(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_ComplyPropertyDemand) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	demandID := monopoly_deal.Identifier(msg.ComplyPropertyDemand.DemandId)
	sourceUUID, sourcePropertySets, targetPropertySets, err := game.ComplyPropertyDemand(tp.PlayerID, demandID)
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
				Payload: &monopoly_deal_schema.ServerMessage_CompliedDemand{
					CompliedDemand: &monopoly_deal_schema.CompliedDemand{
						SeqNum:   int32(game.SequenceNum),
						DemandId: string(demandID),
						PlayerId: tp.PlayerID.String(),
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
	if err != nil {
		return err
	}

	e = &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_TransferProperty{
					TransferProperty: &monopoly_deal_schema.TransferProperty{
						SeqNum:             int32(game.SequenceNum),
						SourceId:           sourceUUID.String(),
						TargetId:           tp.PlayerID.String(),
						SourcePropertySets: sourcePropertySets.Proto(sourceUUID),
						TargetPropertySets: targetPropertySets.Proto(tp.PlayerID),
					},
				},
			},
		},
	}

	buf, err = proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
}

func (c *Controller) handleMonopolyDealPlayForcedDeal(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayForcedDeal) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	lock := c.getGameLock(g.GameID)
	lock.Lock()
	defer lock.Unlock()

	var game monopoly_deal.Game
	err = msgpack.Unmarshal(g.GameState, &game)
	if err != nil {
		return err
	}

	cardID := monopoly_deal.Identifier(msg.PlayForcedDeal.CardId)
	targetID, err := uuid.Parse(msg.PlayForcedDeal.TargetId)
	if err != nil {
		return err
	}
	sourceCardID := monopoly_deal.Identifier(msg.PlayForcedDeal.SourceCardId)
	targetCardID := monopoly_deal.Identifier(msg.PlayForcedDeal.TargetCardId)

	demands, card, err := game.PlayForcedDeal(tp.PlayerID, targetID, cardID, sourceCardID, targetCardID)
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
				Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
					PlayActionRes: &monopoly_deal_schema.PlayActionRes{
						SeqNum:         int32(game.SequenceNum),
						PlayerId:       tp.PlayerID.String(),
						LastPlayedCard: card.Proto(),
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
	if err != nil {
		return err
	}

	for _, d := range demands {
		targetUUID, _ := game.IDTranslator.GetUUID(d.TargetID)

		e := &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
					Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
						DemandCreated: &monopoly_deal_schema.DemandCreated{
							Demand: d.Proto(tp.PlayerID, targetUUID),
						},
					},
				},
			},
		}

		buf, err = proto.Marshal(e)
		if err != nil {
			return err
		}

		err = c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
		if err != nil {
			return err
		}
	}

	return nil
}
