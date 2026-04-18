package engine

import (
	"errors"
	monopoly_deal "fun-kames/internal/engine/monopoly-deal"
	"fun-kames/internal/store"

	"github.com/vmihailenco/msgpack/v5"
)

type Settings interface {
	Raw() any
	Encode() ([]byte, error)
}

func ParseSettings(game store.GameType, settings []byte) (Settings, error) {
	switch game {
	case store.GameTypeMonopolyDeal:
		var set monopoly_deal.Settings
		err := msgpack.Unmarshal(settings, &set)
		return set, err
	}

	return nil, errors.New("game not supported")
}
