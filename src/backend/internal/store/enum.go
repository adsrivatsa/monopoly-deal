package store

import (
	"the-deal/internal/schema/room_schema"
)

func (g GameType) Proto() room_schema.Game {
	switch g {
	case GameTypeMonopolyDeal:
		return room_schema.Game_MonopolyDeal
	case GameTypeDealNoMercy:
		return room_schema.Game_DealNoMercy
	default:
		return room_schema.Game_MonopolyDeal
	}
}
