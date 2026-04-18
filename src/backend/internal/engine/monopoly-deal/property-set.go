package monopoly_deal

import (
	"slices"
	"strings"
)

type PropertySet struct {
	ID    Identifier `json:"id" msgpack:"a"`
	Color Color      `json:"color" msgpack:"b"`
	Cards Cards      `json:"cards" msgpack:"c"`
}

func NewPropertySet(id Identifier, color Color) PropertySet {
	return PropertySet{
		ID:    id,
		Color: color,
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

func (ps *PropertySet) IsComplete() bool {
	maxSetSize := CompleteSetSize[ps.Color]
	return len(ps.Cards) >= maxSetSize
}

var Rent = map[Color]map[int]int{
	ColorNone:     {0: 0},
	ColorBrown:    {1: 1, 2: 2},
	ColorSky:      {1: 1, 2: 2, 3: 3},
	ColorPink:     {1: 1, 2: 2, 3: 4},
	ColorOrange:   {1: 1, 2: 3, 3: 5},
	ColorRed:      {1: 2, 2: 3, 3: 6},
	ColorYellow:   {1: 2, 2: 4, 3: 6},
	ColorGreen:    {1: 2, 2: 4, 3: 7},
	ColorBlue:     {1: 3, 2: 8},
	ColorUtility:  {1: 1, 2: 2},
	ColorRailroad: {1: 1, 2: 2, 3: 3, 4: 4},
}

func (ps *PropertySet) Rent() int {
	n := ps.Cards.Len()

	additional := 0
	if ps.HasHouseLast() {
		n = n - 1
		additional = 3
	} else if ps.HasHotelLast() {
		n = n - 2
		additional = 4
	}

	return Rent[ps.Color][n] + additional
}

func (ps *PropertySet) HasHouseLast() bool {
	n := ps.Cards.Len()
	if n == 0 {
		return false
	}
	return ps.Cards[n-1].AssetKey == AssetKeyHouse
}

func (ps *PropertySet) HasHotelLast() bool {
	n := ps.Cards.Len()
	if n == 0 {
		return false
	}
	return ps.Cards[n-1].AssetKey == AssetKeyHotel
}

func (ps *PropertySet) IsLocked() bool {
	return ps.HasHouseLast() || ps.HasHotelLast()
}

func (ps *PropertySet) Index(cardID Identifier) int {
	for i, card := range ps.Cards {
		if card.ID == cardID {
			return i
		}
	}
	return -1
}

type PropertySets []PropertySet

func (ps *PropertySets) IndexBySetID(id Identifier) int {
	i, ok := slices.BinarySearchFunc(*ps, id, func(set PropertySet, id Identifier) int {
		return strings.Compare(string(set.ID), string(id))
	})
	if !ok {
		return -1
	}
	return i
}

func (ps *PropertySets) IndexByCardID(cardID Identifier) (int, int) {
	for i, set := range *ps {
		j := set.Index(cardID)
		if j != -1 {
			return i, j
		}
	}
	return -1, -1
}

func (ps *PropertySets) Add(p PropertySet) {
	*ps = append(*ps, p)
}

func (ps *PropertySets) RemoveByIdx(idx int) (PropertySet, bool) {
	if ps == nil {
		return PropertySet{}, false
	}

	sets := *ps
	if idx < 0 || idx >= len(sets) {
		return PropertySet{}, false
	}

	set := sets[idx]
	*ps = append(sets[:idx], sets[idx+1:]...)
	return set, true
}

func (ps *PropertySets) ColorRent(colors ...Color) int {
	if len(colors) == 0 {
		return 0
	}

	allowed := make(map[Color]struct{}, len(colors))
	for _, c := range colors {
		allowed[c] = struct{}{}
	}

	maxRent := 0
	for _, p := range *ps {
		if _, ok := allowed[p.Color]; ok {
			r := p.Rent()
			if r > maxRent {
				maxRent = r
			}
		}
	}
	return maxRent
}

func (ps *PropertySets) Rent() int {
	var maxRent int
	for _, p := range *ps {
		if p.Rent() > maxRent {
			maxRent = p.Rent()
		}
	}
	return maxRent
}

func (ps *PropertySets) Valid() bool {
	incompleteByColor := make(map[Color]int)
	for _, set := range *ps {
		if set.IsComplete() {
			continue
		}

		incompleteByColor[set.Color]++
		if incompleteByColor[set.Color] > 1 {
			return false
		}
	}

	return true
}

func (ps *PropertySets) Clean() {
	if ps == nil {
		return
	}

	sets := *ps
	filtered := sets[:0]
	for _, p := range sets {
		if p.Cards.Len() > 0 {
			filtered = append(filtered, p)
		}
	}

	*ps = filtered
}
