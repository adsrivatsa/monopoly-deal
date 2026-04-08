package service

import (
	"context"
	"monopoly-deal/internal/errors"
	"monopoly-deal/internal/store"
)

func (c *Controller) CreateRoom(ctx context.Context, args CreateRoomParams) (Room, error) {
	_, err := c.store.GetRoomPlayerByPlayer(ctx, args.PlayerID)
	if err == nil { // checking if player is already part of a room
		return Room{}, errors.EntityAlreadyExists(errors.EntityRoom, err)
	}
	if errors.DBErrorCode(err) != errors.NoDataFound { // if player not part of a room and if some error occurred
		return Room{}, errors.Internal(err)
	}

	r, err := c.store.CreateRoom(ctx, store.CreateRoomParams{
		DisplayName: args.DisplayName,
		Capacity:    args.Capacity,
	})
	if err != nil {
		return Room{}, errors.Internal(err)
	}

	_, err = c.store.CreateRoomPlayer(ctx, store.CreateRoomPlayerParams{
		RoomID:   r.RoomID,
		PlayerID: args.PlayerID,
		IsHost:   true,
	})
	if err != nil {
		return Room{}, errors.Internal(err)
	}

	return r, nil
}
