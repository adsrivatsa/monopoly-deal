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

func (c *Controller) ListRooms(ctx context.Context, callback func(state *schema.ServerMessage)) error {
	return c.bus.ListRooms(ctx, func(room *schema.Room) {
		callback(&schema.ServerMessage{
			Payload: &schema.ServerMessage_LobbyMessage{
				LobbyMessage: &schema.ServerLobbyMessage{
					Payload: &schema.ServerLobbyMessage_RoomCreated{
						RoomCreated: &schema.RoomCreated{
							Room: room,
						},
					},
				},
			},
		})
	})
}

func (c *Controller) CreateRoom(ctx context.Context, tp token.Payload, payload *schema.CreateRoom) error {
	c.mu.Lock()
	_, ok := c.playerRoomMap[tp.PlayerID]
	c.mu.Unlock()
	if ok {
		return errors.EntityAlreadyExists(errors.EntityRoom)
	}

	p, err := c.store.GetPlayer(ctx, store.GetPlayerParams{
		PlayerID: &tp.PlayerID,
	})
	if err != nil {
		return err
	}

	roomID := uuid.New()
	room := &schema.Room{
		RoomId:      roomID.String(),
		DisplayName: payload.DisplayName,
		Players: []*schema.Player{
			{
				PlayerId:    p.PlayerID.String(),
				DisplayName: p.DisplayName,
				AvatarUrl:   p.ImageUrl,
				IsReady:     false,
				IsHost:      true,
			},
		},
		Status:   schema.RoomStatus_LOBBY,
		Capacity: payload.Capacity,
	}

	err = c.bus.SetRoom(ctx, room)
	if err != nil {
		return err
	}

	err = c.bus.Publish(ctx, event.LobbyChannel, &schema.ServerMessage{
		Payload: &schema.ServerMessage_LobbyMessage{
			LobbyMessage: &schema.ServerLobbyMessage{
				Payload: &schema.ServerLobbyMessage_RoomCreated{
					RoomCreated: &schema.RoomCreated{
						Room: room,
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.playerRoomMap[p.PlayerID] = roomID
	c.mu.Unlock()

	return nil
}

func (c *Controller) JoinRoom(ctx context.Context, tp token.Payload, payload *schema.JoinRoom) error {
	c.mu.Lock()
	_, ok := c.playerRoomMap[tp.PlayerID]
	c.mu.Unlock()
	if ok {
		return errors.EntityAlreadyExists(errors.EntityRoom)
	}

	room, err := c.bus.GetRoom(ctx, payload.RoomId)
	if err != nil {
		return err
	}

	p, err := c.store.GetPlayer(ctx, store.GetPlayerParams{
		PlayerID: &tp.PlayerID,
	})
	if err != nil {
		return err
	}

	player := &schema.Player{
		PlayerId:    p.PlayerID.String(),
		DisplayName: p.DisplayName,
		AvatarUrl:   p.ImageUrl,
		IsReady:     false,
		IsHost:      false,
	}

	room.Players = append(room.Players, player)

	err = c.bus.SetRoom(ctx, room)
	if err != nil {
		return err
	}

	err = c.bus.Publish(ctx, event.LobbyChannel, &schema.ServerMessage{
		Payload: &schema.ServerMessage_LobbyMessage{
			LobbyMessage: &schema.ServerLobbyMessage{
				Payload: &schema.ServerLobbyMessage_PlayerJoinedRoom{
					PlayerJoinedRoom: &schema.PlayerJoinedRoom{
						RoomId: room.RoomId,
						Player: player,
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	err = c.bus.Publish(ctx, event.RoomChannelPre+room.RoomId, &schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: &schema.ServerRoomMessage{
				Payload: &schema.ServerRoomMessage_PlayerJoinedRoom{
					PlayerJoinedRoom: &schema.PlayerJoinedRoom{
						RoomId: room.RoomId,
						Player: player,
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	roomID, err := uuid.Parse(room.RoomId)
	if err != nil {
		return errors.Internal(err)
	}

	c.mu.Lock()
	c.playerRoomMap[p.PlayerID] = roomID
	c.mu.Unlock()

	return nil
}

//func (c *Controller) LeaveRoom(ctx context.Context, tp token.Payload, payload *schema.LeaveRoom) error {
//
//}
