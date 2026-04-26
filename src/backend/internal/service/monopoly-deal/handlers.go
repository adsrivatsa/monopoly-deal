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

	var action monopoly_deal.Action

	switch p := msg.MonopolyDealMessage.GetPayload().(type) {
	case *monopoly_deal_schema.ClientMessage_PlayMoney:
		action, err = c.handlePlayMoney(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayProperty:
		action, err = c.handlePlayProperty(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayHouse:
		action, err = c.handlePlayHouse(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayHotel:
		action, err = c.handlePlayHotel(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayPassGo:
		action, err = c.handlePlayPassGo(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayItsMyBirthday:
		action, err = c.handlePlayItsMyBirthday(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayDebtCollector:
		action, err = c.handlePlayDebtCollector(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayRent:
		action, err = c.handlePlayRent(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayWildRent:
		action, err = c.handlePlayWildRent(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayDoubleTheRent:
		action, err = c.handleDoubleTheRent(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlaySlyDeal:
		action, err = c.handlerMonopolyDealPlaySlyDeal(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayForcedDeal:
		action, err = c.handlePlayForcedDeal(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_PlayDealBreaker:
		action, err = c.handlePlayDealBreaker(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_ResolvePendingRent:
		action, err = c.handleResolveRent(&game, tp)
	case *monopoly_deal_schema.ClientMessage_ComplyPaymentDemand:
		action, err = c.handleComplyPaymentDemand(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_ComplyPropertyDemand:
		action, err = c.handleComplyPropertyDemand(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_ComplyPropertySetDemand:
		action, err = c.handleComplyPropertySetDemand(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_DenyDemand:
		action, err = c.handleDenyDemand(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_DiscardCards:
		action, err = c.handleDiscardCards(&game, tp, p)
	case *monopoly_deal_schema.ClientMessage_CompleteTurn:
		action, err = c.handleCompleteTurn(&game, tp)
	case *monopoly_deal_schema.ClientMessage_RearrangeCard:
		action, err = c.handleRearrangeCard(&game, tp, p)
	}
	if err != nil {
		return err
	}

	buf, err := game.EncodeMsgpack()
	if err != nil {
		return err
	}

	g, err = c.store.UpdateGameState(ctx, store.UpdateGameStateParams{
		GameState: buf,
		GameID:    g.GameID,
	})
	if err != nil {
		return err
	}

	buf, err = msgpack.Marshal(action)
	if err != nil {
		return err
	}

	_, err = c.store.CreateGameHistory(ctx, store.CreateGameHistoryParams{
		GameID:        g.GameID,
		SeqNum:        int16(action.GetSeqNum()),
		ActionKind:    string(action.GetKind()),
		ActionVersion: int32(action.GetVersion()),
		Action:        buf,
	})
	if err != nil {
		return err
	}

	completeSets, moneyValue, didWin, err := game.CheckWinConditions(tp.PlayerID)
	if err != nil {
		return err
	}

	e := &monopoly_deal_schema.ServerMessage{
		Payload: &monopoly_deal_schema.ServerMessage_Action{
			Action: action.Proto(),
		},
	}

	wrappedEvent := &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: e,
		},
	}

	buf, err = proto.Marshal(wrappedEvent)
	if err != nil {
		return err
	}

	err = c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
	if err != nil {
		return err
	}

	if didWin {
		e = &monopoly_deal_schema.ServerMessage{
			Payload: &monopoly_deal_schema.ServerMessage_WonGame{
				WonGame: &monopoly_deal_schema.WonGame{
					PlayerId:         tp.PlayerID.String(),
					NumCompletedSets: int32(completeSets),
					Money:            int32(moneyValue),
				},
			},
		}

		wrappedEvent = &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: e,
			},
		}

		buf, err = proto.Marshal(wrappedEvent)
		if err != nil {
			return err
		}

		err = c.bus.Publish(ctx, event.GameChannelPre+g.GameID.String(), event.NewMonopolyDealEvent(buf))
		if err != nil {
			return err
		}

		_, err = c.store.CompleteGame(ctx, store.CompleteGameParams{
			Winner: &tp.PlayerID,
			GameID: g.GameID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) handlePlayMoney(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayMoney) (monopoly_deal.Action, error) {
	cardID := monopoly_deal.Identifier(msg.PlayMoney.CardId)
	return game.PlayMoney(tp.PlayerID, cardID)
}

func (c *Controller) handlePlayProperty(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayProperty) (monopoly_deal.Action, error) {
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
	return game.PlayProperty(tp.PlayerID, cardID, propSetID, propSetColor)
}

func (c *Controller) handlePlayHouse(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayHouse) (monopoly_deal.Action, error) {
	cardID := monopoly_deal.Identifier(msg.PlayHouse.CardId)
	propertySetID := monopoly_deal.Identifier(msg.PlayHouse.PropertySetId)
	return game.PlayHouse(tp.PlayerID, cardID, propertySetID)
}

func (c *Controller) handlePlayHotel(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayHotel) (monopoly_deal.Action, error) {
	cardID := monopoly_deal.Identifier(msg.PlayHotel.CardId)
	propertySetID := monopoly_deal.Identifier(msg.PlayHotel.PropertySetId)
	return game.PlayHotel(tp.PlayerID, cardID, propertySetID)
}

func (c *Controller) handlePlayPassGo(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayPassGo) (monopoly_deal.Action, error) {
	cardID := monopoly_deal.Identifier(msg.PlayPassGo.CardId)
	return game.PlayPassGo(tp.PlayerID, cardID)
}

func (c *Controller) handlePlayItsMyBirthday(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayItsMyBirthday) (monopoly_deal.Action, error) {
	cardID := monopoly_deal.Identifier(msg.PlayItsMyBirthday.CardId)
	return game.PlayItsMyBirthday(tp.PlayerID, cardID)
}

func (c *Controller) handlePlayDebtCollector(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayDebtCollector) (monopoly_deal.Action, error) {
	targetUUID, err := uuid.Parse(msg.PlayDebtCollector.TargetId)
	if err != nil {
		return nil, err
	}

	cardID := monopoly_deal.Identifier(msg.PlayDebtCollector.CardId)
	return game.PlayDebtCollector(tp.PlayerID, targetUUID, cardID)
}

func (c *Controller) handlePlayRent(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayRent) (monopoly_deal.Action, error) {
	cardID := monopoly_deal.Identifier(msg.PlayRent.CardId)
	return game.PlayRent(tp.PlayerID, cardID)
}

func (c *Controller) handlePlayWildRent(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayWildRent) (monopoly_deal.Action, error) {
	targetUUID, err := uuid.Parse(msg.PlayWildRent.TargetId)
	if err != nil {
		return nil, err
	}

	cardID := monopoly_deal.Identifier(msg.PlayWildRent.CardId)
	return game.PlayWildRent(tp.PlayerID, targetUUID, cardID)
}

func (c *Controller) handleDoubleTheRent(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayDoubleTheRent) (monopoly_deal.Action, error) {
	cardID := monopoly_deal.Identifier(msg.PlayDoubleTheRent.CardId)
	return game.PlayDoubleTheRent(tp.PlayerID, cardID)
}

func (c *Controller) handlerMonopolyDealPlaySlyDeal(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlaySlyDeal) (monopoly_deal.Action, error) {
	cardID := monopoly_deal.Identifier(msg.PlaySlyDeal.CardId)
	targetID, err := uuid.Parse(msg.PlaySlyDeal.TargetId)
	if err != nil {
		return nil, err
	}
	targetCardID := monopoly_deal.Identifier(msg.PlaySlyDeal.TargetCardId)

	return game.PlaySlyDeal(tp.PlayerID, targetID, cardID, targetCardID)
}

func (c *Controller) handlePlayForcedDeal(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayForcedDeal) (monopoly_deal.Action, error) {
	cardID := monopoly_deal.Identifier(msg.PlayForcedDeal.CardId)
	targetID, err := uuid.Parse(msg.PlayForcedDeal.TargetId)
	if err != nil {
		return nil, err
	}
	sourceCardID := monopoly_deal.Identifier(msg.PlayForcedDeal.SourceCardId)
	targetCardID := monopoly_deal.Identifier(msg.PlayForcedDeal.TargetCardId)

	return game.PlayForcedDeal(tp.PlayerID, targetID, cardID, sourceCardID, targetCardID)
}

func (c *Controller) handlePlayDealBreaker(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_PlayDealBreaker) (monopoly_deal.Action, error) {
	cardID := monopoly_deal.Identifier(msg.PlayDealBreaker.CardId)
	targetUUID, err := uuid.Parse(msg.PlayDealBreaker.TargetId)
	if err != nil {
		return nil, err
	}
	propertySetID := monopoly_deal.Identifier(msg.PlayDealBreaker.PropertySetId)

	return game.PlayDealBreaker(tp.PlayerID, targetUUID, cardID, propertySetID)
}

func (c *Controller) handleResolveRent(game *monopoly_deal.Game, tp token.Payload) (monopoly_deal.Action, error) {
	return game.ResolvePendingRent(tp.PlayerID)
}

func (c *Controller) handleComplyPaymentDemand(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_ComplyPaymentDemand) (monopoly_deal.Action, error) {
	demandID := monopoly_deal.Identifier(msg.ComplyPaymentDemand.DemandId)
	cardIDs := make([]monopoly_deal.Identifier, 0, len(game.Cards))
	for _, cardID := range msg.ComplyPaymentDemand.CardIds {
		cardIDs = append(cardIDs, monopoly_deal.Identifier(cardID))
	}
	return game.ComplyPaymentDemand(tp.PlayerID, demandID, cardIDs...)
}

func (c *Controller) handleComplyPropertyDemand(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_ComplyPropertyDemand) (monopoly_deal.Action, error) {
	demandID := monopoly_deal.Identifier(msg.ComplyPropertyDemand.DemandId)
	return game.ComplyPropertyDemand(tp.PlayerID, demandID)
}

func (c *Controller) handleComplyPropertySetDemand(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_ComplyPropertySetDemand) (monopoly_deal.Action, error) {
	demandID := monopoly_deal.Identifier(msg.ComplyPropertySetDemand.DemandId)
	return game.ComplyPropertySetDemand(tp.PlayerID, demandID)
}

func (c *Controller) handleDenyDemand(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_DenyDemand) (monopoly_deal.Action, error) {
	demandID := monopoly_deal.Identifier(msg.DenyDemand.DemandId)
	cardID := monopoly_deal.Identifier(msg.DenyDemand.CardId)
	return game.DenyDemand(tp.PlayerID, demandID, cardID)
}

func (c *Controller) handleDiscardCards(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_DiscardCards) (monopoly_deal.Action, error) {
	cardIDs := make([]monopoly_deal.Identifier, 0, len(game.Cards))
	for _, cardID := range msg.DiscardCards.CardIds {
		cardIDs = append(cardIDs, monopoly_deal.Identifier(cardID))
	}
	return game.DiscardCards(tp.PlayerID, cardIDs...)
}

func (c *Controller) handleCompleteTurn(game *monopoly_deal.Game, tp token.Payload) (monopoly_deal.Action, error) {
	return game.CompleteTurn(tp.PlayerID)
}

func (c *Controller) handleRearrangeCard(game *monopoly_deal.Game, tp token.Payload, msg *monopoly_deal_schema.ClientMessage_RearrangeCard) (monopoly_deal.Action, error) {
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
	return game.RearrangeProperty(tp.PlayerID, cardID, targetSetID, color)
}
