package service

import (
	"context"
	"fmt"
	"monopoly-deal/internal/schema"
	"monopoly-deal/internal/store"
	"monopoly-deal/internal/token"

	"github.com/google/uuid"
)

const (
	LobbyFmt = "lobby"
	RoomFmt  = "room:%s"
)

func (c *Controller) CreateRoom(ctx context.Context, tp token.Payload, payload *schema.CreateRoom) error {
	p, err := c.store.GetPlayer(ctx, store.GetPlayerParams{
		PlayerID: &tp.PlayerID,
	})
	if err != nil {
		return err
	}

	room := &schema.Room{
		RoomId:      uuid.NewString(),
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

	roomKey := fmt.Sprintf(RoomFmt, room.RoomId)
	err = c.bus.Set(ctx, roomKey, room)
	if err != nil {
		return err
	}

	res := &schema.LobbyMessage{
		Payload: &schema.LobbyMessage_RoomCreated{
			RoomCreated: &schema.RoomCreated{
				Room: room,
			},
		},
	}
	return c.bus.Publish(ctx, LobbyFmt, res)
}

func (c *Controller) JoinRoom(ctx context.Context, tp token.Payload, payload *schema.JoinRoom) error {
	p, err := c.store.GetPlayer(ctx, store.GetPlayerParams{
		PlayerID: &tp.PlayerID,
	})
	if err != nil {
		return err
	}

	roomKey := fmt.Sprintf(RoomFmt, payload.RoomId)

	var room *schema.Room
	err = c.bus.Get(ctx, roomKey, &room)
	if err != nil {
		return err
	}

	sp := &schema.Player{
		PlayerId:    p.PlayerID.String(),
		DisplayName: p.DisplayName,
		AvatarUrl:   p.ImageUrl,
		IsReady:     false,
		IsHost:      false,
	}
	room.Players = append(room.Players, sp)
	err = c.bus.Set(ctx, roomKey, room)
	if err != nil {
		return err
	}

	res := &schema.LobbyMessage{
		Payload: &schema.LobbyMessage_PlayerJoinedRoom{
			PlayerJoinedRoom: &schema.PlayerJoinedRoom{
				RoomId: room.RoomId,
				Player: sp,
			},
		},
	}
	return c.bus.Publish(ctx, LobbyFmt, res)
}
