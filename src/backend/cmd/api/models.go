package main

import (
	"monopoly-deal/internal/service"

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

type CreateRoomParams struct {
	DisplayName string `json:"display_name"`
	Capacity    int    `json:"capacity" validate:"min=2,max=5"`
}

type Room = service.Room

type ListRoomRes = service.ListRoomsRes

type ListRoomsParams struct {
	Limit  int32   `json:"limit" validate:"min=0,max=100"`
	Offset int32   `json:"offset" validate:"min=0"`
	Search *string `json:"search"`
}
