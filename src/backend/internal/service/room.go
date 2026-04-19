package service

import (
	"context"
	"fun-kames/internal/errors"
	"fun-kames/internal/event"
	"fun-kames/internal/schema"
	"fun-kames/internal/schema/room_schema"
	"fun-kames/internal/store"
	"fun-kames/internal/token"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (c *Controller) ListenRoomEvents(ctx context.Context, tp token.Payload, callback func(message *schema.ServerMessage)) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom, err)
		}
		return err
	}

	ch, err := c.bus.Subscribe(ctx, event.RoomChannelPre+rp.RoomID.String())
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-ch:
			if e.Kind == event.KindServerMessage {
				var msg schema.ServerMessage
				err = proto.Unmarshal(e.Message, &msg)
				if err != nil {
					return err
				}
				callback(&msg)
			}
		}
	}
}

func (c *Controller) GetRoom(ctx context.Context, tp token.Payload) (LongRoom, error) {
	r, err := c.store.GetRoomByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return LongRoom{}, errors.EntityNotFound(errors.EntityRoom, err)
		}
		return LongRoom{}, err
	}

	ps, err := c.store.GetPlayersByRoom(ctx, r.RoomID)
	if err != nil {
		return LongRoom{}, err
	}

	players := make([]ShortPlayer, len(ps))
	for i, p := range ps {
		players[i] = ShortPlayer{
			PlayerID:    p.PlayerID,
			DisplayName: p.DisplayName,
			ImageUrl:    p.ImageUrl,
			IsHost:      p.IsHost,
			IsReady:     p.IsReady,
		}
	}

	return LongRoom{
		RoomID:      r.RoomID,
		DisplayName: r.DisplayName,
		Capacity:    r.Capacity,
		Occupied:    r.Occupied,
		Players:     players,
		Game:        r.Game,
		Settings:    r.Settings,
	}, nil
}

func (c *Controller) ListRooms(ctx context.Context, args ListRoomsParams) (ListRoomsRes, error) {
	var game store.NullGameType
	if args.Game != nil {
		game.GameType = *args.Game
		game.Valid = true
	}

	rps, err := c.store.ListRooms(ctx, store.ListRoomsParams{
		Limit:  args.Limit,
		Offset: args.Offset,
		Search: args.Search,
		Game:   game,
	})
	if err != nil {
		return ListRoomsRes{}, err
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
			Players: []ShortPlayer{
				{
					PlayerID:    rp.PlayerID,
					DisplayName: rp.HostDisplayName,
					ImageUrl:    rp.HostImageUrl,
				},
			},
			Game:     rp.RoomGame,
			Settings: rp.RoomSettings,
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
		return Room{}, err
	}

	r, err := c.store.CreateRoom(ctx, store.CreateRoomParams{
		DisplayName: args.DisplayName,
		Capacity:    args.Capacity,
		Game:        args.Game,
		Settings:    args.Settings,
	})
	if err != nil {
		return Room{}, err
	}

	_, err = c.store.CreateRoomPlayer(ctx, store.CreateRoomPlayerParams{
		RoomID:   r.RoomID,
		PlayerID: tp.PlayerID,
		IsHost:   true,
	})
	if err != nil {
		return Room{}, err
	}

	return Room{
		RoomID:      r.RoomID,
		DisplayName: r.DisplayName,
		Capacity:    r.Capacity,
		Occupied:    r.Occupied,
		Game:        r.Game,
		Settings:    r.Settings,
	}, nil
}

func (c *Controller) JoinRoom(ctx context.Context, tp token.Payload, roomID uuid.UUID) error {
	_, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err == nil { // room exists
		return errors.EntityAlreadyExists(errors.EntityRoomPlayer)
	}
	if errors.DBErrorCode(err) != errors.NoDataFound {
		return err
	}

	p, err := c.store.GetPlayer(ctx, store.GetPlayerParams{
		PlayerID: &tp.PlayerID,
	})
	if err != nil {
		return err
	}

	r, err := c.store.GetRoom(ctx, roomID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom, err)
		}
		return err
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
		return err
	}

	_, err = c.store.IncrementRoomOccupied(ctx, r.RoomID)
	if err != nil {
		return err
	}

	player := &room_schema.Player{
		PlayerId:    p.PlayerID.String(),
		DisplayName: p.DisplayName,
		AvatarUrl:   p.ImageUrl,
		IsReady:     rp.IsReady,
		IsHost:      rp.IsHost,
		JoinedAt:    rp.JoinedAt.UnixMilli(),
	}

	e := &schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: &room_schema.ServerMessage{
				Payload: &room_schema.ServerMessage_PlayerJoinedRoom{
					PlayerJoinedRoom: &room_schema.PlayerJoinedRoom{
						RoomId: roomID.String(),
						Player: player,
					},
				},
			},
		},
	}

	buf, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.RoomChannelPre+roomID.String(), event.NewServerMessageEvent(buf))
}

func (c *Controller) LeaveRoom(ctx context.Context, tp token.Payload) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return err
	}

	var newHostPlayerID *string
	deleteRoom := false
	if rp.IsHost { // host is leaving, find new host
		newHost, err := c.store.GetOldestRoomPlayer(ctx, store.GetOldestRoomPlayerParams{
			RoomID:          rp.RoomID,
			LeavingPlayerID: tp.PlayerID,
		})
		if err != nil {
			if errors.DBErrorCode(err) == errors.NoDataFound {
				// no other player exists to become new host, delete the room
				deleteRoom = true
			} else {
				return err
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
				return err
			}
		}
	}

	err = c.store.DeleteRoomPlayer(ctx, store.DeleteRoomPlayerParams{
		RoomID:   rp.RoomID,
		PlayerID: tp.PlayerID,
	})
	if err != nil {
		return err
	}

	if deleteRoom {
		err = c.store.DeleteRoom(ctx, rp.RoomID)
	} else {
		_, err = c.store.DecrementRoomOccupied(ctx, rp.RoomID)
		if err != nil {
			return err
		}

		e := &schema.ServerMessage{
			Payload: &schema.ServerMessage_RoomMessage{
				RoomMessage: &room_schema.ServerMessage{
					Payload: &room_schema.ServerMessage_PlayerLeftRoom{
						PlayerLeftRoom: &room_schema.PlayerLeftRoom{
							RoomId:          rp.RoomID.String(),
							PlayedId:        tp.PlayerID.String(),
							NewHostPlayerId: newHostPlayerID,
						},
					},
				},
			},
		}

		buf, err := proto.Marshal(e)
		if err != nil {
			return err
		}

		err = c.bus.Publish(ctx, event.RoomChannelPre+rp.RoomID.String(), event.NewServerMessageEvent(buf))
	}

	return err
}

func (c *Controller) ToggleIsReady(ctx context.Context, tp token.Payload) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return err
	}

	rp, err = c.store.ToggleRoomPlayerIsReady(ctx, store.ToggleRoomPlayerIsReadyParams{
		RoomID:   rp.RoomID,
		PlayerID: tp.PlayerID,
	})
	if err != nil {
		return err
	}

	e := &schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: &room_schema.ServerMessage{
				Payload: &room_schema.ServerMessage_PlayerToggledReady{
					PlayerToggledReady: &room_schema.PlayerToggledReady{
						PlayerId: tp.PlayerID.String(),
						IsReady:  rp.IsReady,
					},
				},
			},
		},
	}

	buf, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.RoomChannelPre+rp.RoomID.String(), event.NewServerMessageEvent(buf))
}

func (c *Controller) UpdateRoomSettings(ctx context.Context, tp token.Payload, args UpdateRoomSettingsParams) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return err
	}

	if !rp.IsHost {
		return errors.PlayerIsNotHost
	}

	r, err := c.store.UpdateRoomSettings(ctx, store.UpdateRoomSettingsParams{
		Capacity: args.Capacity,
		Game:     args.Game,
		Settings: args.Settings,
		RoomID:   rp.RoomID,
	})
	if err != nil {
		return err
	}

	e := &schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: &room_schema.ServerMessage{
				Payload: &room_schema.ServerMessage_SettingsUpdated{
					SettingsUpdated: &room_schema.SettingsUpdated{
						Capacity: r.Capacity,
						Game:     r.Game.Proto(),
						Settings: r.Settings,
					},
				},
			},
		},
	}

	buf, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.RoomChannelPre+rp.RoomID.String(), event.NewServerMessageEvent(buf))
}

func (c *Controller) HandleRoomEvent(ctx context.Context, tp token.Payload, msg *schema.ClientMessage_RoomMessage) error {
	switch p := msg.RoomMessage.GetPayload().(type) {
	case *room_schema.ClientMessage_Chat:
		return c.handleRoomChat(ctx, tp, p)
	default:
		return nil
	}
}

func (c *Controller) handleRoomChat(ctx context.Context, tp token.Payload, msg *room_schema.ClientMessage_Chat) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return err
	}

	e := &schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: &room_schema.ServerMessage{
				Payload: &room_schema.ServerMessage_ChatReceived{
					ChatReceived: &room_schema.ChatReceived{
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

	err = c.bus.Publish(ctx, event.RoomChannelPre+rp.RoomID.String(), event.NewServerMessageEvent(buf))
	return err
}
