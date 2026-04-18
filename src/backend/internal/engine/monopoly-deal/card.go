package monopoly_deal

import (
	"fmt"
	"slices"
)

type AssetKey string

const (
	AssetKeyBalticAve            AssetKey = "baltic-ave"
	AssetKeyMediterraneanAve     AssetKey = "mediterranean-ave"
	AssetKeyConnecticutAve       AssetKey = "connecticut-ave"
	AssetKeyOrientalAve          AssetKey = "oriental-ave"
	AssetKeyVermontAve           AssetKey = "vermont-ave"
	AssetKeyStCharlesPlace       AssetKey = "st-charles-place"
	AssetKeyVirginiaAve          AssetKey = "virginia-ave"
	AssetKeyStateAve             AssetKey = "state-ave"
	AssetKeyNewYorkAve           AssetKey = "new-york-ave"
	AssetKeyStJamesPlace         AssetKey = "st-james-place"
	AssetKeyTennesseeAve         AssetKey = "tennessee-ave"
	AssetKeyKentuckyAve          AssetKey = "kentucky-ave"
	AssetKeyIndianaAve           AssetKey = "indiana-ave"
	AssetKeyIllinoisAve          AssetKey = "illinois-ave"
	AssetKeyVentnorAve           AssetKey = "ventnor-ave"
	AssetKeyMarvinGardens        AssetKey = "marvin-gardens"
	AssetKeyAtlanticAve          AssetKey = "atlantic-ave"
	AssetKeyNorthCarolinaAve     AssetKey = "north-carolina-ave"
	AssetKeyPacificAve           AssetKey = "pacific-ave"
	AssetKeyPennsylvaniaAve      AssetKey = "pennsylvania-ave"
	AssetKeyBoardwalk            AssetKey = "boardwalk"
	AssetKeyParkPlace            AssetKey = "park-place"
	AssetKeyWaterWorks           AssetKey = "water-works"
	AssetKeyElectricCompany      AssetKey = "electric-company"
	AssetKeyShortLine            AssetKey = "short-line"
	AssetKeyBandORailRoad        AssetKey = "b-and-o-railroad"
	AssetKeyReadingRailroad      AssetKey = "reading-railroad"
	AssetKeyPennsylvaniaRailroad AssetKey = "pennsylvania-railroad"
	AssetKeyWildBrownSky         AssetKey = "wild-brown-sky"
	AssetKeyWildSkyRailroad      AssetKey = "wild-sky-railroad"
	AssetKeyWildPinkOrange       AssetKey = "wild-pink-orange"
	AssetKeyWildRedYellow        AssetKey = "wild-red-yellow"
	AssetKeyWildGreenBlue        AssetKey = "wild-green-blue"
	AssetKeyWildGreenRailroad    AssetKey = "wild-green-railroad"
	AssetKeyWildUtilityRailroad  AssetKey = "wild-utility-railroad"
	AssetKeyWildWild             AssetKey = "wild-wild"
	AssetKeyMoney10              AssetKey = "money-10"
	AssetKeyMoney5               AssetKey = "money-5"
	AssetKeyMoney4               AssetKey = "money-4"
	AssetKeyMoney3               AssetKey = "money-3"
	AssetKeyMoney2               AssetKey = "money-2"
	AssetKeyMoney1               AssetKey = "money-1"
	AssetKeyDealBreaker          AssetKey = "deal-breaker"
	AssetKeyJustSayNo            AssetKey = "just-say-no"
	AssetKeyHotel                AssetKey = "hotel"
	AssetKeyDebtCollector        AssetKey = "debt-collector"
	AssetKeyForcedDeal           AssetKey = "forced-deal"
	AssetKeySlyDeal              AssetKey = "sly-deal"
	AssetKeyHouse                AssetKey = "house"
	AssetKeyItsMyBirthday        AssetKey = "its-my-birthday"
	AssetKeyDoubleTheRent        AssetKey = "double-the-rent"
	AssetKeyPassGo               AssetKey = "pass-go"
	AssetKeyRentWild             AssetKey = "rent-wild"
	AssetKeyRentBrownSky         AssetKey = "rent-brown-sky"
	AssetKeyRentPinkOrange       AssetKey = "rent-pink-orange"
	AssetKeyRentRedYellow        AssetKey = "rent-red-yellow"
	AssetKeyRentGreenBlue        AssetKey = "rent-green-blue"
	AssetKeyRentUtilityRailroad  AssetKey = "rent-utility-railroad"
)

//type AssetKey int
//
//const (
//	AssetKeyBalticAve AssetKey = iota
//	AssetKeyMediterraneanAve
//	AssetKeyConnecticutAve
//	AssetKeyOrientalAve
//	AssetKeyVermontAve
//	AssetKeyStCharlesPlace
//	AssetKeyVirginiaAve
//	AssetKeyStateAve
//	AssetKeyNewYorkAve
//	AssetKeyStJamesPlace
//	AssetKeyTennesseeAve
//	AssetKeyKentuckyAve
//	AssetKeyIndianaAve
//	AssetKeyIllinoisAve
//	AssetKeyVentnorAve
//	AssetKeyMarvinGardens
//	AssetKeyAtlanticAve
//	AssetKeyNorthCarolinaAve
//	AssetKeyPacificAve
//	AssetKeyPennsylvaniaAve
//	AssetKeyBoardwalk
//	AssetKeyParkPlace
//	AssetKeyWaterWorks
//	AssetKeyElectricCompany
//	AssetKeyShortLine
//	AssetKeyBandORailRoad
//	AssetKeyReadingRailroad
//	AssetKeyPennsylvaniaRailroad
//	AssetKeyWildBrownSky
//	AssetKeyWildSkyRailroad
//	AssetKeyWildPinkOrange
//	AssetKeyWildRedYellow
//	AssetKeyWildGreenBlue
//	AssetKeyWildGreenRailroad
//	AssetKeyWildUtilityRailroad
//	AssetKeyWildWild
//	AssetKeyMoney10
//	AssetKeyMoney5
//	AssetKeyMoney4
//	AssetKeyMoney3
//	AssetKeyMoney2
//	AssetKeyMoney1
//	AssetKeyDealBreaker
//	AssetKeyJustSayNo
//	AssetKeyHotel
//	AssetKeyDebtCollector
//	AssetKeyForcedDeal
//	AssetKeySlyDeal
//	AssetKeyHouse
//	AssetKeyItsMyBirthday
//	AssetKeyDoubleTheRent
//	AssetKeyPassGo
//	AssetKeyRentWild
//	AssetKeyRentBrownSky
//	AssetKeyRentPinkOrange
//	AssetKeyRentRedYellow
//	AssetKeyRentGreenBlue
//	AssetKeyRentUtilityRailroad
//)

type Category int

const (
	CategoryPureProperty Category = iota
	CategoryWildProperty
	CategoryMoney
	CategoryAction
)

type Color int

const (
	ColorNone Color = iota
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

type Card struct {
	ID          Identifier `json:"id"`
	Category    Category   `json:"category"`
	AssetKey    AssetKey   `json:"card_key"`
	Value       int        `json:"value"`
	Colors      []Color    `json:"colors"`
	ActiveColor Color      `json:"active_color"`
}

func NewCard(id Identifier, c Category, ak AssetKey, v int, colors ...Color) Card {
	slices.Sort(colors)

	activeColor := ColorNone
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

// DO NOT USE THIS MAP DO FETCH CARDS!
// Cards are NOT unique here.
var CardByAssetKey = map[AssetKey]Card{
	AssetKeyBalticAve:            NewCard("", CategoryPureProperty, AssetKeyBalticAve, 1, ColorBrown),
	AssetKeyMediterraneanAve:     NewCard("", CategoryPureProperty, AssetKeyMediterraneanAve, 1, ColorBrown),
	AssetKeyConnecticutAve:       NewCard("", CategoryPureProperty, AssetKeyConnecticutAve, 1, ColorSky),
	AssetKeyOrientalAve:          NewCard("", CategoryPureProperty, AssetKeyOrientalAve, 1, ColorSky),
	AssetKeyVermontAve:           NewCard("", CategoryPureProperty, AssetKeyVermontAve, 1, ColorSky),
	AssetKeyStCharlesPlace:       NewCard("", CategoryPureProperty, AssetKeyStCharlesPlace, 2, ColorPink),
	AssetKeyVirginiaAve:          NewCard("", CategoryPureProperty, AssetKeyVirginiaAve, 2, ColorPink),
	AssetKeyStateAve:             NewCard("", CategoryPureProperty, AssetKeyStateAve, 2, ColorPink),
	AssetKeyNewYorkAve:           NewCard("", CategoryPureProperty, AssetKeyNewYorkAve, 2, ColorOrange),
	AssetKeyStJamesPlace:         NewCard("", CategoryPureProperty, AssetKeyStJamesPlace, 2, ColorOrange),
	AssetKeyTennesseeAve:         NewCard("", CategoryPureProperty, AssetKeyTennesseeAve, 2, ColorOrange),
	AssetKeyKentuckyAve:          NewCard("", CategoryPureProperty, AssetKeyKentuckyAve, 3, ColorRed),
	AssetKeyIndianaAve:           NewCard("", CategoryPureProperty, AssetKeyIndianaAve, 3, ColorRed),
	AssetKeyIllinoisAve:          NewCard("", CategoryPureProperty, AssetKeyIllinoisAve, 3, ColorRed),
	AssetKeyVentnorAve:           NewCard("", CategoryPureProperty, AssetKeyVentnorAve, 3, ColorYellow),
	AssetKeyMarvinGardens:        NewCard("", CategoryPureProperty, AssetKeyMarvinGardens, 3, ColorYellow),
	AssetKeyAtlanticAve:          NewCard("", CategoryPureProperty, AssetKeyAtlanticAve, 3, ColorYellow),
	AssetKeyNorthCarolinaAve:     NewCard("", CategoryPureProperty, AssetKeyNorthCarolinaAve, 4, ColorGreen),
	AssetKeyPacificAve:           NewCard("", CategoryPureProperty, AssetKeyPacificAve, 4, ColorGreen),
	AssetKeyPennsylvaniaAve:      NewCard("", CategoryPureProperty, AssetKeyPennsylvaniaAve, 4, ColorGreen),
	AssetKeyBoardwalk:            NewCard("", CategoryPureProperty, AssetKeyBoardwalk, 4, ColorBlue),
	AssetKeyParkPlace:            NewCard("", CategoryPureProperty, AssetKeyParkPlace, 4, ColorBlue),
	AssetKeyWaterWorks:           NewCard("", CategoryPureProperty, AssetKeyWaterWorks, 2, ColorUtility),
	AssetKeyElectricCompany:      NewCard("", CategoryPureProperty, AssetKeyElectricCompany, 2, ColorUtility),
	AssetKeyShortLine:            NewCard("", CategoryPureProperty, AssetKeyShortLine, 2, ColorRailroad),
	AssetKeyBandORailRoad:        NewCard("", CategoryPureProperty, AssetKeyBandORailRoad, 2, ColorRailroad),
	AssetKeyReadingRailroad:      NewCard("", CategoryPureProperty, AssetKeyReadingRailroad, 2, ColorRailroad),
	AssetKeyPennsylvaniaRailroad: NewCard("", CategoryPureProperty, AssetKeyPennsylvaniaRailroad, 2, ColorRailroad),

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

	AssetKeyPassGo:              NewCard("", CategoryAction, AssetKeyPassGo, 1),
	AssetKeyDoubleTheRent:       NewCard("", CategoryAction, AssetKeyDoubleTheRent, 1),
	AssetKeyItsMyBirthday:       NewCard("", CategoryAction, AssetKeyItsMyBirthday, 2),
	AssetKeyHouse:               NewCard("", CategoryAction, AssetKeyHouse, 3),
	AssetKeySlyDeal:             NewCard("", CategoryAction, AssetKeySlyDeal, 3),
	AssetKeyForcedDeal:          NewCard("", CategoryAction, AssetKeyForcedDeal, 3),
	AssetKeyDebtCollector:       NewCard("", CategoryAction, AssetKeyDebtCollector, 3),
	AssetKeyHotel:               NewCard("", CategoryAction, AssetKeyHotel, 4),
	AssetKeyJustSayNo:           NewCard("", CategoryAction, AssetKeyJustSayNo, 4),
	AssetKeyDealBreaker:         NewCard("", CategoryAction, AssetKeyDealBreaker, 5),
	AssetKeyRentBrownSky:        NewCard("", CategoryAction, AssetKeyRentBrownSky, 1),
	AssetKeyRentPinkOrange:      NewCard("", CategoryAction, AssetKeyRentPinkOrange, 1),
	AssetKeyRentRedYellow:       NewCard("", CategoryAction, AssetKeyRentRedYellow, 1),
	AssetKeyRentGreenBlue:       NewCard("", CategoryAction, AssetKeyRentGreenBlue, 1),
	AssetKeyRentUtilityRailroad: NewCard("", CategoryAction, AssetKeyRentUtilityRailroad, 1),
	AssetKeyRentWild:            NewCard("", CategoryAction, AssetKeyRentWild, 3),
}

func (c Card) String() string {
	return fmt.Sprintf("%s:%s", c.ID, c.AssetKey)
}

func (c Card) HasColor(color Color) bool {
	_, ok := slices.BinarySearch(c.Colors, color)
	return ok
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
