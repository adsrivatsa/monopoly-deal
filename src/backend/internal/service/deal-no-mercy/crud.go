package deal_no_mercy

import (
	"context"
	stderrors "errors"
	"fmt"
	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"
	"the-deal/internal/errors"
	"the-deal/internal/schema"
	"the-deal/internal/schema/deal_no_mercy_schema"
	"the-deal/internal/store"
	"the-deal/internal/token"
	"time"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

func protoDealNoMercyError(err error) *schema.ServerMessage {
	var intErr errors.Error
	if !stderrors.As(err, &intErr) {
		intErr = errors.Internal(err)
	}

	return &schema.ServerMessage{
		Payload: &schema.ServerMessage_DealNoMercyMessage{
			DealNoMercyMessage: &deal_no_mercy_schema.ServerMessage{
				Payload: &deal_no_mercy_schema.ServerMessage_Error{
					Error: &deal_no_mercy_schema.Error{
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
			return protoDealNoMercyError(errors.EntityNotFound(errors.EntityGame, err))
		}
		return protoDealNoMercyError(err)
	}

	ps, err := c.store.GetPlayersByGame(ctx, g.GameID)
	if err != nil {
		return protoDealNoMercyError(err)
	}

	game, err := deal_no_mercy.DecodeMsgpack(g.GameState)
	if err != nil {
		return protoDealNoMercyError(err)
	}

	playerProtos := make([]*deal_no_mercy_schema.Player, 0, len(ps))
	for _, p := range ps {
		money, _ := game.CountMoney(p.PlayerID)
		completedSets, _ := game.CountCompletedSets(p.PlayerID)
		handLen, _ := game.CountHands(p.PlayerID)
		availableChips, _ := game.CountAvailableDebtChips(p.PlayerID)

		playerProtos = append(playerProtos, &deal_no_mercy_schema.Player{
			PlayerId:           p.PlayerID.String(),
			DisplayName:        p.DisplayName,
			AvatarUrl:          p.ImageUrl,
			Money:              int32(money),
			CompletedSets:      int32(completedSets),
			HandCards:          int32(handLen),
			AvailableDebtChips: int32(availableChips),
		})
	}

	gameState := game.Proto(tp.PlayerID)
	gameState.Players = playerProtos

	gts, err := c.store.ListGameTimeouts(ctx, g.GameID)
	if err != nil {
		return protoDealNoMercyError(err)
	}

	deadlines := make([]*deal_no_mercy_schema.Deadline, 0, len(gts))
	for _, gt := range gts {
		deadlines = append(deadlines, &deal_no_mercy_schema.Deadline{
			DeadlineMs: gt.Deadline.UnixMilli(),
			DemandId:   gt.DemandID,
		})
	}

	gameState.Deadlines = deadlines

	gameState.AssetImages = append(c.smallCardImages(), c.largeCardImages()...)

	return &schema.ServerMessage{
		Payload: &schema.ServerMessage_DealNoMercyMessage{
			DealNoMercyMessage: &deal_no_mercy_schema.ServerMessage{
				Payload: &deal_no_mercy_schema.ServerMessage_GameState{
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
			var action deal_no_mercy.Action
			switch deal_no_mercy.ActionKind(gh.ActionKind) {
			case deal_no_mercy.ActionKindPlayMoney:
				action = new(deal_no_mercy.ActionPlayMoney)
			case deal_no_mercy.ActionKindPlayProperty:
				action = new(deal_no_mercy.ActionPlayProperty)
			case deal_no_mercy.ActionKindPlayShack:
				action = new(deal_no_mercy.ActionPlayShack)
			case deal_no_mercy.ActionKindPlayBigPayday:
				action = new(deal_no_mercy.ActionPlayBigPayday)
			case deal_no_mercy.ActionKindPlayGoAgain:
				action = new(deal_no_mercy.ActionPlayGoAgain)
			case deal_no_mercy.ActionKindDemandCreated:
				action = new(deal_no_mercy.ActionDemandsCreated)
			case deal_no_mercy.ActionKindDemandComplied:
				action = new(deal_no_mercy.ActionDemandComplied)
			case deal_no_mercy.ActionKindDebtSettled:
				action = new(deal_no_mercy.ActionDebtSettled)
			case deal_no_mercy.ActionKindDiscardCards:
				action = new(deal_no_mercy.ActionDiscardCards)
			case deal_no_mercy.ActionKindStartTurn:
				action = new(deal_no_mercy.ActionStartTurn)
			case deal_no_mercy.ActionKindRearrangeCard:
				action = new(deal_no_mercy.ActionRearrangeCard)
			default:
				return fmt.Errorf("unsupported action kind: %s", gh.ActionKind)
			}
			err = msgpack.Unmarshal(gh.Action, action)
			if err != nil {
				return err
			}

			msg := c.MaskEvent(tp, &deal_no_mercy_schema.ServerMessage{
				Payload: &deal_no_mercy_schema.ServerMessage_ActionHistory{
					ActionHistory: &deal_no_mercy_schema.ActionHistory{
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

func (c *Controller) CreateGameFromRoomTx(ctx context.Context, q *store.Queries, roomID uuid.UUID, playerIDs []uuid.UUID) (gameID uuid.UUID, action deal_no_mercy.Action, moveDeadline time.Time, err error) {
	var zeroAction deal_no_mercy.Action

	r, err := q.GetRoom(ctx, roomID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			err = errors.EntityNotFound(errors.EntityRoom)
		}
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	var settings deal_no_mercy.Settings
	err = settings.Decode(r.Settings)
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	game := deal_no_mercy.NewGame(settings, playerIDs)
	currPlayerID := game.Players[game.CurrPlayerIdx]
	action, err = game.StartTurn(currPlayerID)
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	buf, err := game.EncodeMsgpack()
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	gameID, err = uuid.NewV7()
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	g, err := q.CreateGame(ctx, store.CreateGameParams{
		GameID:      gameID,
		DisplayName: r.DisplayName,
		Game:        r.Game,
		GameState:   buf,
	})
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	err = q.CreateGamePlayersFromRoom(ctx, store.CreateGamePlayersFromRoomParams{
		GameID: g.GameID,
		RoomID: r.RoomID,
	})
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	err = q.DeleteRoom(ctx, r.RoomID)
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	err = q.DeleteRoomPlayersByRoom(ctx, r.RoomID)
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	buf, err = msgpack.Marshal(action)
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	_, err = q.CreateGameHistory(ctx, store.CreateGameHistoryParams{
		GameID:        gameID,
		SeqNum:        int16(action.GetSeqNum()),
		ActionKind:    string(action.GetKind()),
		ActionVersion: int32(action.GetVersion()),
		Action:        buf,
	})
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	moveDeadline = time.Now().Add(settings.MoveTimeout)
	err = c.scheduleDefaultMove(ctx, q, gameID, currPlayerID, settings.MoveTimeout, moveDeadline)
	if err != nil {
		return uuid.UUID{}, zeroAction, time.Time{}, err
	}

	return gameID, action, moveDeadline, nil
}

func (c *Controller) CreateGame(ctx context.Context, roomID uuid.UUID, playerIDs []uuid.UUID) (uuid.UUID, error) {
	var gameID uuid.UUID
	var action deal_no_mercy.Action
	var moveDeadline time.Time

	err := c.store.ExecTx(ctx, func(q *store.Queries) error {
		var err error
		gameID, action, moveDeadline, err = c.CreateGameFromRoomTx(ctx, q, roomID, playerIDs)
		return err
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	err = c.PublishAction(ctx, gameID, action, moveDeadline)
	if err != nil {
		return uuid.UUID{}, err
	}

	return gameID, nil
}
