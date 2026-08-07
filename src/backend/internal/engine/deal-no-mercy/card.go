package deal_no_mercy

import (
	"fmt"
)

type AssetKey string

const (
	AssetKeyUnspecified AssetKey = ""

	// Pure properties (same roster as classic)
	AssetKeyBrown1  AssetKey = "brown1"
	AssetKeyBrown2  AssetKey = "brown2"
	AssetKeySky1    AssetKey = "sky1"
	AssetKeySky2    AssetKey = "sky2"
	AssetKeySky3    AssetKey = "sky3"
	AssetKeyPink1   AssetKey = "pink1"
	AssetKeyPink2   AssetKey = "pink2"
	AssetKeyPink3   AssetKey = "pink3"
	AssetKeyOrange1 AssetKey = "orange1"
	AssetKeyOrange2 AssetKey = "orange2"
	AssetKeyOrange3 AssetKey = "orange3"
	AssetKeyRed1    AssetKey = "red1"
	AssetKeyRed2    AssetKey = "red2"
	AssetKeyRed3    AssetKey = "red3"
	AssetKeyYellow1 AssetKey = "yellow1"
	AssetKeyYellow2 AssetKey = "yellow2"
	AssetKeyYellow3 AssetKey = "yellow3"
	AssetKeyGreen1  AssetKey = "green1"
	AssetKeyGreen2  AssetKey = "green2"
	AssetKeyGreen3  AssetKey = "green3"
	AssetKeyBlue1   AssetKey = "blue1"
	AssetKeyBlue2   AssetKey = "blue2"
	AssetKeyUtil1   AssetKey = "util1"
	AssetKeyUtil2   AssetKey = "util2"
	AssetKeyRail1   AssetKey = "rail1"
	AssetKeyRail2   AssetKey = "rail2"
	AssetKeyRail3   AssetKey = "rail3"
	AssetKeyRail4   AssetKey = "rail4"

	// Wild properties (same roster as classic)
	AssetKeyWildBrownSky        AssetKey = "wild_brown_sky"
	AssetKeyWildSkyRailroad     AssetKey = "wild_sky_rail"
	AssetKeyWildPinkOrange      AssetKey = "wild_pink_orange"
	AssetKeyWildRedYellow       AssetKey = "wild_red_yellow"
	AssetKeyWildGreenBlue       AssetKey = "wild_green_blue"
	AssetKeyWildGreenRailroad   AssetKey = "wild_green_rail"
	AssetKeyWildUtilityRailroad AssetKey = "wild_util_rail"
	AssetKeyWildWild            AssetKey = "wild_wild"

	// Money (incl. the new 15M)
	AssetKeyMoney1  AssetKey = "money1"
	AssetKeyMoney2  AssetKey = "money2"
	AssetKeyMoney3  AssetKey = "money3"
	AssetKeyMoney4  AssetKey = "money4"
	AssetKeyMoney5  AssetKey = "money5"
	AssetKeyMoney10 AssetKey = "money10"
	AssetKeyMoney15 AssetKey = "money15"

	// No Mercy actions
	AssetKeySetSnatcher  AssetKey = "set_snatcher"
	AssetKeyDebtTrap     AssetKey = "debt_trap"
	AssetKeyGoAgain      AssetKey = "go_again"
	AssetKeyHeist        AssetKey = "heist"
	AssetKeyMarketCrash  AssetKey = "market_crash"
	AssetKeyBigPayday    AssetKey = "big_payday"
	AssetKeyRepoMan      AssetKey = "repo_man"
	AssetKeyShack        AssetKey = "shack"
	AssetKeyPropertyRaid AssetKey = "property_raid"
	AssetKeyTaxDay       AssetKey = "tax_day"
	AssetKeyPickpocket   AssetKey = "pickpocket"
	AssetKeyBankSwap     AssetKey = "bank_swap"
	AssetKeyYoink        AssetKey = "yoink"
	AssetKeyNah          AssetKey = "nah"

	// Rent actions (all charge every opponent)
	AssetKeyRentWild            AssetKey = "rent_wild"
	AssetKeyRentBrownSky        AssetKey = "rent_brown_sky"
	AssetKeyRentPinkOrange      AssetKey = "rent_pink_orange"
	AssetKeyRentRedYellow       AssetKey = "rent_red_yellow"
	AssetKeyRentGreenBlue       AssetKey = "rent_green_blue"
	AssetKeyRentUtilityRailroad AssetKey = "rent_util_rail"

	AssetKeyDoubleRentWild            AssetKey = "double_rent_wild"
	AssetKeyDoubleRentBrownSky        AssetKey = "double_rent_brown_sky"
	AssetKeyDoubleRentPinkOrange      AssetKey = "double_rent_pink_orange"
	AssetKeyDoubleRentRedYellow       AssetKey = "double_rent_red_yellow"
	AssetKeyDoubleRentGreenBlue       AssetKey = "double_rent_green_blue"
	AssetKeyDoubleRentUtilityRailroad AssetKey = "double_rent_util_rail"
)

func AllAssetKeys() []AssetKey {
	return []AssetKey{
		AssetKeyBrown1, AssetKeyBrown2, AssetKeySky1, AssetKeySky2, AssetKeySky3,
		AssetKeyPink1, AssetKeyPink2, AssetKeyPink3, AssetKeyOrange1, AssetKeyOrange2, AssetKeyOrange3,
		AssetKeyRed1, AssetKeyRed2, AssetKeyRed3, AssetKeyYellow1, AssetKeyYellow2, AssetKeyYellow3,
		AssetKeyGreen1, AssetKeyGreen2, AssetKeyGreen3, AssetKeyBlue1, AssetKeyBlue2,
		AssetKeyUtil1, AssetKeyUtil2, AssetKeyRail1, AssetKeyRail2, AssetKeyRail3, AssetKeyRail4,
		AssetKeyWildBrownSky, AssetKeyWildSkyRailroad, AssetKeyWildPinkOrange, AssetKeyWildRedYellow,
		AssetKeyWildGreenBlue, AssetKeyWildGreenRailroad, AssetKeyWildUtilityRailroad, AssetKeyWildWild,
		AssetKeyMoney1, AssetKeyMoney2, AssetKeyMoney3, AssetKeyMoney4, AssetKeyMoney5, AssetKeyMoney10, AssetKeyMoney15,
		AssetKeySetSnatcher, AssetKeyDebtTrap, AssetKeyGoAgain, AssetKeyHeist, AssetKeyMarketCrash,
		AssetKeyBigPayday, AssetKeyRepoMan, AssetKeyShack, AssetKeyPropertyRaid, AssetKeyTaxDay,
		AssetKeyPickpocket, AssetKeyBankSwap, AssetKeyYoink, AssetKeyNah,
		AssetKeyRentWild, AssetKeyRentBrownSky, AssetKeyRentPinkOrange, AssetKeyRentRedYellow,
		AssetKeyRentGreenBlue, AssetKeyRentUtilityRailroad,
		AssetKeyDoubleRentWild, AssetKeyDoubleRentBrownSky, AssetKeyDoubleRentPinkOrange,
		AssetKeyDoubleRentRedYellow, AssetKeyDoubleRentGreenBlue, AssetKeyDoubleRentUtilityRailroad,
	}
}

type Category int

const (
	CategoryUnspecified Category = iota
	CategoryPureProperty
	CategoryWildProperty
	CategoryMoney
	CategoryAction
)

type Color int

const (
	ColorUnspecified Color = iota
	ColorBrown
	ColorSky
	ColorPink
	ColorOrange
	ColorRed
	ColorYellow
	ColorGreen
	ColorBlue
	ColorUtility
	ColorRailroad
)

func AllColors() []Color {
	return []Color{ColorBrown, ColorSky, ColorPink, ColorOrange, ColorRed, ColorYellow, ColorGreen, ColorBlue, ColorUtility, ColorRailroad}
}

// StealCategory is the category named when playing PICKPOCKET.
type StealCategory int

const (
	StealCategoryUnspecified StealCategory = iota
	StealCategoryProperty
	StealCategoryMoney
	StealCategoryAction
)

func (sc StealCategory) Matches(c Card) bool {
	switch sc {
	case StealCategoryProperty:
		return c.Category == CategoryPureProperty || c.Category == CategoryWildProperty
	case StealCategoryMoney:
		return c.Category == CategoryMoney
	case StealCategoryAction:
		return c.Category == CategoryAction
	default:
		return false
	}
}

type Card struct {
	ID          Identifier `json:"id" msgpack:"a"`
	Category    Category   `json:"category" msgpack:"b"`
	AssetKey    AssetKey   `json:"card_key" msgpack:"c"`
	Value       int        `json:"value" msgpack:"d"`
	Colors      []Color    `json:"colors" msgpack:"e"`
	ActiveColor Color      `json:"active_color" msgpack:"f"`
}

func NewCard(id Identifier, c Category, ak AssetKey, v int, colors ...Color) Card {
	activeColor := ColorUnspecified
	if len(colors) > 0 {
		activeColor = colors[0]
	}

	return Card{
		ID:          id,
		Category:    c,
		AssetKey:    ak,
		Value:       v,
		Colors:      append([]Color(nil), colors...),
		ActiveColor: activeColor,
	}
}

// DO NOT USE THIS MAP TO FETCH CARDS!
// Cards are NOT unique here.
var CardByAssetKey = map[AssetKey]Card{
	AssetKeyBrown1:  NewCard("", CategoryPureProperty, AssetKeyBrown1, 1, ColorBrown),
	AssetKeyBrown2:  NewCard("", CategoryPureProperty, AssetKeyBrown2, 1, ColorBrown),
	AssetKeySky1:    NewCard("", CategoryPureProperty, AssetKeySky1, 1, ColorSky),
	AssetKeySky2:    NewCard("", CategoryPureProperty, AssetKeySky2, 1, ColorSky),
	AssetKeySky3:    NewCard("", CategoryPureProperty, AssetKeySky3, 1, ColorSky),
	AssetKeyPink1:   NewCard("", CategoryPureProperty, AssetKeyPink1, 2, ColorPink),
	AssetKeyPink2:   NewCard("", CategoryPureProperty, AssetKeyPink2, 2, ColorPink),
	AssetKeyPink3:   NewCard("", CategoryPureProperty, AssetKeyPink3, 2, ColorPink),
	AssetKeyOrange1: NewCard("", CategoryPureProperty, AssetKeyOrange1, 2, ColorOrange),
	AssetKeyOrange2: NewCard("", CategoryPureProperty, AssetKeyOrange2, 2, ColorOrange),
	AssetKeyOrange3: NewCard("", CategoryPureProperty, AssetKeyOrange3, 2, ColorOrange),
	AssetKeyRed1:    NewCard("", CategoryPureProperty, AssetKeyRed1, 3, ColorRed),
	AssetKeyRed2:    NewCard("", CategoryPureProperty, AssetKeyRed2, 3, ColorRed),
	AssetKeyRed3:    NewCard("", CategoryPureProperty, AssetKeyRed3, 3, ColorRed),
	AssetKeyYellow1: NewCard("", CategoryPureProperty, AssetKeyYellow1, 3, ColorYellow),
	AssetKeyYellow2: NewCard("", CategoryPureProperty, AssetKeyYellow2, 3, ColorYellow),
	AssetKeyYellow3: NewCard("", CategoryPureProperty, AssetKeyYellow3, 3, ColorYellow),
	AssetKeyGreen1:  NewCard("", CategoryPureProperty, AssetKeyGreen1, 4, ColorGreen),
	AssetKeyGreen2:  NewCard("", CategoryPureProperty, AssetKeyGreen2, 4, ColorGreen),
	AssetKeyGreen3:  NewCard("", CategoryPureProperty, AssetKeyGreen3, 4, ColorGreen),
	AssetKeyBlue1:   NewCard("", CategoryPureProperty, AssetKeyBlue1, 4, ColorBlue),
	AssetKeyBlue2:   NewCard("", CategoryPureProperty, AssetKeyBlue2, 4, ColorBlue),
	AssetKeyUtil1:   NewCard("", CategoryPureProperty, AssetKeyUtil1, 2, ColorUtility),
	AssetKeyUtil2:   NewCard("", CategoryPureProperty, AssetKeyUtil2, 2, ColorUtility),
	AssetKeyRail1:   NewCard("", CategoryPureProperty, AssetKeyRail1, 2, ColorRailroad),
	AssetKeyRail2:   NewCard("", CategoryPureProperty, AssetKeyRail2, 2, ColorRailroad),
	AssetKeyRail3:   NewCard("", CategoryPureProperty, AssetKeyRail3, 2, ColorRailroad),
	AssetKeyRail4:   NewCard("", CategoryPureProperty, AssetKeyRail4, 2, ColorRailroad),

	AssetKeyWildBrownSky:        NewCard("", CategoryWildProperty, AssetKeyWildBrownSky, 1, ColorBrown, ColorSky),
	AssetKeyWildSkyRailroad:     NewCard("", CategoryWildProperty, AssetKeyWildSkyRailroad, 4, ColorSky, ColorRailroad),
	AssetKeyWildPinkOrange:      NewCard("", CategoryWildProperty, AssetKeyWildPinkOrange, 2, ColorPink, ColorOrange),
	AssetKeyWildRedYellow:       NewCard("", CategoryWildProperty, AssetKeyWildRedYellow, 3, ColorRed, ColorYellow),
	AssetKeyWildGreenBlue:       NewCard("", CategoryWildProperty, AssetKeyWildGreenBlue, 4, ColorGreen, ColorBlue),
	AssetKeyWildGreenRailroad:   NewCard("", CategoryWildProperty, AssetKeyWildGreenRailroad, 4, ColorGreen, ColorRailroad),
	AssetKeyWildUtilityRailroad: NewCard("", CategoryWildProperty, AssetKeyWildUtilityRailroad, 2, ColorUtility, ColorRailroad),
	AssetKeyWildWild:            NewCard("", CategoryWildProperty, AssetKeyWildWild, 0, ColorBrown, ColorSky, ColorPink, ColorOrange, ColorRed, ColorYellow, ColorGreen, ColorBlue, ColorUtility, ColorRailroad),

	AssetKeyMoney1:  NewCard("", CategoryMoney, AssetKeyMoney1, 1),
	AssetKeyMoney2:  NewCard("", CategoryMoney, AssetKeyMoney2, 2),
	AssetKeyMoney3:  NewCard("", CategoryMoney, AssetKeyMoney3, 3),
	AssetKeyMoney4:  NewCard("", CategoryMoney, AssetKeyMoney4, 4),
	AssetKeyMoney5:  NewCard("", CategoryMoney, AssetKeyMoney5, 5),
	AssetKeyMoney10: NewCard("", CategoryMoney, AssetKeyMoney10, 10),
	AssetKeyMoney15: NewCard("", CategoryMoney, AssetKeyMoney15, 15),

	AssetKeySetSnatcher:  NewCard("", CategoryAction, AssetKeySetSnatcher, 3),
	AssetKeyDebtTrap:     NewCard("", CategoryAction, AssetKeyDebtTrap, 2),
	AssetKeyGoAgain:      NewCard("", CategoryAction, AssetKeyGoAgain, 1),
	AssetKeyHeist:        NewCard("", CategoryAction, AssetKeyHeist, 4),
	AssetKeyMarketCrash:  NewCard("", CategoryAction, AssetKeyMarketCrash, 4),
	AssetKeyBigPayday:    NewCard("", CategoryAction, AssetKeyBigPayday, 1),
	AssetKeyRepoMan:      NewCard("", CategoryAction, AssetKeyRepoMan, 5),
	AssetKeyShack:        NewCard("", CategoryAction, AssetKeyShack, 3),
	AssetKeyPropertyRaid: NewCard("", CategoryAction, AssetKeyPropertyRaid, 3),
	AssetKeyTaxDay:       NewCard("", CategoryAction, AssetKeyTaxDay, 5),
	AssetKeyPickpocket:   NewCard("", CategoryAction, AssetKeyPickpocket, 5),
	AssetKeyBankSwap:     NewCard("", CategoryAction, AssetKeyBankSwap, 4),
	AssetKeyYoink:        NewCard("", CategoryAction, AssetKeyYoink, 2),
	AssetKeyNah:          NewCard("", CategoryAction, AssetKeyNah, 4),

	AssetKeyRentWild:            NewCard("", CategoryAction, AssetKeyRentWild, 3),
	AssetKeyRentBrownSky:        NewCard("", CategoryAction, AssetKeyRentBrownSky, 1),
	AssetKeyRentPinkOrange:      NewCard("", CategoryAction, AssetKeyRentPinkOrange, 1),
	AssetKeyRentRedYellow:       NewCard("", CategoryAction, AssetKeyRentRedYellow, 1),
	AssetKeyRentGreenBlue:       NewCard("", CategoryAction, AssetKeyRentGreenBlue, 1),
	AssetKeyRentUtilityRailroad: NewCard("", CategoryAction, AssetKeyRentUtilityRailroad, 1),

	AssetKeyDoubleRentWild:            NewCard("", CategoryAction, AssetKeyDoubleRentWild, 3),
	AssetKeyDoubleRentBrownSky:        NewCard("", CategoryAction, AssetKeyDoubleRentBrownSky, 1),
	AssetKeyDoubleRentPinkOrange:      NewCard("", CategoryAction, AssetKeyDoubleRentPinkOrange, 1),
	AssetKeyDoubleRentRedYellow:       NewCard("", CategoryAction, AssetKeyDoubleRentRedYellow, 1),
	AssetKeyDoubleRentGreenBlue:       NewCard("", CategoryAction, AssetKeyDoubleRentGreenBlue, 1),
	AssetKeyDoubleRentUtilityRailroad: NewCard("", CategoryAction, AssetKeyDoubleRentUtilityRailroad, 1),
}

// rentCardSpec resolves a rent-family card into the set of colors it may
// charge in and its rent multiplier. ok is false for non-rent cards.
func rentCardSpec(key AssetKey) (allowed []Color, multiplier int, ok bool) {
	switch key {
	case AssetKeyRentWild:
		return AllColors(), 1, true
	case AssetKeyDoubleRentWild:
		return AllColors(), 2, true
	case AssetKeyRentBrownSky:
		return []Color{ColorBrown, ColorSky}, 1, true
	case AssetKeyRentPinkOrange:
		return []Color{ColorPink, ColorOrange}, 1, true
	case AssetKeyRentRedYellow:
		return []Color{ColorRed, ColorYellow}, 1, true
	case AssetKeyRentGreenBlue:
		return []Color{ColorGreen, ColorBlue}, 1, true
	case AssetKeyRentUtilityRailroad:
		return []Color{ColorUtility, ColorRailroad}, 1, true
	case AssetKeyDoubleRentBrownSky:
		return []Color{ColorBrown, ColorSky}, 2, true
	case AssetKeyDoubleRentPinkOrange:
		return []Color{ColorPink, ColorOrange}, 2, true
	case AssetKeyDoubleRentRedYellow:
		return []Color{ColorRed, ColorYellow}, 2, true
	case AssetKeyDoubleRentGreenBlue:
		return []Color{ColorGreen, ColorBlue}, 2, true
	case AssetKeyDoubleRentUtilityRailroad:
		return []Color{ColorUtility, ColorRailroad}, 2, true
	default:
		return nil, 0, false
	}
}

func (c Card) String() string {
	return fmt.Sprintf("%s:%s", c.ID, c.AssetKey)
}

func (c Card) HasColor(color Color) bool {
	for _, col := range c.Colors {
		if col == color {
			return true
		}
	}
	return false
}

type Cards []Card

func (c *Cards) Len() int {
	return len(*c)
}

func (c *Cards) Add(card ...Card) {
	*c = append(*c, card...)
}

func (c *Cards) RemoveByID(id Identifier) (Card, bool) {
	if c == nil {
		return Card{}, false
	}

	cards := *c
	for i, c1 := range cards {
		if c1.ID == id {
			*c = append(cards[:i], cards[i+1:]...)
			return c1, true
		}
	}

	return Card{}, false
}

func (c *Cards) RemoveByIdx(idx int) (Card, bool) {
	if c == nil {
		return Card{}, false
	}

	if idx < 0 || idx >= c.Len() {
		return Card{}, false
	}

	cards := *c
	card := cards[idx]
	*c = append(cards[:idx], cards[idx+1:]...)
	return card, true
}

func (c *Cards) Value() int {
	value := 0
	for _, card := range *c {
		value += card.Value
	}
	return value
}
