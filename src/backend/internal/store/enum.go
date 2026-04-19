package store

import (
	"fun-kames/internal/schema/room_schema"
)

func (g GameType) Proto() room_schema.Game {
	switch g {
	case GameTypeMonopolyDeal:
		return room_schema.Game_MonopolyDeal
	default:
		return room_schema.Game_MonopolyDeal
	}
}
