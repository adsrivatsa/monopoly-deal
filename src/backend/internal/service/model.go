package service

import (
	"monopoly-deal/internal/store"

	"github.com/google/uuid"
)

type Player = store.Player

type ShortPlayer struct {
	PlayerID    uuid.UUID `json:"player_id"`
	DisplayName string    `json:"display_name"`
	ImageUrl    string    `json:"image_url"`
}

type CreatePlayerParams = store.CreatePlayerParams

type GetPlayerParams = store.GetPlayerParams

type UpdatePlayerParams = store.UpdatePlayerParams

type CreateRoomParams = store.CreateRoomParams

type Room = store.Room

type ShortRoom struct {
}

type LongRoom struct {
	RoomID      uuid.UUID   `json:"room_id"`
	DisplayName string      `json:"display_name"`
	Capacity    int32       `json:"capacity"`
	Occupied    int32       `json:"occupied"`
	Host        ShortPlayer `json:"host"`
}

type ListRoomsParams = store.ListRoomPlayersParams

type ListRoomsRes struct {
	TotalCount int64      `json:"total_count"`
	Rooms      []LongRoom `json:"rooms"`
}
