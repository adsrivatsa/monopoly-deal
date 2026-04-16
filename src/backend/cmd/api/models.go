package main

import (
	"fun-kames/internal/store"

	"github.com/google/uuid"
)

type GetPlayerParams struct {
	PlayerID *uuid.UUID `json:"player_id,omitempty" validate:"omitempty,uuid"`
	Email    *string    `json:"email,omitempty" validate:"omitempty,email"`
}

type Player struct {
	PlayerID    uuid.UUID `json:"player_id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	ImageUrl    string    `json:"image_url"`
}

type UpdatePlayerParams struct {
	DisplayName string `json:"display_name"`
}

type CreateRoomParams struct {
	DisplayName string     `json:"display_name" validate:"required"`
	Capacity    int        `json:"capacity" validate:"min=2,max=15"`
	Game        store.Game `json:"game" validate:"required,game"`
	Settings    string     `json:"settings" validate:"required,json"`
}

type ListRoomsParams struct {
	Limit  int32       `json:"limit" validate:"min=0,max=100"`
	Offset int32       `json:"offset" validate:"min=0"`
	Search *string     `json:"search"`
	Game   *store.Game `json:"game" validate:"omitempty,game"`
}

type UpdateRoomSettingsParams struct {
	Capacity int        `json:"capacity" validate:"min=2,max=15"`
	Game     store.Game `json:"game" validate:"game"`
	Settings string     `json:"settings" validate:"required,json"`
}
