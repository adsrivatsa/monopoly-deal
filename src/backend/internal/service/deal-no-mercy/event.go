package deal_no_mercy

import (
	"context"
	"fmt"
	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"
	"the-deal/internal/errors"
	"the-deal/internal/event"
	"the-deal/internal/schema"
	"the-deal/internal/schema/deal_no_mercy_schema"
	"the-deal/internal/store"
	"the-deal/internal/token"
	"time"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
)

func (c *Controller) PublishAction(ctx context.Context, gameID uuid.UUID, action deal_no_mercy.Action, deadline time.Time) error {
	actionProto := action.Proto()
	actionProto.TurnDeadlineMs = deadline.UnixMilli()

	e := &deal_no_mercy_schema.ServerMessage{
		Payload: &deal_no_mercy_schema.ServerMessage_Action{
			Action: actionProto,
		},
	}

	return c.PublishEvent(ctx, gameID, e)
}

func (c *Controller) PublishEvent(ctx context.Context, gameID uuid.UUID, e *deal_no_mercy_schema.ServerMessage) error {
	s := &schema.ServerMessage{
		Payload: &schema.ServerMessage_DealNoMercyMessage{
			DealNoMercyMessage: e,
		},
	}

	buf, err := proto.Marshal(s)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.GameChannelPre+gameID.String(), event.NewDealNoMercyEvent(buf))
}

func (c *Controller) HandleEvent(ctx context.Context, tp token.Payload, msg *schema.ClientMessage_DealNoMercyMessage) error {
	switch p := msg.DealNoMercyMessage.GetPayload().(type) {
	case *deal_no_mercy_schema.ClientMessage_Chat:
		return c.handleChat(ctx, tp, p)
	default:
		return c.handleGameEvent(ctx, tp, msg)
	}
}

func (c *Controller) handleChat(ctx context.Context, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_Chat) error {
	g, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityGame)
		}
		return err
	}

	e := &deal_no_mercy_schema.ServerMessage{
		Payload: &deal_no_mercy_schema.ServerMessage_ChatReceived{
			ChatReceived: &deal_no_mercy_schema.ChatReceived{
				PlayerId: tp.PlayerID.String(),
				Payload:  msg.Chat.Payload,
			},
		},
	}

	return c.PublishEvent(ctx, g.GameID, e)
}

func (c *Controller) handleGameEvent(ctx context.Context, tp token.Payload, msg *schema.ClientMessage_DealNoMercyMessage) error {
	var gameID uuid.UUID
	var action deal_no_mercy.Action
	var totalSets, totalMoney int
	var didWin bool
	var deadline time.Time

	err := c.store.ExecTx(ctx, func(q *store.Queries) error {
		g, err := q.GetGameByPlayerForUpdate(ctx, tp.PlayerID)
		if err != nil {
			if errors.DBErrorCode(err) == errors.NoDataFound {
				return errors.EntityNotFound(errors.EntityGame)
			}
			return err
		}

		gameID = g.GameID

		game, err := deal_no_mercy.DecodeMsgpack(g.GameState)
		if err != nil {
			return err
		}

		switch p := msg.DealNoMercyMessage.GetPayload().(type) {
		case *deal_no_mercy_schema.ClientMessage_PlayMoney:
			action, err = c.handlePlayMoney(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayProperty:
			action, err = c.handlePlayProperty(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayShack:
			action, err = c.handlePlayShack(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayBigPayday:
			action, err = c.handlePlayBigPayday(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayGoAgain:
			action, err = c.handlePlayGoAgain(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlaySetSnatcher:
			action, err = c.handlePlaySetSnatcher(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayDebtTrap:
			action, err = c.handlePlayDebtTrap(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayHeist:
			action, err = c.handlePlayHeist(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayMarketCrash:
			action, err = c.handlePlayMarketCrash(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayRepoMan:
			action, err = c.handlePlayRepoMan(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayPropertyRaid:
			action, err = c.handlePlayPropertyRaid(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayTaxDay:
			action, err = c.handlePlayTaxDay(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayPickpocket:
			action, err = c.handlePlayPickpocket(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayBankSwap:
			action, err = c.handlePlayBankSwap(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayYoink:
			action, err = c.handlePlayYoink(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_PlayRent:
			action, err = c.handlePlayRent(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_ComplyPaymentDemand:
			action, err = c.handleComplyPaymentDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_ComplyPropertyDemand:
			action, err = c.handleComplyPropertyDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_ComplyPropertySetDemand:
			action, err = c.handleComplyPropertySetDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_ComplyColorPropertiesDemand:
			action, err = c.handleComplyColorPropertiesDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_ComplyBankCardDemand:
			action, err = c.handleComplyBankCardDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_ComplyBankSwapDemand:
			action, err = c.handleComplyBankSwapDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_ComplyDebtTrapDemand:
			action, err = c.handleComplyDebtTrapDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_ComplyRepoManDemand:
			action, err = c.handleComplyRepoManDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_ComplyTaxDayDemand:
			action, err = c.handleComplyTaxDayDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_ComplyPickpocketDemand:
			action, err = c.handleComplyPickpocketDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_DenyDemand:
			action, err = c.handleDenyDemand(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_SettleDebt:
			action, err = c.handleSettleDebt(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_DiscardCards:
			action, err = c.handleDiscardCards(game, tp, p)
		case *deal_no_mercy_schema.ClientMessage_CompleteTurn:
			action, err = c.handleCompleteTurn(game, tp)
		case *deal_no_mercy_schema.ClientMessage_RearrangeCard:
			action, err = c.handleRearrangeCard(game, tp, p)
		default:
			return errors.InvalidMessageType[schema.ClientMessage_DealNoMercyMessage]()
		}
		if err != nil {
			return err
		}

		buf, err := game.EncodeMsgpack()
		if err != nil {
			return err
		}

		g, err = q.UpdateGameState(ctx, store.UpdateGameStateParams{
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

		_, err = q.CreateGameHistory(ctx, store.CreateGameHistoryParams{
			GameID:        g.GameID,
			SeqNum:        int16(action.GetSeqNum()),
			ActionKind:    string(action.GetKind()),
			ActionVersion: int32(action.GetVersion()),
			Action:        buf,
		})
		if err != nil {
			return err
		}

		totalSets, totalMoney, didWin, err = game.CheckWinConditions(tp.PlayerID)
		if err != nil {
			return err
		}

		if didWin {
			_, err = q.DeleteGameTimeout(ctx, store.DeleteGameTimeoutParams{
				GameID:   gameID,
				PlayerID: tp.PlayerID,
				DemandID: nil,
			})
			if err != nil && errors.DBErrorCode(err) != errors.NoDataFound {
				return err
			}

			_, err = q.CompleteGame(ctx, store.CompleteGameParams{
				Winner: &tp.PlayerID,
				GameID: gameID,
			})
			if err != nil {
				return err
			}

			return nil
		}

		deadline, err = c.scheduleTimeouts(ctx, q, game, gameID, tp, msg, action)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = c.PublishAction(ctx, gameID, action, deadline)
	if err != nil {
		return err
	}

	if didWin {
		e := &deal_no_mercy_schema.ServerMessage{
			Payload: &deal_no_mercy_schema.ServerMessage_WonGame{
				WonGame: &deal_no_mercy_schema.WonGame{
					PlayerId:         tp.PlayerID.String(),
					NumCompletedSets: int32(totalSets),
					Money:            int32(totalMoney),
				},
			},
		}

		err = c.PublishEvent(ctx, gameID, e)
		if err != nil {
			return err
		}
	}

	return nil
}

// clearMoveTimeoutIfIdle removes the current player's move timeout and
// reschedules it once every demand raised by the last aggressive play has
// been resolved, so the current player's turn clock resumes.
func (c *Controller) clearMoveTimeoutIfIdle(ctx context.Context, q *store.Queries, game *deal_no_mercy.Game, gameID uuid.UUID) (time.Time, error) {
	var deadline time.Time
	if len(game.Demands) != 0 {
		return deadline, nil
	}

	currPlayerID := game.Players[game.CurrPlayerIdx]
	_, err := q.DeleteGameTimeout(ctx, store.DeleteGameTimeoutParams{
		GameID:   gameID,
		PlayerID: currPlayerID,
		DemandID: nil,
	})
	if err != nil && errors.DBErrorCode(err) != errors.NoDataFound {
		return deadline, err
	}

	deadline = time.Now().Add(game.Config.MoveTimeout)
	err = c.scheduleDefaultMove(ctx, q, gameID, currPlayerID, game.Config.MoveTimeout, deadline)
	return deadline, err
}

// scheduleDemandTimeouts registers a demand timeout for every demand carried
// by a just-created ActionDemandsCreated (a play or a NAH! re-issue).
func (c *Controller) scheduleDemandTimeouts(ctx context.Context, q *store.Queries, game *deal_no_mercy.Game, gameID uuid.UUID, action deal_no_mercy.Action) (time.Time, error) {
	dcAct, ok := action.(*deal_no_mercy.ActionDemandsCreated)
	if !ok {
		return time.Time{}, fmt.Errorf("deal_no_mercy: expected ActionDemandsCreated")
	}

	deadline := time.Now().Add(game.Config.DemandTimeout)
	for _, demand := range dcAct.Demands {
		demandID := string(demand.ID)
		err := c.scheduleDefaultDemand(ctx, q, gameID, demand.TargetID, demandID, game.Config.DemandTimeout, deadline)
		if err != nil {
			return time.Time{}, err
		}
	}
	return deadline, nil
}

// scheduleTimeouts advances the timeout bookkeeping after a play/comply has
// been applied and persisted. It returns the deadline stamped on the action
// published to clients.
func (c *Controller) scheduleTimeouts(ctx context.Context, q *store.Queries, game *deal_no_mercy.Game, gameID uuid.UUID, tp token.Payload, msg *schema.ClientMessage_DealNoMercyMessage, action deal_no_mercy.Action) (time.Time, error) {
	switch msg.DealNoMercyMessage.GetPayload().(type) {

	// Plays that consume a move but raise no demand: keep the current
	// player's move clock running.
	case *deal_no_mercy_schema.ClientMessage_PlayMoney, *deal_no_mercy_schema.ClientMessage_PlayProperty,
		*deal_no_mercy_schema.ClientMessage_PlayShack, *deal_no_mercy_schema.ClientMessage_PlayBigPayday,
		*deal_no_mercy_schema.ClientMessage_PlayGoAgain, *deal_no_mercy_schema.ClientMessage_SettleDebt,
		*deal_no_mercy_schema.ClientMessage_DiscardCards, *deal_no_mercy_schema.ClientMessage_RearrangeCard:
		deadline := time.Now().Add(game.Config.MoveTimeout)
		err := c.scheduleDefaultMove(ctx, q, gameID, tp.PlayerID, game.Config.MoveTimeout, deadline)
		return deadline, err

	// Aggressive plays: clear the current player's move clock (it resumes
	// when demands resolve) and start a demand clock per target.
	case *deal_no_mercy_schema.ClientMessage_PlaySetSnatcher, *deal_no_mercy_schema.ClientMessage_PlayDebtTrap,
		*deal_no_mercy_schema.ClientMessage_PlayHeist, *deal_no_mercy_schema.ClientMessage_PlayMarketCrash,
		*deal_no_mercy_schema.ClientMessage_PlayRepoMan, *deal_no_mercy_schema.ClientMessage_PlayPropertyRaid,
		*deal_no_mercy_schema.ClientMessage_PlayTaxDay, *deal_no_mercy_schema.ClientMessage_PlayPickpocket,
		*deal_no_mercy_schema.ClientMessage_PlayBankSwap, *deal_no_mercy_schema.ClientMessage_PlayYoink,
		*deal_no_mercy_schema.ClientMessage_PlayRent:
		_, err := q.DeleteGameTimeout(ctx, store.DeleteGameTimeoutParams{
			GameID:   gameID,
			PlayerID: tp.PlayerID,
			DemandID: nil,
		})
		if err != nil && errors.DBErrorCode(err) != errors.NoDataFound {
			return time.Time{}, err
		}
		return c.scheduleDemandTimeouts(ctx, q, game, gameID, action)

	// NAH!: the denied demand is replaced by a fresh (swapped) demand; clear
	// the old demand timeout and register the new one(s).
	case *deal_no_mercy_schema.ClientMessage_DenyDemand:
		demandID := msg.DealNoMercyMessage.GetDenyDemand().DemandId
		_, err := q.DeleteGameTimeout(ctx, store.DeleteGameTimeoutParams{
			GameID:   gameID,
			PlayerID: tp.PlayerID,
			DemandID: &demandID,
		})
		if err != nil && errors.DBErrorCode(err) != errors.NoDataFound {
			return time.Time{}, err
		}
		return c.scheduleDemandTimeouts(ctx, q, game, gameID, action)

	// Compliances: clear this demand's timeout, then resume the current
	// player's move clock once no demands remain.
	case *deal_no_mercy_schema.ClientMessage_ComplyPaymentDemand:
		return c.clearDemandAndResume(ctx, q, game, gameID, tp, msg.DealNoMercyMessage.GetComplyPaymentDemand().DemandId)
	case *deal_no_mercy_schema.ClientMessage_ComplyPropertyDemand:
		return c.clearDemandAndResume(ctx, q, game, gameID, tp, msg.DealNoMercyMessage.GetComplyPropertyDemand().DemandId)
	case *deal_no_mercy_schema.ClientMessage_ComplyPropertySetDemand:
		return c.clearDemandAndResume(ctx, q, game, gameID, tp, msg.DealNoMercyMessage.GetComplyPropertySetDemand().DemandId)
	case *deal_no_mercy_schema.ClientMessage_ComplyColorPropertiesDemand:
		return c.clearDemandAndResume(ctx, q, game, gameID, tp, msg.DealNoMercyMessage.GetComplyColorPropertiesDemand().DemandId)
	case *deal_no_mercy_schema.ClientMessage_ComplyBankCardDemand:
		return c.clearDemandAndResume(ctx, q, game, gameID, tp, msg.DealNoMercyMessage.GetComplyBankCardDemand().DemandId)
	case *deal_no_mercy_schema.ClientMessage_ComplyBankSwapDemand:
		return c.clearDemandAndResume(ctx, q, game, gameID, tp, msg.DealNoMercyMessage.GetComplyBankSwapDemand().DemandId)
	case *deal_no_mercy_schema.ClientMessage_ComplyDebtTrapDemand:
		return c.clearDemandAndResume(ctx, q, game, gameID, tp, msg.DealNoMercyMessage.GetComplyDebtTrapDemand().DemandId)
	case *deal_no_mercy_schema.ClientMessage_ComplyRepoManDemand:
		return c.clearDemandAndResume(ctx, q, game, gameID, tp, msg.DealNoMercyMessage.GetComplyRepoManDemand().DemandId)
	case *deal_no_mercy_schema.ClientMessage_ComplyTaxDayDemand:
		return c.clearDemandAndResume(ctx, q, game, gameID, tp, msg.DealNoMercyMessage.GetComplyTaxDayDemand().DemandId)
	case *deal_no_mercy_schema.ClientMessage_ComplyPickpocketDemand:
		return c.clearDemandAndResume(ctx, q, game, gameID, tp, msg.DealNoMercyMessage.GetComplyPickpocketDemand().DemandId)

	// Turn end: clear this player's move clock, start the next player's.
	case *deal_no_mercy_schema.ClientMessage_CompleteTurn:
		_, err := q.DeleteGameTimeout(ctx, store.DeleteGameTimeoutParams{
			GameID:   gameID,
			PlayerID: tp.PlayerID,
			DemandID: nil,
		})
		if err != nil && errors.DBErrorCode(err) != errors.NoDataFound {
			return time.Time{}, err
		}

		nextPlayerID := game.Players[game.CurrPlayerIdx]
		deadline := time.Now().Add(game.Config.MoveTimeout)
		err = c.scheduleDefaultMove(ctx, q, gameID, nextPlayerID, game.Config.MoveTimeout, deadline)
		return deadline, err
	}

	return time.Time{}, nil
}

// clearDemandAndResume deletes the resolved demand's timeout and, if that was
// the last outstanding demand, restarts the current player's move clock.
func (c *Controller) clearDemandAndResume(ctx context.Context, q *store.Queries, game *deal_no_mercy.Game, gameID uuid.UUID, tp token.Payload, demandID string) (time.Time, error) {
	_, err := q.DeleteGameTimeout(ctx, store.DeleteGameTimeoutParams{
		GameID:   gameID,
		PlayerID: tp.PlayerID,
		DemandID: &demandID,
	})
	if err != nil && errors.DBErrorCode(err) != errors.NoDataFound {
		return time.Time{}, err
	}

	return c.clearMoveTimeoutIfIdle(ctx, q, game, gameID)
}

func (c *Controller) handlePlayMoney(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayMoney) (deal_no_mercy.Action, error) {
	cardID := deal_no_mercy.Identifier(msg.PlayMoney.CardId)
	return game.PlayMoney(tp.PlayerID, cardID)
}

func (c *Controller) handlePlayProperty(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayProperty) (deal_no_mercy.Action, error) {
	cardID := deal_no_mercy.Identifier(msg.PlayProperty.CardId)
	var propSetID *deal_no_mercy.Identifier
	if msg.PlayProperty.PropertySetId != nil {
		id := deal_no_mercy.Identifier(*msg.PlayProperty.PropertySetId)
		propSetID = &id
	}
	var propSetColor *deal_no_mercy.Color
	if msg.PlayProperty.ActiveColor != nil {
		color := deal_no_mercy.ColorFromProto(*msg.PlayProperty.ActiveColor)
		propSetColor = &color
	}
	return game.PlayProperty(tp.PlayerID, cardID, propSetID, propSetColor)
}

func (c *Controller) handlePlayShack(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayShack) (deal_no_mercy.Action, error) {
	cardID := deal_no_mercy.Identifier(msg.PlayShack.CardId)
	propSetID := deal_no_mercy.Identifier(msg.PlayShack.PropertySetId)
	return game.PlayShack(tp.PlayerID, cardID, propSetID)
}

func (c *Controller) handlePlayBigPayday(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayBigPayday) (deal_no_mercy.Action, error) {
	cardID := deal_no_mercy.Identifier(msg.PlayBigPayday.CardId)
	return game.PlayBigPayday(tp.PlayerID, cardID)
}

func (c *Controller) handlePlayGoAgain(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayGoAgain) (deal_no_mercy.Action, error) {
	cardID := deal_no_mercy.Identifier(msg.PlayGoAgain.CardId)
	return game.PlayGoAgain(tp.PlayerID, cardID)
}

func (c *Controller) handlePlaySetSnatcher(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlaySetSnatcher) (deal_no_mercy.Action, error) {
	targetID, err := uuid.Parse(msg.PlaySetSnatcher.TargetId)
	if err != nil {
		return nil, errors.InvalidUUID(err)
	}
	cardID := deal_no_mercy.Identifier(msg.PlaySetSnatcher.CardId)
	setID := deal_no_mercy.Identifier(msg.PlaySetSnatcher.PropertySetId)
	return game.PlaySetSnatcher(tp.PlayerID, targetID, cardID, setID)
}

func (c *Controller) handlePlayDebtTrap(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayDebtTrap) (deal_no_mercy.Action, error) {
	targetID, err := uuid.Parse(msg.PlayDebtTrap.TargetId)
	if err != nil {
		return nil, errors.InvalidUUID(err)
	}
	cardID := deal_no_mercy.Identifier(msg.PlayDebtTrap.CardId)
	return game.PlayDebtTrap(tp.PlayerID, targetID, cardID)
}

func (c *Controller) handlePlayHeist(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayHeist) (deal_no_mercy.Action, error) {
	cardID := deal_no_mercy.Identifier(msg.PlayHeist.CardId)
	picks, err := parsePicks(msg.PlayHeist.Picks)
	if err != nil {
		return nil, err
	}
	return game.PlayHeist(tp.PlayerID, cardID, picks)
}

func (c *Controller) handlePlayMarketCrash(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayMarketCrash) (deal_no_mercy.Action, error) {
	cardID := deal_no_mercy.Identifier(msg.PlayMarketCrash.CardId)
	picks, err := parsePicks(msg.PlayMarketCrash.Picks)
	if err != nil {
		return nil, err
	}
	return game.PlayMarketCrash(tp.PlayerID, cardID, picks)
}

func (c *Controller) handlePlayRepoMan(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayRepoMan) (deal_no_mercy.Action, error) {
	targetID, err := uuid.Parse(msg.PlayRepoMan.TargetId)
	if err != nil {
		return nil, errors.InvalidUUID(err)
	}
	cardID := deal_no_mercy.Identifier(msg.PlayRepoMan.CardId)
	return game.PlayRepoMan(tp.PlayerID, targetID, cardID)
}

func (c *Controller) handlePlayPropertyRaid(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayPropertyRaid) (deal_no_mercy.Action, error) {
	cardID := deal_no_mercy.Identifier(msg.PlayPropertyRaid.CardId)
	color := deal_no_mercy.ColorFromProto(msg.PlayPropertyRaid.Color)
	return game.PlayPropertyRaid(tp.PlayerID, cardID, color)
}

func (c *Controller) handlePlayTaxDay(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayTaxDay) (deal_no_mercy.Action, error) {
	targetID, err := uuid.Parse(msg.PlayTaxDay.TargetId)
	if err != nil {
		return nil, errors.InvalidUUID(err)
	}
	cardID := deal_no_mercy.Identifier(msg.PlayTaxDay.CardId)
	return game.PlayTaxDay(tp.PlayerID, targetID, cardID)
}

func (c *Controller) handlePlayPickpocket(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayPickpocket) (deal_no_mercy.Action, error) {
	targetID, err := uuid.Parse(msg.PlayPickpocket.TargetId)
	if err != nil {
		return nil, errors.InvalidUUID(err)
	}
	cardID := deal_no_mercy.Identifier(msg.PlayPickpocket.CardId)
	category := deal_no_mercy.StealCategoryFromProto(msg.PlayPickpocket.Category)
	return game.PlayPickpocket(tp.PlayerID, targetID, cardID, category)
}

func (c *Controller) handlePlayBankSwap(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayBankSwap) (deal_no_mercy.Action, error) {
	targetID, err := uuid.Parse(msg.PlayBankSwap.TargetId)
	if err != nil {
		return nil, errors.InvalidUUID(err)
	}
	cardID := deal_no_mercy.Identifier(msg.PlayBankSwap.CardId)
	return game.PlayBankSwap(tp.PlayerID, targetID, cardID)
}

func (c *Controller) handlePlayYoink(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayYoink) (deal_no_mercy.Action, error) {
	targetID, err := uuid.Parse(msg.PlayYoink.TargetId)
	if err != nil {
		return nil, errors.InvalidUUID(err)
	}
	cardID := deal_no_mercy.Identifier(msg.PlayYoink.CardId)
	return game.PlayYoink(tp.PlayerID, targetID, cardID)
}

func (c *Controller) handlePlayRent(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_PlayRent) (deal_no_mercy.Action, error) {
	cardID := deal_no_mercy.Identifier(msg.PlayRent.CardId)
	color := deal_no_mercy.ColorFromProto(msg.PlayRent.Color)
	return game.PlayRent(tp.PlayerID, cardID, color)
}

func (c *Controller) handleComplyPaymentDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_ComplyPaymentDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.ComplyPaymentDemand.DemandId)
	cardIDs := make([]deal_no_mercy.Identifier, 0, len(msg.ComplyPaymentDemand.CardIds))
	for _, cardID := range msg.ComplyPaymentDemand.CardIds {
		cardIDs = append(cardIDs, deal_no_mercy.Identifier(cardID))
	}
	return game.ComplyPaymentDemand(tp.PlayerID, demandID, cardIDs...)
}

func (c *Controller) handleComplyPropertyDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_ComplyPropertyDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.ComplyPropertyDemand.DemandId)
	return game.ComplyPropertyDemand(tp.PlayerID, demandID)
}

func (c *Controller) handleComplyPropertySetDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_ComplyPropertySetDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.ComplyPropertySetDemand.DemandId)
	return game.ComplyPropertySetDemand(tp.PlayerID, demandID)
}

func (c *Controller) handleComplyColorPropertiesDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_ComplyColorPropertiesDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.ComplyColorPropertiesDemand.DemandId)
	return game.ComplyColorPropertiesDemand(tp.PlayerID, demandID)
}

func (c *Controller) handleComplyBankCardDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_ComplyBankCardDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.ComplyBankCardDemand.DemandId)
	return game.ComplyBankCardDemand(tp.PlayerID, demandID)
}

func (c *Controller) handleComplyBankSwapDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_ComplyBankSwapDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.ComplyBankSwapDemand.DemandId)
	return game.ComplyBankSwapDemand(tp.PlayerID, demandID)
}

func (c *Controller) handleComplyDebtTrapDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_ComplyDebtTrapDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.ComplyDebtTrapDemand.DemandId)
	return game.ComplyDebtTrapDemand(tp.PlayerID, demandID)
}

func (c *Controller) handleComplyRepoManDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_ComplyRepoManDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.ComplyRepoManDemand.DemandId)
	keepCardID := deal_no_mercy.Identifier(msg.ComplyRepoManDemand.KeepCardId)
	distribution, err := parseDistribution(msg.ComplyRepoManDemand.Distribution)
	if err != nil {
		return nil, err
	}
	return game.ComplyRepoManDemand(tp.PlayerID, demandID, keepCardID, distribution)
}

func (c *Controller) handleComplyTaxDayDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_ComplyTaxDayDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.ComplyTaxDayDemand.DemandId)
	keepCardID := deal_no_mercy.Identifier(msg.ComplyTaxDayDemand.KeepCardId)
	distribution, err := parseDistribution(msg.ComplyTaxDayDemand.Distribution)
	if err != nil {
		return nil, err
	}
	return game.ComplyTaxDayDemand(tp.PlayerID, demandID, keepCardID, distribution)
}

func (c *Controller) handleComplyPickpocketDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_ComplyPickpocketDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.ComplyPickpocketDemand.DemandId)
	return game.ComplyPickpocketDemand(tp.PlayerID, demandID)
}

func (c *Controller) handleDenyDemand(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_DenyDemand) (deal_no_mercy.Action, error) {
	demandID := deal_no_mercy.Identifier(msg.DenyDemand.DemandId)
	cardID := deal_no_mercy.Identifier(msg.DenyDemand.CardId)
	return game.DenyDemand(tp.PlayerID, demandID, cardID)
}

func (c *Controller) handleSettleDebt(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_SettleDebt) (deal_no_mercy.Action, error) {
	debtID := deal_no_mercy.Identifier(msg.SettleDebt.DebtId)
	cardID := deal_no_mercy.Identifier(msg.SettleDebt.CardId)
	return game.SettleDebt(tp.PlayerID, debtID, cardID)
}

func (c *Controller) handleDiscardCards(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_DiscardCards) (deal_no_mercy.Action, error) {
	cardIDs := make([]deal_no_mercy.Identifier, 0, len(msg.DiscardCards.CardIds))
	for _, cardID := range msg.DiscardCards.CardIds {
		cardIDs = append(cardIDs, deal_no_mercy.Identifier(cardID))
	}
	return game.DiscardCards(tp.PlayerID, cardIDs...)
}

func (c *Controller) handleCompleteTurn(game *deal_no_mercy.Game, tp token.Payload) (deal_no_mercy.Action, error) {
	return game.CompleteTurn(tp.PlayerID)
}

func (c *Controller) handleRearrangeCard(game *deal_no_mercy.Game, tp token.Payload, msg *deal_no_mercy_schema.ClientMessage_RearrangeCard) (deal_no_mercy.Action, error) {
	cardID := deal_no_mercy.Identifier(msg.RearrangeCard.CardId)
	var targetSetID *deal_no_mercy.Identifier
	if msg.RearrangeCard.PropertySetId != nil {
		id := deal_no_mercy.Identifier(*msg.RearrangeCard.PropertySetId)
		targetSetID = &id
	}
	var color *deal_no_mercy.Color
	if msg.RearrangeCard.Color != nil {
		col := deal_no_mercy.ColorFromProto(*msg.RearrangeCard.Color)
		color = &col
	}
	return game.RearrangeProperty(tp.PlayerID, cardID, targetSetID, color)
}

func parsePicks(picks []*deal_no_mercy_schema.PropertyPick) (map[uuid.UUID]deal_no_mercy.Identifier, error) {
	out := make(map[uuid.UUID]deal_no_mercy.Identifier, len(picks))
	for _, pick := range picks {
		targetID, err := uuid.Parse(pick.TargetId)
		if err != nil {
			return nil, errors.InvalidUUID(err)
		}
		out[targetID] = deal_no_mercy.Identifier(pick.CardId)
	}
	return out, nil
}

func parseDistribution(assignments []*deal_no_mercy_schema.DistributionAssignment) (map[deal_no_mercy.Identifier]uuid.UUID, error) {
	out := make(map[deal_no_mercy.Identifier]uuid.UUID, len(assignments))
	for _, assignment := range assignments {
		recipientID, err := uuid.Parse(assignment.RecipientId)
		if err != nil {
			return nil, errors.InvalidUUID(err)
		}
		out[deal_no_mercy.Identifier(assignment.CardId)] = recipientID
	}
	return out, nil
}
