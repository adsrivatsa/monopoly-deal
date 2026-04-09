package service

import (
	"context"
	"monopoly-deal/internal/errors"
	"monopoly-deal/internal/event"
	"monopoly-deal/internal/schema"
	"monopoly-deal/internal/store"
	"monopoly-deal/internal/token"
	"time"

	"github.com/google/uuid"
)

func (c *Controller) ListRooms(ctx context.Context, callback func(state *schema.ServerMessage)) error {
	return c.bus.List(ctx, event.RoomStatePre, func(key string, state *schema.ServerMessage) {
		callback(state)
	})
}

func (c *Controller) CreateRoom(ctx context.Context, tp token.Payload, payload *schema.CreateRoom) error {
	p, err := c.store.GetPlayer(ctx, store.GetPlayerParams{
		PlayerID: &tp.PlayerID,
	})
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.playerRoomMap[p.PlayerID]
	if ok {
		return errors.EntityAlreadyExists(errors.EntityRoom)
	}

	roomID := uuid.New()
	res := &schema.ServerMessage{
		Payload: &schema.ServerMessage_LobbyMessage{
			LobbyMessage: &schema.ServerLobbyMessage{
				Payload: &schema.ServerLobbyMessage_RoomCreated{
					RoomCreated: &schema.RoomCreated{
						Room: &schema.Room{
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
						},
					},
				},
			},
		},
	}

	roomKey := event.RoomStatePre + roomID.String()
	err = c.bus.Set(ctx, roomKey, res, time.Hour*24)
	if err != nil {
		return err
	}

	err = c.bus.Publish(ctx, event.LobbyChannel, res)
	if err != nil {
		return err
	}

	c.playerRoomMap[p.PlayerID] = roomID

	return nil
}

//func (c *Controller)

//func (c *Controller) JoinRoom(ctx context.Context, tp token.Payload, payload *schema.JoinRoom) error {
//	p, err := c.store.GetPlayer(ctx, store.GetPlayerParams{
//		PlayerID: &tp.PlayerID,
//	})
//	if err != nil {
//		return err
//	}
//
//	roomKey := event.RoomStatePre + payload.RoomId
//
//	var room schema.Room
//	err = c.bus.Get(ctx, roomKey, &room)
//	if err != nil {
//		return err
//	}
//
//	sp := &schema.Player{
//		PlayerId:    p.PlayerID.String(),
//		DisplayName: p.DisplayName,
//		AvatarUrl:   p.ImageUrl,
//		IsReady:     false,
//		IsHost:      false,
//	}
//	room.Players = append(room.Players, sp)
//	err = c.bus.Set(ctx, roomKey, &room)
//	if err != nil {
//		return err
//	}
//
//	res := &schema.LobbyMessage{
//		Payload: &schema.LobbyMessage_PlayerJoinedRoom{
//			PlayerJoinedRoom: &schema.PlayerJoinedRoom{
//				RoomId: room.RoomId,
//				Player: sp,
//			},
//		},
//	}
//	return c.bus.Publish(ctx, LobbyChannel, res)
//}
//

//func (c *Controller) LeaveRoom(ctx context.Context, tp token.Payload, payload *schema.LeaveRoom) error {
//
//}
