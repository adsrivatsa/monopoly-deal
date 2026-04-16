package game_settings

import (
	"encoding/json"
	"errors"
	"fun-kames/internal/store"
)

type Settings interface {
	Raw() any
}

type MonopolyDeal struct {
	NumDecks int32 `json:"num_decks" validate:"required,min=1,max=3"`
}

func (m MonopolyDeal) Raw() any {
	return m
}

func ParseSettings(game store.Game, settings string) (Settings, error) {
	switch game {
	case store.GameMonopolyDeal:
		var set MonopolyDeal
		err := json.Unmarshal([]byte(settings), &set)
		return set, err
	}

	return nil, errors.New("game not supported")
}
