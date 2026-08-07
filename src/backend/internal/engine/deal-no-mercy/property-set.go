package deal_no_mercy

// PropertySet is a group of property cards of one color. Unlike classic
// (which stored house/hotel inside Cards with locking rules), the No Mercy
// SHACK is modeled as a dedicated attachment: it is immovable, travels with
// the set, is never part of Cards, and is discarded when the set empties.
type PropertySet struct {
	ID    Identifier `json:"id" msgpack:"a"`
	Color Color      `json:"color" msgpack:"b"`
	Cards Cards      `json:"cards" msgpack:"c"`
	Shack *Card      `json:"shack" msgpack:"d"`
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

func (ps *PropertySet) HasShack() bool {
	return ps.Shack != nil
}

var RentTable = map[Color]map[int]int{
	ColorUnspecified: {0: 0},
	ColorBrown:       {1: 1, 2: 2},
	ColorSky:         {1: 1, 2: 2, 3: 3},
	ColorPink:        {1: 1, 2: 2, 3: 4},
	ColorOrange:      {1: 1, 2: 3, 3: 5},
	ColorRed:         {1: 2, 2: 3, 3: 6},
	ColorYellow:      {1: 2, 2: 4, 3: 6},
	ColorGreen:       {1: 2, 2: 4, 3: 7},
	ColorBlue:        {1: 3, 2: 8},
	ColorUtility:     {1: 1, 2: 2},
	ColorRailroad:    {1: 1, 2: 2, 3: 3, 4: 4},
}

// Rent returns the rent for the set including the shack bonus (from
// Settings.ShackRentBonus) if a shack is attached.
func (ps *PropertySet) Rent(shackBonus int) int {
	n := ps.Cards.Len()
	if n == 0 {
		return 0
	}

	if max, ok := CompleteSetSize[ps.Color]; ok && n > max {
		n = max
	}

	rent := RentTable[ps.Color][n]
	if ps.HasShack() {
		rent += shackBonus
	}
	return rent
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
	for i, set := range *ps {
		if set.ID == id {
			return i
		}
	}

	return -1
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

// ColorRent returns the best rent among the player's sets of the given
// colors, including shack bonuses.
func (ps *PropertySets) ColorRent(shackBonus int, colors ...Color) int {
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
			r := p.Rent(shackBonus)
			if r > maxRent {
				maxRent = r
			}
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

// Clean removes sets whose Cards are empty and returns the shacks that were
// attached to those emptied sets — the caller must discard them back to the
// deck (a shack cannot exist without a set).
func (ps *PropertySets) Clean() Cards {
	if ps == nil {
		return nil
	}

	var orphanedShacks Cards
	sets := *ps
	filtered := sets[:0]
	for _, p := range sets {
		if p.Cards.Len() > 0 {
			filtered = append(filtered, p)
			continue
		}
		if p.Shack != nil {
			orphanedShacks.Add(*p.Shack)
		}
	}

	*ps = filtered
	return orphanedShacks
}

func (ps *PropertySets) CompleteCount() int {
	if ps == nil {
		return 0
	}

	count := 0
	for _, p := range *ps {
		if p.IsComplete() {
			count++
		}
	}
	return count
}
