package service

import (
	"context"
	"monopoly-deal/internal/errors"
	"monopoly-deal/internal/event"
	"monopoly-deal/internal/schema"
	"monopoly-deal/internal/store"
	"monopoly-deal/internal/token"

	"github.com/google/uuid"
)

func (c *Controller) ListenRoomEvents(ctx context.Context, tp token.Payload, callback func(message *schema.ServerMessage)) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom, err)
		}
		return errors.Internal(err)
	}

	ch, err := c.bus.Subscribe(ctx, event.RoomChannelPre+rp.RoomID.String())
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-ch:
			callback(msg)
		}
	}
}

func (c *Controller) ListRooms(ctx context.Context, args ListRoomsParams) (ListRoomsRes, error) {
	rps, err := c.store.ListRoomPlayers(ctx, args)
	if err != nil {
		return ListRoomsRes{}, errors.Internal(err)
	}

	res := ListRoomsRes{}

	rooms := make([]LongRoom, len(rps))
	for i, rp := range rps {
		res.TotalCount = rp.TotalCount
		rooms[i] = LongRoom{
			RoomID:      rp.RoomID,
			DisplayName: rp.RoomDisplayName,
			Capacity:    rp.RoomCapacity,
			Occupied:    rp.RoomOccupied,
			Host: ShortPlayer{
				PlayerID:    rp.PlayerID,
				DisplayName: rp.HostDisplayName,
				ImageUrl:    rp.HostImageUrl,
			},
		}
	}

	res.Rooms = rooms
	return res, nil
}

func (c *Controller) CreateRoom(ctx context.Context, tp token.Payload, args CreateRoomParams) (Room, error) {
	_, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err == nil { // room exists
		return Room{}, errors.EntityAlreadyExists(errors.EntityRoom)
	}
	if errors.DBErrorCode(err) != errors.NoDataFound {
		return Room{}, errors.Internal(err)
	}

	_, err = c.store.GetPlayer(ctx, store.GetPlayerParams{
		PlayerID: &tp.PlayerID,
	})
	if err != nil {
		return Room{}, errors.Internal(err)
	}

	r, err := c.store.CreateRoom(ctx, args)
	if err != nil {
		return Room{}, errors.Internal(err)
	}

	_, err = c.store.CreateRoomPlayer(ctx, store.CreateRoomPlayerParams{
		RoomID:   r.RoomID,
		PlayerID: tp.PlayerID,
		IsHost:   true,
	})
	if err != nil {
		return Room{}, errors.Internal(err)
	}

	return r, nil
}

func (c *Controller) JoinRoom(ctx context.Context, tp token.Payload, roomID uuid.UUID) error {
	_, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err == nil { // room exists
		return errors.EntityAlreadyExists(errors.EntityRoomPlayer)
	}
	if errors.DBErrorCode(err) != errors.NoDataFound {
		return errors.Internal(err)
	}

	p, err := c.store.GetPlayer(ctx, store.GetPlayerParams{
		PlayerID: &tp.PlayerID,
	})
	if err != nil {
		return errors.Internal(err)
	}

	r, err := c.store.GetRoom(ctx, roomID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom, err)
		}
		return errors.Internal(err)
	}

	if r.Occupied >= r.Capacity {
		return errors.RoomIsFull
	}

	rp, err := c.store.CreateRoomPlayer(ctx, store.CreateRoomPlayerParams{
		RoomID:   roomID,
		PlayerID: tp.PlayerID,
		IsHost:   false,
	})
	if err != nil {
		if errors.DBErrorCode(err) == errors.ForeignKeyViolation {
			return errors.EntityNotFound(errors.EntityRoom, err)
		}
		return errors.Internal(err)
	}

	_, err = c.store.IncrementRoomOccupied(ctx, r.RoomID)
	if err != nil {
		return errors.Internal(err)
	}

	player := &schema.Player{
		PlayerId:    p.PlayerID.String(),
		DisplayName: p.DisplayName,
		AvatarUrl:   p.ImageUrl,
		IsReady:     rp.IsReady,
		IsHost:      rp.IsHost,
		JoinedAt:    rp.JoinedAt.UnixMilli(),
	}

	e := &schema.PlayerJoinedRoom{
		RoomId: roomID.String(),
		Player: player,
	}

	return c.bus.Publish(ctx, event.RoomChannelPre+roomID.String(), &schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: &schema.ServerRoomMessage{
				Payload: &schema.ServerRoomMessage_PlayerJoinedRoom{
					PlayerJoinedRoom: e,
				},
			},
		},
	})
}

func (c *Controller) LeaveRoom(ctx context.Context, tp token.Payload) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return errors.Internal(err)
	}

	var newHostPlayerID *string
	deleteRoom := false
	if rp.IsHost { // host is leaving, find new host
		newHost, err := c.store.GetOldestRoomPlayer(ctx, rp.RoomID)
		if err != nil {
			if errors.DBErrorCode(err) == errors.NoDataFound {
				// no other player exists to become new host, delete the room
				deleteRoom = true
			} else {
				return errors.Internal(err)
			}
		}

		if !deleteRoom { // set the new host
			playerID := newHost.PlayerID.String()
			newHostPlayerID = &playerID

			_, err := c.store.UpdateRoomPlayerHost(ctx, store.UpdateRoomPlayerHostParams{
				IsHost:   true,
				RoomID:   rp.RoomID,
				PlayerID: newHost.PlayerID,
			})
			if err != nil {
				return errors.Internal(err)
			}
		}
	}

	err = c.store.DeleteRoomPlayer(ctx, store.DeleteRoomPlayerParams{
		RoomID:   rp.RoomID,
		PlayerID: tp.PlayerID,
	})
	if err != nil {
		return errors.Internal(err)
	}

	if deleteRoom {
		err = c.store.DeleteRoom(ctx, rp.RoomID)
	} else {
		_, err = c.store.DecrementRoomOccupied(ctx, rp.RoomID)
		if err != nil {
			return errors.Internal(err)
		}

		e := &schema.PlayerLeftRoom{
			RoomId:          rp.RoomID.String(),
			PlayedId:        tp.PlayerID.String(),
			NewHostPlayerId: newHostPlayerID,
		}

		err = c.bus.Publish(ctx, event.RoomChannelPre+rp.RoomID.String(), &schema.ServerMessage{
			Payload: &schema.ServerMessage_RoomMessage{
				RoomMessage: &schema.ServerRoomMessage{
					Payload: &schema.ServerRoomMessage_PlayerLeftRoom{
						PlayerLeftRoom: e,
					},
				},
			},
		})
	}

	return err
}
