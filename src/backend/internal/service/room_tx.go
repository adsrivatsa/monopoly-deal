package service

import (
	"context"

	"the-deal/internal/errors"
	"the-deal/internal/store"

	"github.com/google/uuid"
)

// quickPlayAdvisoryLockClass namespaces quick-play bucket locking in PostgreSQL advisory locks.
const quickPlayAdvisoryLockClass int32 = 712889

func quickPlayGameLockKey(game store.GameType) (int32, error) {
	switch game {
	case store.GameTypeMonopolyDeal:
		return 1, nil
	case store.GameTypeDealNoMercy:
		return 2, nil
	default:
		return 0, errors.GameNotSupported
	}
}

type joinRoomTxParams struct {
	RoomID   uuid.UUID
	PlayerID uuid.UUID
	IsHost   bool
	IsReady  bool
}

func (c *Controller) joinRoomTx(ctx context.Context, q *store.Queries, p joinRoomTxParams) (store.Room, store.RoomPlayer, error) {
	r, err := q.GetRoomForUpdate(ctx, p.RoomID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return store.Room{}, store.RoomPlayer{}, errors.EntityNotFound(errors.EntityRoom, err)
		}
		return store.Room{}, store.RoomPlayer{}, err
	}

	if r.Occupied >= r.Capacity {
		return store.Room{}, store.RoomPlayer{}, errors.RoomIsFull
	}

	rp, err := q.CreateRoomPlayer(ctx, store.CreateRoomPlayerParams{
		RoomID:   p.RoomID,
		PlayerID: p.PlayerID,
		IsHost:   p.IsHost,
		IsReady:  p.IsReady,
	})
	if err != nil {
		if errors.DBErrorCode(err) == errors.ForeignKeyViolation {
			return store.Room{}, store.RoomPlayer{}, errors.EntityNotFound(errors.EntityRoom, err)
		}
		return store.Room{}, store.RoomPlayer{}, err
	}

	r, err = q.IncrementRoomOccupied(ctx, r.RoomID)
	if err != nil {
		return store.Room{}, store.RoomPlayer{}, err
	}

	return r, rp, nil
}

type createRoomTxParams struct {
	DisplayName string
	Capacity    int32
	Game        store.GameType
	Settings    []byte
	IsPrivate   bool
	IsQuickPlay bool
	PlayerID    uuid.UUID
	IsHost      bool
	IsReady     bool
}

func (c *Controller) createRoomTx(ctx context.Context, q *store.Queries, p createRoomTxParams) (store.Room, store.RoomPlayer, error) {
	roomID, err := uuid.NewV7()
	if err != nil {
		return store.Room{}, store.RoomPlayer{}, err
	}

	r, err := q.CreateRoom(ctx, store.CreateRoomParams{
		RoomID:      roomID,
		DisplayName: p.DisplayName,
		Capacity:    p.Capacity,
		Game:        p.Game,
		Settings:    p.Settings,
		IsPrivate:   p.IsPrivate,
		IsQuickPlay: p.IsQuickPlay,
	})
	if err != nil {
		return store.Room{}, store.RoomPlayer{}, err
	}

	rp, err := q.CreateRoomPlayer(ctx, store.CreateRoomPlayerParams{
		RoomID:   r.RoomID,
		PlayerID: p.PlayerID,
		IsHost:   p.IsHost,
		IsReady:  p.IsReady,
	})
	if err != nil {
		return store.Room{}, store.RoomPlayer{}, err
	}

	return r, rp, nil
}
