package store

import "fun-kames/internal/schema"

func (g Game) Proto() schema.Game {
	switch g {
	case GameMonopolyDeal:
		return schema.Game_MonopolyDeal
	default:
		return schema.Game_MonopolyDeal
	}
}
