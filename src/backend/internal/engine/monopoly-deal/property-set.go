package monopoly_deal

import (
	"slices"

	"github.com/google/uuid"
)

type PropertySet struct {
	ID    uuid.UUID `json:"id"`
	Color Color     `json:"color"`
	Cards Cards     `json:"cards"`
}

func NewPropertySet(card Card) PropertySet {
	id, _ := uuid.NewV7()

	return PropertySet{
		ID:    id,
		Color: card.ActiveColor,
		Cards: Cards{card},
	}
}

var CompleteSetSize = map[Color]int{
	ColorBrown:    2,
	ColorSky:      3,
	ColorPink:     3,
	ColorOrange:   3,
	ColorRed:      3,
	ColorYellow:   3,
	ColorGreen:    3,
	ColorBlue:     2,
	ColorUtility:  2,
	ColorRailroad: 4,
}

func (ps PropertySet) IsFull() bool {
	maxSetSize := CompleteSetSize[ps.Color]
	return len(ps.Cards) >= maxSetSize
}

type PropertySets []PropertySet

func (ps *PropertySets) Index(setID uuid.UUID) int {
	i, ok := slices.BinarySearchFunc(*ps, setID, func(set PropertySet, u uuid.UUID) int {
		return slices.Compare(set.ID[:], u[:])
	})
	if !ok {
		return -1
	}
	return i
}

func (ps *PropertySets) Add(p PropertySet) {
	*ps = append(*ps, p)
}
