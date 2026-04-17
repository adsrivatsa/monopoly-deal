package engine

import (
	"encoding/json"
	"errors"
	monopoly_deal "fun-kames/internal/engine/monopoly-deal"
	"fun-kames/internal/store"
)

type Settings interface {
	Raw() any
}

func ParseSettings(game store.GameType, settings string) (Settings, error) {
	switch game {
	case store.GameTypeMonopolyDeal:
		var set monopoly_deal.Settings
		err := json.Unmarshal([]byte(settings), &set)
		return set, err
	}

	return nil, errors.New("game not supported")
}
