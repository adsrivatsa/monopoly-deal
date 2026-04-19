package monopoly_deal

import (
	"fmt"
	"fun-kames/internal/schema/monopoly_deal_schema"
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

var AssetKeyProtoMap = map[AssetKey]monopoly_deal_schema.AssetKey{
	AssetKeyBalticAve:            monopoly_deal_schema.AssetKey_BALTIC_AVE,
	AssetKeyMediterraneanAve:     monopoly_deal_schema.AssetKey_MEDITERRANEAN_AVE,
	AssetKeyConnecticutAve:       monopoly_deal_schema.AssetKey_CONNECTICUT_AVE,
	AssetKeyOrientalAve:          monopoly_deal_schema.AssetKey_ORIENTAL_AVE,
	AssetKeyVermontAve:           monopoly_deal_schema.AssetKey_VERMONT_AVE,
	AssetKeyStCharlesPlace:       monopoly_deal_schema.AssetKey_ST_CHARLES_PLACE,
	AssetKeyVirginiaAve:          monopoly_deal_schema.AssetKey_VIRGINIA_AVE,
	AssetKeyStateAve:             monopoly_deal_schema.AssetKey_STATE_AVE,
	AssetKeyNewYorkAve:           monopoly_deal_schema.AssetKey_NEW_YORK_AVE,
	AssetKeyStJamesPlace:         monopoly_deal_schema.AssetKey_ST_JAMES_PLACE,
	AssetKeyTennesseeAve:         monopoly_deal_schema.AssetKey_TENNESSEE_AVE,
	AssetKeyKentuckyAve:          monopoly_deal_schema.AssetKey_KENTUCKY_AVE,
	AssetKeyIndianaAve:           monopoly_deal_schema.AssetKey_INDIANA_AVE,
	AssetKeyIllinoisAve:          monopoly_deal_schema.AssetKey_ILLINOIS_AVE,
	AssetKeyVentnorAve:           monopoly_deal_schema.AssetKey_VENTNOR_AVE,
	AssetKeyMarvinGardens:        monopoly_deal_schema.AssetKey_MARVIN_GARDENS,
	AssetKeyAtlanticAve:          monopoly_deal_schema.AssetKey_ATLANTIC_AVE,
	AssetKeyNorthCarolinaAve:     monopoly_deal_schema.AssetKey_NORTH_CAROLINA_AVE,
	AssetKeyPacificAve:           monopoly_deal_schema.AssetKey_PACIFIC_AVE,
	AssetKeyPennsylvaniaAve:      monopoly_deal_schema.AssetKey_PENNSYLVANIA_AVE,
	AssetKeyBoardwalk:            monopoly_deal_schema.AssetKey_BOARDWALK,
	AssetKeyParkPlace:            monopoly_deal_schema.AssetKey_PARK_PLACE,
	AssetKeyWaterWorks:           monopoly_deal_schema.AssetKey_WATER_WORKS,
	AssetKeyElectricCompany:      monopoly_deal_schema.AssetKey_ELECTRIC_COMPANY,
	AssetKeyShortLine:            monopoly_deal_schema.AssetKey_SHORT_LINE,
	AssetKeyBandORailRoad:        monopoly_deal_schema.AssetKey_B_AND_O_RAILROAD,
	AssetKeyReadingRailroad:      monopoly_deal_schema.AssetKey_READING_RAILROAD,
	AssetKeyPennsylvaniaRailroad: monopoly_deal_schema.AssetKey_PENNSYLVANIA_RAILROAD,
	AssetKeyWildBrownSky:         monopoly_deal_schema.AssetKey_WILD_BROWN_SKY,
	AssetKeyWildSkyRailroad:      monopoly_deal_schema.AssetKey_WILD_SKY_RAILROAD,
	AssetKeyWildPinkOrange:       monopoly_deal_schema.AssetKey_WILD_PINK_ORANGE,
	AssetKeyWildRedYellow:        monopoly_deal_schema.AssetKey_WILD_RED_YELLOW,
	AssetKeyWildGreenBlue:        monopoly_deal_schema.AssetKey_WILD_GREEN_BLUE,
	AssetKeyWildGreenRailroad:    monopoly_deal_schema.AssetKey_WILD_GREEN_RAILROAD,
	AssetKeyWildUtilityRailroad:  monopoly_deal_schema.AssetKey_WILD_UTILITY_RAILROAD,
	AssetKeyWildWild:             monopoly_deal_schema.AssetKey_WILD_WILD,
	AssetKeyMoney10:              monopoly_deal_schema.AssetKey_MONEY_10,
	AssetKeyMoney5:               monopoly_deal_schema.AssetKey_MONEY_5,
	AssetKeyMoney4:               monopoly_deal_schema.AssetKey_MONEY_4,
	AssetKeyMoney3:               monopoly_deal_schema.AssetKey_MONEY_3,
	AssetKeyMoney2:               monopoly_deal_schema.AssetKey_MONEY_2,
	AssetKeyMoney1:               monopoly_deal_schema.AssetKey_MONEY_1,
	AssetKeyDealBreaker:          monopoly_deal_schema.AssetKey_DEAL_BREAKER,
	AssetKeyJustSayNo:            monopoly_deal_schema.AssetKey_JUST_SAY_NO,
	AssetKeyHotel:                monopoly_deal_schema.AssetKey_HOTEL,
	AssetKeyDebtCollector:        monopoly_deal_schema.AssetKey_DEBT_COLLECTOR,
	AssetKeyForcedDeal:           monopoly_deal_schema.AssetKey_FORCED_DEAL,
	AssetKeySlyDeal:              monopoly_deal_schema.AssetKey_SLY_DEAL,
	AssetKeyHouse:                monopoly_deal_schema.AssetKey_HOUSE,
	AssetKeyItsMyBirthday:        monopoly_deal_schema.AssetKey_ITS_MY_BIRTHDAY,
	AssetKeyDoubleTheRent:        monopoly_deal_schema.AssetKey_DOUBLE_THE_RENT,
	AssetKeyPassGo:               monopoly_deal_schema.AssetKey_PASS_GO,
	AssetKeyRentWild:             monopoly_deal_schema.AssetKey_RENT_WILD,
	AssetKeyRentBrownSky:         monopoly_deal_schema.AssetKey_RENT_BROWN_SKY,
	AssetKeyRentPinkOrange:       monopoly_deal_schema.AssetKey_RENT_PINK_ORANGE,
	AssetKeyRentRedYellow:        monopoly_deal_schema.AssetKey_RENT_RED_YELLOW,
	AssetKeyRentGreenBlue:        monopoly_deal_schema.AssetKey_RENT_GREEN_BLUE,
	AssetKeyRentUtilityRailroad:  monopoly_deal_schema.AssetKey_RENT_UTILITY_RAILROAD,
}

func (a AssetKey) Proto() monopoly_deal_schema.AssetKey {
	return AssetKeyProtoMap[a]
}

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

var CategoryProtoMap = map[Category]monopoly_deal_schema.Category{
	CategoryPureProperty: monopoly_deal_schema.Category_PURE_PROPERTY,
	CategoryWildProperty: monopoly_deal_schema.Category_WILD_PROPERTY,
	CategoryMoney:        monopoly_deal_schema.Category_MONEY,
	CategoryAction:       monopoly_deal_schema.Category_ACTION,
}

func (c Category) Proto() monopoly_deal_schema.Category {
	return CategoryProtoMap[c]
}

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

var ColorProtoMap = map[Color]monopoly_deal_schema.Color{
	ColorNone:     monopoly_deal_schema.Color_NONE,
	ColorBrown:    monopoly_deal_schema.Color_BROWN,
	ColorSky:      monopoly_deal_schema.Color_SKY,
	ColorPink:     monopoly_deal_schema.Color_PINK,
	ColorOrange:   monopoly_deal_schema.Color_ORANGE,
	ColorRed:      monopoly_deal_schema.Color_RED,
	ColorYellow:   monopoly_deal_schema.Color_YELLOW,
	ColorGreen:    monopoly_deal_schema.Color_GREEN,
	ColorBlue:     monopoly_deal_schema.Color_BLUE,
	ColorUtility:  monopoly_deal_schema.Color_UTILITY,
	ColorRailroad: monopoly_deal_schema.Color_RAILROAD,
}

func (c Color) Proto() monopoly_deal_schema.Color {
	return ColorProtoMap[c]
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

func (c Card) Proto() *monopoly_deal_schema.Card {
	colors := make([]monopoly_deal_schema.Color, len(c.Colors))
	for i, color := range c.Colors {
		colors[i] = color.Proto()
	}

	return &monopoly_deal_schema.Card{
		CardId:      string(c.ID),
		AssetKey:    c.AssetKey.Proto(),
		Category:    c.Category.Proto(),
		ActiveColor: c.ActiveColor.Proto(),
		Colors:      colors,
		Value:       int32(c.Value),
	}
}

// DO NOT USE THIS MAP TO FETCH CARDS!
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

func (c *Cards) Proto() []*monopoly_deal_schema.Card {
	cards := make([]*monopoly_deal_schema.Card, len(*c))
	for i, card := range *c {
		cards[i] = card.Proto()
	}
	return cards
}

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
