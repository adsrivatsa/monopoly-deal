package service

import (
	"monopoly-deal/internal/store"

	"github.com/google/uuid"
)

type Player = store.Player

type CreatePlayerParams = store.CreatePlayerParams

type GetPlayerParams = store.GetPlayerParams

type Room = store.Room

type CreateRoomParams struct {
	PlayerID    uuid.UUID
	DisplayName string
	Capacity    int32
}
