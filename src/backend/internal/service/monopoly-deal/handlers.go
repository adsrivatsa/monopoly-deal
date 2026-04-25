package monopoly_deal

import (
	"context"
	"sync"
	"the-deal/internal/config"
	monopoly_deal "the-deal/internal/engine/monopoly-deal"
	"the-deal/internal/errors"
	"the-deal/internal/event"
	"the-deal/internal/schema"
	"the-deal/internal/schema/monopoly_deal_schema"
	"the-deal/internal/store"
	"the-deal/internal/token"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
)

type Controller struct {
	cfg       config.Config
	store     store.Store
	bus       *event.Bus
	mu        sync.Mutex
	gameLocks map[uuid.UUID]*sync.RWMutex
}

func NewController(cfg config.Config, pool *pgxpool.Pool, client *redis.Client) *Controller {
	c := &Controller{
		cfg:       cfg,
		store:     store.NewSQLStore(pool, nil),
		bus:       event.NewBus(client),
		gameLocks: make(map[uuid.UUID]*sync.RWMutex),
	}

	return c
}

func (c *Controller) getGameLock(gameID uuid.UUID) *sync.RWMutex {
	c.mu.Lock()
	defer c.mu.Unlock()
	lock, ok := c.gameLocks[gameID]
	if !ok {
		lock = &sync.RWMutex{}
		c.gameLocks[gameID] = lock
	}
	return lock
}

func (c *Controller) HandleEvent(ctx context.Context, tp token.Payload, msg *schema.ClientMessage_MonopolyDealMessage) error {
	switch p := msg.MonopolyDealMessage.GetPayload().(type) {
	case *monopoly_deal_schema.ClientMessage_Chat:
		return c.handleChat(ctx, tp, p)
	default:
		return c.handleGameEvent(ctx, tp, msg)
	}
}

func (c *Controller) handleChat(ctx context.Context, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_Chat) error {
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

func (c *Controller) handleGameEvent(ctx context.Context, tp token.Payload, msg *schema.ClientMessage_MonopolyDealMessage) error {
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

	var events []*monopoly_deal_schema.ServerMessage

	switch p := msg.MonopolyDealMessage.GetPayload().(type) {
	case *monopoly_deal_schema.ClientMessage_PlayMoney:
		events, err = c.handlePlayMoney(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayProperty:
		events, err = c.handlePlayProperty(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_CompleteTurn:
		events, err = c.handleCompleteTurn(&game, tp)
	case *monopoly_deal_schema.ClientMessage_PlayPassGo:
		events, err = c.handlePlayPassGo(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayItsMyBirthday:
		events, err = c.handlePlayItsMyBirthday(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_ComplyPaymentDemand:
		events, err = c.handleComplyPaymentDemand(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayDebtCollector:
		events, err = c.handlePlayDebtCollector(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayRent:
		events, err = c.handlePlayRent(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayDoubleTheRent:
		events, err = c.handleDoubleTheRent(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_ResolvePendingRent:
		events, err = c.handleResolveRent(&game, tp)
	case *monopoly_deal_schema.ClientMessage_RearrangeCard:
		events, err = c.handleRearrangeCard(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_DiscardCards:
		events, err = c.handleDiscardCards(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayWildRent:
		events, err = c.handlePlayWildRent(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_DenyDemand:
		events, err = c.handleDenyDemand(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlaySlyDeal:
		events, err = c.handlerMonopolyDealPlaySlyDeal(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_ComplyPropertyDemand:
		events, err = c.handleComplyPropertyDemand(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayForcedDeal:
		events, err = c.handlePlayForcedDeal(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayDealBreaker:
		events, err = c.handlePlayDealBreaker(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_ComplyPropertySetDemand:
		events, err = c.handleComplyPropertySetDemand(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayHouse:
		events, err = c.handlePlayHouse(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayHotel:
		events, err = c.handlePlayHotel(&game, tp, p)
	}
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

	completeSets, moneyValue, didWin, err := game.CheckWinConditions(tp.PlayerID)
	if err != nil {
		return err
	}

	if didWin {
		events = append(events, &monopoly_deal_schema.ServerMessage{
			Payload: &monopoly_deal_schema.ServerMessage_WonGame{
				WonGame: &monopoly_deal_schema.WonGame{
					PlayerId:         tp.PlayerID.String(),
					NumCompletedSets: int32(completeSets),
					Money:            int32(moneyValue),
				},
			},
		})

		_, err = c.store.CompleteGame(ctx, store.CompleteGameParams{
			Winner: &tp.PlayerID,
			GameID: g.GameID,
		})
		if err != nil {
			return err
		}
	}

	for _, e := range events {
		wrappedEvent := &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: e,
			},
		}

		buf, err := proto.Marshal(wrappedEvent)
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

func (c *Controller) handlePlayMoney(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayMoney) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardID := monopoly_deal.Identifier(msg.PlayMoney.CardId)
	card, err := game.PlayMoney(tp.PlayerID, cardID)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayMoneyRes{
				PlayMoneyRes: &monopoly_deal_schema.PlayMoneyRes{
					SeqNum:   int32(game.SequenceNum),
					PlayerId: tp.PlayerID.String(),
					Card:     card.Proto(),
				},
			},
		},
	}, nil

}

func (c *Controller) handlePlayProperty(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayProperty) ([]*monopoly_deal_schema.ServerMessage, error) {
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
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayPropertyRes{
				PlayPropertyRes: &monopoly_deal_schema.PlayPropertyRes{
					SeqNum:      int32(game.SequenceNum),
					PlayerId:    tp.PlayerID.String(),
					PropertySet: propSet.Proto(tp.PlayerID),
				},
			},
		},
	}, nil
}

func (c *Controller) handleCompleteTurn(game *monopoly_deal.Game, tp token.Payload) ([]*monopoly_deal_schema.ServerMessage, error) {
	drawn, nextPlayerID, err := game.CompleteTurn(tp.PlayerID)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_StartTurnRes{
				StartTurnRes: &monopoly_deal_schema.StartTurnRes{
					SeqNum:    int32(game.SequenceNum),
					PlayerId:  nextPlayerID.String(),
					Cards:     drawn.Proto(),
					MovesLeft: int32(game.Config.MovesPerTurn),
				},
			},
		},
	}, nil
}

func (c *Controller) handlePlayPassGo(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayPassGo) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardID := monopoly_deal.Identifier(msg.PlayPassGo.CardId)
	drawn, err := game.PlayPassGo(tp.PlayerID, cardID)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayPassGoRes{
				PlayPassGoRes: &monopoly_deal_schema.PlayPassGoRes{
					SeqNum:         int32(game.SequenceNum),
					PlayerId:       tp.PlayerID.String(),
					Cards:          drawn.Proto(),
					LastPlayedCard: game.LastAction.Proto(),
				},
			},
		},
	}, nil
}

func (c *Controller) handlePlayItsMyBirthday(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayItsMyBirthday) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardID := monopoly_deal.Identifier(msg.PlayItsMyBirthday.CardId)
	demands, card, err := game.PlayItsMyBirthday(tp.PlayerID, cardID)
	if err != nil {
		return nil, err
	}

	es := []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
				PlayActionRes: &monopoly_deal_schema.PlayActionRes{
					SeqNum:         int32(game.SequenceNum),
					PlayerId:       tp.PlayerID.String(),
					LastPlayedCard: card.Proto(),
				},
			},
		},
	}

	for _, d := range demands {
		targetUUID, _ := game.IDTranslator.GetUUID(d.TargetID)

		es = append(es, &monopoly_deal_schema.ServerMessage{
			Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
				DemandCreated: &monopoly_deal_schema.DemandCreated{
					Demand: d.Proto(tp.PlayerID, targetUUID),
				},
			},
		})
	}

	return es, nil
}

func (c *Controller) handlePlayDebtCollector(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayDebtCollector) ([]*monopoly_deal_schema.ServerMessage, error) {
	targetUUID, err := uuid.Parse(msg.PlayDebtCollector.TargetId)
	if err != nil {
		return nil, err
	}

	cardID := monopoly_deal.Identifier(msg.PlayDebtCollector.CardId)
	demands, card, err := game.PlayDebtCollector(tp.PlayerID, targetUUID, cardID)
	if err != nil {
		return nil, err
	}

	es := []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
				PlayActionRes: &monopoly_deal_schema.PlayActionRes{
					SeqNum:         int32(game.SequenceNum),
					PlayerId:       tp.PlayerID.String(),
					LastPlayedCard: card.Proto(),
				},
			},
		},
	}

	for _, d := range demands {
		targetUUID, _ = game.IDTranslator.GetUUID(d.TargetID)

		es = append(es, &monopoly_deal_schema.ServerMessage{
			Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
				DemandCreated: &monopoly_deal_schema.DemandCreated{
					Demand: d.Proto(tp.PlayerID, targetUUID),
				},
			},
		})
	}

	return es, nil
}

func (c *Controller) handlePlayRent(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayRent) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardID := monopoly_deal.Identifier(msg.PlayRent.CardId)
	pendingRent, card, err := game.PlayRent(tp.PlayerID, cardID)
	if err != nil {
		return nil, err
	}

	targetUUIDs := make([]uuid.UUID, 0, len(pendingRent.TargetIDs))
	for _, targetID := range pendingRent.TargetIDs {
		targetUUID, _ := game.IDTranslator.GetUUID(targetID)
		targetUUIDs = append(targetUUIDs, targetUUID)
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
				PlayActionRes: &monopoly_deal_schema.PlayActionRes{
					SeqNum:         int32(game.SequenceNum),
					PlayerId:       tp.PlayerID.String(),
					LastPlayedCard: card.Proto(),
				},
			},
		},
		{
			Payload: &monopoly_deal_schema.ServerMessage_PendingRentCreated{
				PendingRentCreated: &monopoly_deal_schema.PendingRentCreated{
					PendingRent: pendingRent.Proto(tp.PlayerID, targetUUIDs),
				},
			},
		},
	}, nil
}

func (c *Controller) handleDoubleTheRent(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayDoubleTheRent) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardID := monopoly_deal.Identifier(msg.PlayDoubleTheRent.CardId)
	pendingRent, card, err := game.PlayDoubleTheRent(tp.PlayerID, cardID)
	if err != nil {
		return nil, err
	}

	targetUUIDs := make([]uuid.UUID, 0, len(pendingRent.TargetIDs))
	for _, targetID := range pendingRent.TargetIDs {
		targetUUID, _ := game.IDTranslator.GetUUID(targetID)
		targetUUIDs = append(targetUUIDs, targetUUID)
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
				PlayActionRes: &monopoly_deal_schema.PlayActionRes{
					SeqNum:         int32(game.SequenceNum),
					PlayerId:       tp.PlayerID.String(),
					LastPlayedCard: card.Proto(),
				},
			},
		},
		{
			Payload: &monopoly_deal_schema.ServerMessage_PendingRentCreated{
				PendingRentCreated: &monopoly_deal_schema.PendingRentCreated{
					PendingRent: pendingRent.Proto(tp.PlayerID, targetUUIDs),
				},
			},
		},
	}, nil
}

func (c *Controller) handleResolveRent(game *monopoly_deal.Game, tp token.Payload) ([]*monopoly_deal_schema.ServerMessage, error) {
	demands, err := game.ResolvePendingRent(tp.PlayerID)
	if err != nil {
		return nil, err
	}

	es := []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PendingRentResolved{
				PendingRentResolved: &monopoly_deal_schema.PendingRentResolved{
					SeqNum:   int32(game.SequenceNum),
					PlayerId: tp.PlayerID.String(),
				},
			},
		},
	}

	for _, d := range demands {
		targetUUID, _ := game.IDTranslator.GetUUID(d.TargetID)

		es = append(es, &monopoly_deal_schema.ServerMessage{
			Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
				DemandCreated: &monopoly_deal_schema.DemandCreated{
					Demand: d.Proto(tp.PlayerID, targetUUID),
				},
			},
		})
	}

	return es, nil
}

func (c *Controller) handleComplyPaymentDemand(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_ComplyPaymentDemand) ([]*monopoly_deal_schema.ServerMessage, error) {
	demandID := monopoly_deal.Identifier(msg.ComplyPaymentDemand.DemandId)
	cardIDs := make([]monopoly_deal.Identifier, 0, len(game.Cards))
	for _, cardID := range msg.ComplyPaymentDemand.CardIds {
		cardIDs = append(cardIDs, monopoly_deal.Identifier(cardID))
	}
	demandSourceID, cards, propSets, sourceSets, sourceMoney, targetSets, targetMoney, err := game.ComplyPaymentDemand(tp.PlayerID, demandID, cardIDs...)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_CompliedDemand{
				CompliedDemand: &monopoly_deal_schema.CompliedDemand{
					SeqNum:   int32(game.SequenceNum),
					DemandId: string(demandID),
					PlayerId: tp.PlayerID.String(),
				},
			},
		},
		{
			Payload: &monopoly_deal_schema.ServerMessage_TransferCards{
				TransferCards: &monopoly_deal_schema.TransferCards{
					SourceId:     tp.PlayerID.String(),
					TargetId:     demandSourceID.String(),
					Cards:        cards.Proto(),
					PropertySets: propSets.Proto(demandSourceID),
					SourceSets:   int32(targetSets),
					SourceMoney:  int32(targetMoney),
					TargetSets:   int32(sourceSets),
					TargetMoney:  int32(sourceMoney),
				},
			},
		},
	}, nil
}

func (c *Controller) handleRearrangeCard(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_RearrangeCard) ([]*monopoly_deal_schema.ServerMessage, error) {
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
	propertySet, sets, card, err := game.RearrangeProperty(tp.PlayerID, cardID, targetSetID, color)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{{
		Payload: &monopoly_deal_schema.ServerMessage_RearrangeCardRes{
			RearrangeCardRes: &monopoly_deal_schema.RearrangeCardRes{
				SeqNum:      int32(game.SequenceNum),
				PlayerId:    tp.PlayerID.String(),
				Card:        card.Proto(),
				PropertySet: propertySet.Proto(tp.PlayerID),
				Sets:        int32(sets),
			},
		},
	},
	}, nil
}

func (c *Controller) handleDiscardCards(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_DiscardCards) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardIDs := make([]monopoly_deal.Identifier, 0, len(game.Cards))
	for _, cardID := range msg.DiscardCards.CardIds {
		cardIDs = append(cardIDs, monopoly_deal.Identifier(cardID))
	}
	cards, err := game.DiscardCards(tp.PlayerID, cardIDs...)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_DiscardCardsRes{
				DiscardCardsRes: &monopoly_deal_schema.DiscardCardsRes{
					SeqNum:   int32(game.SequenceNum),
					PlayerId: tp.PlayerID.String(),
					Cards:    cards.Proto(),
				},
			},
		},
	}, nil
}

func (c *Controller) handlePlayWildRent(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayWildRent) ([]*monopoly_deal_schema.ServerMessage, error) {
	targetUUID, err := uuid.Parse(msg.PlayWildRent.TargetId)
	if err != nil {
		return nil, err
	}

	cardID := monopoly_deal.Identifier(msg.PlayWildRent.CardId)
	pendingRent, card, err := game.PlayWildRent(tp.PlayerID, targetUUID, cardID)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
				PlayActionRes: &monopoly_deal_schema.PlayActionRes{
					SeqNum:         int32(game.SequenceNum),
					PlayerId:       tp.PlayerID.String(),
					LastPlayedCard: card.Proto(),
				},
			},
		},
		{
			Payload: &monopoly_deal_schema.ServerMessage_PendingRentCreated{
				PendingRentCreated: &monopoly_deal_schema.PendingRentCreated{
					PendingRent: pendingRent.Proto(tp.PlayerID, []uuid.UUID{targetUUID}),
				},
			},
		},
	}, nil
}

func (c *Controller) handleDenyDemand(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_DenyDemand) ([]*monopoly_deal_schema.ServerMessage, error) {
	demandID := monopoly_deal.Identifier(msg.DenyDemand.DemandId)
	cardID := monopoly_deal.Identifier(msg.DenyDemand.CardId)
	demand, card, err := game.DenyDemand(tp.PlayerID, demandID, cardID)
	if err != nil {
		return nil, err
	}

	targetUUID, _ := game.IDTranslator.GetUUID(demand.TargetID)

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
				PlayActionRes: &monopoly_deal_schema.PlayActionRes{
					SeqNum:         int32(game.SequenceNum),
					PlayerId:       tp.PlayerID.String(),
					LastPlayedCard: card.Proto(),
				},
			},
		},
		{
			Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
				DemandCreated: &monopoly_deal_schema.DemandCreated{
					Demand: demand.Proto(tp.PlayerID, targetUUID),
				},
			},
		},
		{
			Payload: &monopoly_deal_schema.ServerMessage_DemandDenied{
				DemandDenied: &monopoly_deal_schema.DemandDenied{
					SeqNum:   int32(game.SequenceNum),
					DemandId: string(demandID),
				},
			},
		},
	}, nil
}

func (c *Controller) handlerMonopolyDealPlaySlyDeal(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlaySlyDeal) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardID := monopoly_deal.Identifier(msg.PlaySlyDeal.CardId)
	targetID, err := uuid.Parse(msg.PlaySlyDeal.TargetId)
	if err != nil {
		return nil, err
	}
	targetCardID := monopoly_deal.Identifier(msg.PlaySlyDeal.TargetCardId)

	demands, card, err := game.PlaySlyDeal(tp.PlayerID, targetID, cardID, targetCardID)
	if err != nil {
		return nil, err
	}

	es := []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
				PlayActionRes: &monopoly_deal_schema.PlayActionRes{
					SeqNum:         int32(game.SequenceNum),
					PlayerId:       tp.PlayerID.String(),
					LastPlayedCard: card.Proto(),
				},
			},
		},
	}

	for _, d := range demands {
		targetUUID, _ := game.IDTranslator.GetUUID(d.TargetID)

		es = append(es, &monopoly_deal_schema.ServerMessage{
			Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
				DemandCreated: &monopoly_deal_schema.DemandCreated{
					Demand: d.Proto(tp.PlayerID, targetUUID),
				},
			},
		})
	}

	return es, nil
}

func (c *Controller) handleComplyPropertyDemand(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_ComplyPropertyDemand) ([]*monopoly_deal_schema.ServerMessage, error) {
	demandID := monopoly_deal.Identifier(msg.ComplyPropertyDemand.DemandId)
	sourceUUID, sourcePropertySets, targetPropertySets, err := game.ComplyPropertyDemand(tp.PlayerID, demandID)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_CompliedDemand{
				CompliedDemand: &monopoly_deal_schema.CompliedDemand{
					SeqNum:   int32(game.SequenceNum),
					DemandId: string(demandID),
					PlayerId: tp.PlayerID.String(),
				},
			},
		},
		{
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
	}, nil
}

func (c *Controller) handlePlayForcedDeal(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayForcedDeal) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardID := monopoly_deal.Identifier(msg.PlayForcedDeal.CardId)
	targetID, err := uuid.Parse(msg.PlayForcedDeal.TargetId)
	if err != nil {
		return nil, err
	}
	sourceCardID := monopoly_deal.Identifier(msg.PlayForcedDeal.SourceCardId)
	targetCardID := monopoly_deal.Identifier(msg.PlayForcedDeal.TargetCardId)

	demands, card, err := game.PlayForcedDeal(tp.PlayerID, targetID, cardID, sourceCardID, targetCardID)
	if err != nil {
		return nil, err
	}

	es := []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
				PlayActionRes: &monopoly_deal_schema.PlayActionRes{
					SeqNum:         int32(game.SequenceNum),
					PlayerId:       tp.PlayerID.String(),
					LastPlayedCard: card.Proto(),
				},
			},
		},
	}

	for _, d := range demands {
		targetUUID, _ := game.IDTranslator.GetUUID(d.TargetID)

		es = append(es, &monopoly_deal_schema.ServerMessage{
			Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
				DemandCreated: &monopoly_deal_schema.DemandCreated{
					Demand: d.Proto(tp.PlayerID, targetUUID),
				},
			},
		})
	}

	return es, nil
}

func (c *Controller) handlePlayDealBreaker(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayDealBreaker) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardID := monopoly_deal.Identifier(msg.PlayDealBreaker.CardId)
	targetUUID, err := uuid.Parse(msg.PlayDealBreaker.TargetId)
	if err != nil {
		return nil, err
	}
	propertySetID := monopoly_deal.Identifier(msg.PlayDealBreaker.PropertySetId)

	demands, card, err := game.PlayDealBreaker(tp.PlayerID, targetUUID, cardID, propertySetID)
	if err != nil {
		return nil, err
	}

	es := []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_PlayActionRes{
				PlayActionRes: &monopoly_deal_schema.PlayActionRes{
					SeqNum:         int32(game.SequenceNum),
					PlayerId:       tp.PlayerID.String(),
					LastPlayedCard: card.Proto(),
				},
			},
		},
	}

	for _, d := range demands {
		targetUUID, _ = game.IDTranslator.GetUUID(d.TargetID)

		es = append(es, &monopoly_deal_schema.ServerMessage{
			Payload: &monopoly_deal_schema.ServerMessage_DemandCreated{
				DemandCreated: &monopoly_deal_schema.DemandCreated{
					Demand: d.Proto(tp.PlayerID, targetUUID),
				},
			},
		})
	}

	return es, nil
}

func (c *Controller) handleComplyPropertySetDemand(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_ComplyPropertySetDemand) ([]*monopoly_deal_schema.ServerMessage, error) {
	demandID := monopoly_deal.Identifier(msg.ComplyPropertySetDemand.DemandId)
	sourceUUID, propertySet, err := game.ComplyPropertySetDemand(tp.PlayerID, demandID)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_CompliedDemand{
				CompliedDemand: &monopoly_deal_schema.CompliedDemand{
					SeqNum:   int32(game.SequenceNum),
					DemandId: string(demandID),
					PlayerId: tp.PlayerID.String(),
				},
			},
		},
		{
			Payload: &monopoly_deal_schema.ServerMessage_TransferPropertySet{
				TransferPropertySet: &monopoly_deal_schema.TransferPropertySet{
					SeqNum:      int32(game.SequenceNum),
					SourceId:    sourceUUID.String(),
					TargetId:    tp.PlayerID.String(),
					PropertySet: propertySet.Proto(sourceUUID),
				},
			},
		},
	}, nil
}

func (c *Controller) handlePlayHouse(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayHouse) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardID := monopoly_deal.Identifier(msg.PlayHouse.CardId)
	propertySetID := monopoly_deal.Identifier(msg.PlayHouse.PropertySetId)
	propertySet, card, err := game.PlayHouse(tp.PlayerID, cardID, propertySetID)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_HousePlayed{
				HousePlayed: &monopoly_deal_schema.HousePlayed{
					SeqNum:      int32(game.SequenceNum),
					PlayerId:    tp.PlayerID.String(),
					Card:        card.Proto(),
					PropertySet: propertySet.Proto(tp.PlayerID),
				},
			},
		},
	}, nil
}

func (c *Controller) handlePlayHotel(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayHotel) ([]*monopoly_deal_schema.ServerMessage, error) {
	cardID := monopoly_deal.Identifier(msg.PlayHotel.CardId)
	propertySetID := monopoly_deal.Identifier(msg.PlayHotel.PropertySetId)
	propertySet, card, err := game.PlayHotel(tp.PlayerID, cardID, propertySetID)
	if err != nil {
		return nil, err
	}

	return []*monopoly_deal_schema.ServerMessage{
		{
			Payload: &monopoly_deal_schema.ServerMessage_HotelPlayed{
				HotelPlayed: &monopoly_deal_schema.HotelPlayed{
					SeqNum:      int32(game.SequenceNum),
					PlayerId:    tp.PlayerID.String(),
					Card:        card.Proto(),
					PropertySet: propertySet.Proto(tp.PlayerID),
				},
			},
		},
	}, nil
}
