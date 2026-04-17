package service

import (
	"fun-kames/internal/store"

	"github.com/google/uuid"
)

type Player = store.Player

type ShortPlayer struct {
	PlayerID    uuid.UUID `json:"player_id"`
	DisplayName string    `json:"display_name"`
	ImageUrl    string    `json:"image_url"`
	IsHost      bool      `json:"is_host"`
	IsReady     bool      `json:"is_ready"`
}

type CreatePlayerParams = store.CreatePlayerParams

type GetPlayerParams = store.GetPlayerParams

type UpdatePlayerParams = store.UpdatePlayerParams

type CreateRoomParams struct {
	DisplayName string         `json:"display_name"`
	Capacity    int32          `json:"capacity"`
	Game        store.GameType `json:"game"`
	Settings    string         `json:"game-settings"`
}

type Room struct {
	RoomID      uuid.UUID      `json:"room_id"`
	DisplayName string         `json:"display_name"`
	Capacity    int32          `json:"capacity"`
	Occupied    int32          `json:"occupied"`
	Game        store.GameType `json:"game"`
	Settings    string         `json:"settings"`
}

type ShortRoom struct {
}

type LongRoom struct {
	RoomID      uuid.UUID      `json:"room_id"`
	DisplayName string         `json:"display_name"`
	Capacity    int32          `json:"capacity"`
	Occupied    int32          `json:"occupied"`
	Players     []ShortPlayer  `json:"players"`
	Game        store.GameType `json:"game"`
	Settings    string         `json:"settings"`
}

type ListRoomsParams struct {
	Limit  int32           `json:"limit"`
	Offset int32           `json:"offset"`
	Search *string         `json:"search"`
	Game   *store.GameType `json:"game"`
}

type ListRoomsRes struct {
	TotalCount int64      `json:"total_count"`
	Rooms      []LongRoom `json:"rooms"`
}

type UpdateRoomSettingsParams struct {
	Capacity int32          `json:"capacity"`
	Game     store.GameType `json:"game"`
	Settings string         `json:"game-settings"`
}
