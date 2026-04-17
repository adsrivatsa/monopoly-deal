package store

import "fun-kames/internal/schema"

func (g GameType) Proto() schema.Game {
	switch g {
	case GameTypeMonopolyDeal:
		return schema.Game_MonopolyDeal
	default:
		return schema.Game_MonopolyDeal
	}
}
