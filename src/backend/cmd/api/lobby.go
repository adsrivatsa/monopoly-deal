package main

import (
	"context"
	"monopoly-deal/internal/errors"
	"monopoly-deal/internal/schema"
	"monopoly-deal/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func (s *Server) CreateRoom(ctx context.Context, conn *websocket.Conn, payload *schema.CreateRoom) {
	playerID, err := uuid.Parse(payload.PlayerId)
	if err != nil {
		LobbyError(conn, errors.InvalidUUID(err))
		return
	}

	r, err := s.controller.CreateRoom(ctx, service.CreateRoomParams{
		PlayerID:    playerID,
		DisplayName: payload.DisplayName,
		Capacity:    payload.Capacity,
	})
	if err != nil {
		LobbyError(conn, err)
		return
	}

	res := &schema.LobbyMessage{
		Payload: &schema.LobbyMessage_CreateRoomRes{
			CreateRoomRes: &schema.CreateRoomResponse{
				PlayerId:    payload.PlayerId,
				RoomId:      r.RoomID.String(),
				DisplayName: r.DisplayName,
				Status:      r.RoomStatus.SchemaRoomStatus(),
				Capacity:    r.Capacity,
			},
		},
	}
	WriteWS(conn, res)
}
