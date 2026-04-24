package monopoly_deal

import (
	"fmt"
	"the-deal/internal/schema/monopoly_deal_schema"
)

type AssetKey string

const (
	AssetKeyUnspecified         AssetKey = ""
	AssetKeyBrown1              AssetKey = "brown1"
	AssetKeyBrown2              AssetKey = "brown2"
	AssetKeySky1                AssetKey = "sky1"
	AssetKeySky2                AssetKey = "sky2"
	AssetKeySky3                AssetKey = "sky3"
	AssetKeyPink1               AssetKey = "pink1"
	AssetKeyPink2               AssetKey = "pink2"
	AssetKeyPink3               AssetKey = "pink3"
	AssetKeyOrange1             AssetKey = "orange1"
	AssetKeyOrange2             AssetKey = "orange2"
	AssetKeyOrange3             AssetKey = "orange3"
	AssetKeyRed1                AssetKey = "red1"
	AssetKeyRed2                AssetKey = "red2"
	AssetKeyRed3                AssetKey = "red3"
	AssetKeyYellow1             AssetKey = "yellow1"
	AssetKeyYellow2             AssetKey = "yellow2"
	AssetKeyYellow3             AssetKey = "yellow3"
	AssetKeyGreen1              AssetKey = "green1"
	AssetKeyGreen2              AssetKey = "green2"
	AssetKeyGreen3              AssetKey = "green3"
	AssetKeyBlue1               AssetKey = "blue1"
	AssetKeyBlue2               AssetKey = "blue2"
	AssetKeyUtil1               AssetKey = "util1"
	AssetKeyUtil2               AssetKey = "util2"
	AssetKeyRail1               AssetKey = "rail1"
	AssetKeyRail2               AssetKey = "rail2"
	AssetKeyRail3               AssetKey = "rail3"
	AssetKeyRail4               AssetKey = "rail4"
	AssetKeyWildBrownSky        AssetKey = "wild_brown_sky"
	AssetKeyWildSkyRailroad     AssetKey = "wild_sky_railroad"
	AssetKeyWildPinkOrange      AssetKey = "wild_pink_orange"
	AssetKeyWildRedYellow       AssetKey = "wild_red_yellow"
	AssetKeyWildGreenBlue       AssetKey = "wild_green_blue"
	AssetKeyWildGreenRailroad   AssetKey = "wild_green_railroad"
	AssetKeyWildUtilityRailroad AssetKey = "wild_utility_railroad"
	AssetKeyWildWild            AssetKey = "wild_wild"
	AssetKeyMoney10             AssetKey = "money10"
	AssetKeyMoney5              AssetKey = "money5"
	AssetKeyMoney4              AssetKey = "money4"
	AssetKeyMoney3              AssetKey = "money3"
	AssetKeyMoney2              AssetKey = "money2"
	AssetKeyMoney1              AssetKey = "money1"
	AssetKeyDealBreaker         AssetKey = "deal_breaker"
	AssetKeyJustSayNo           AssetKey = "just_say_no"
	AssetKeyHotel               AssetKey = "hotel"
	AssetKeyDebtCollector       AssetKey = "debt_collector"
	AssetKeyForcedDeal          AssetKey = "forced_deal"
	AssetKeySlyDeal             AssetKey = "sly_deal"
	AssetKeyHouse               AssetKey = "house"
	AssetKeyItsMyBirthday       AssetKey = "its_my_birthday"
	AssetKeyDoubleTheRent       AssetKey = "double_the_rent"
	AssetKeyPassGo              AssetKey = "pass_go"
	AssetKeyRentWild            AssetKey = "rent_wild"
	AssetKeyRentBrownSky        AssetKey = "rent_brown_sky"
	AssetKeyRentPinkOrange      AssetKey = "rent_pink_orange"
	AssetKeyRentRedYellow       AssetKey = "rent_red_yellow"
	AssetKeyRentGreenBlue       AssetKey = "rent_green_blue"
	AssetKeyRentUtilityRailroad AssetKey = "rent_utility_railroad"
)

var AssetKeyProtoMap = map[AssetKey]monopoly_deal_schema.AssetKey{
	AssetKeyUnspecified:         monopoly_deal_schema.AssetKey_ASSET_KEY_UNSPECIFIED,
	AssetKeyBrown1:              monopoly_deal_schema.AssetKey_ASSET_KEY_BROWN1,
	AssetKeyBrown2:              monopoly_deal_schema.AssetKey_ASSET_KEY_BROWN2,
	AssetKeySky1:                monopoly_deal_schema.AssetKey_ASSET_KEY_SKY1,
	AssetKeySky2:                monopoly_deal_schema.AssetKey_ASSET_KEY_SKY2,
	AssetKeySky3:                monopoly_deal_schema.AssetKey_ASSET_KEY_SKY3,
	AssetKeyPink1:               monopoly_deal_schema.AssetKey_ASSET_KEY_PINK1,
	AssetKeyPink2:               monopoly_deal_schema.AssetKey_ASSET_KEY_PINK2,
	AssetKeyPink3:               monopoly_deal_schema.AssetKey_ASSET_KEY_PINK3,
	AssetKeyOrange1:             monopoly_deal_schema.AssetKey_ASSET_KEY_ORANGE1,
	AssetKeyOrange2:             monopoly_deal_schema.AssetKey_ASSET_KEY_ORANGE2,
	AssetKeyOrange3:             monopoly_deal_schema.AssetKey_ASSET_KEY_ORANGE3,
	AssetKeyRed1:                monopoly_deal_schema.AssetKey_ASSET_KEY_RED1,
	AssetKeyRed2:                monopoly_deal_schema.AssetKey_ASSET_KEY_RED2,
	AssetKeyRed3:                monopoly_deal_schema.AssetKey_ASSET_KEY_RED3,
	AssetKeyYellow1:             monopoly_deal_schema.AssetKey_ASSET_KEY_YELLOW1,
	AssetKeyYellow2:             monopoly_deal_schema.AssetKey_ASSET_KEY_YELLOW2,
	AssetKeyYellow3:             monopoly_deal_schema.AssetKey_ASSET_KEY_YELLOW3,
	AssetKeyGreen1:              monopoly_deal_schema.AssetKey_ASSET_KEY_GREEN1,
	AssetKeyGreen2:              monopoly_deal_schema.AssetKey_ASSET_KEY_GREEN2,
	AssetKeyGreen3:              monopoly_deal_schema.AssetKey_ASSET_KEY_GREEN3,
	AssetKeyBlue1:               monopoly_deal_schema.AssetKey_ASSET_KEY_BLUE1,
	AssetKeyBlue2:               monopoly_deal_schema.AssetKey_ASSET_KEY_BLUE2,
	AssetKeyUtil1:               monopoly_deal_schema.AssetKey_ASSET_KEY_UTIL1,
	AssetKeyUtil2:               monopoly_deal_schema.AssetKey_ASSET_KEY_UTIL2,
	AssetKeyRail1:               monopoly_deal_schema.AssetKey_ASSET_KEY_RAIL1,
	AssetKeyRail2:               monopoly_deal_schema.AssetKey_ASSET_KEY_RAIL2,
	AssetKeyRail3:               monopoly_deal_schema.AssetKey_ASSET_KEY_RAIL3,
	AssetKeyRail4:               monopoly_deal_schema.AssetKey_ASSET_KEY_RAIL4,
	AssetKeyWildBrownSky:        monopoly_deal_schema.AssetKey_ASSET_KEY_WILD_BROWN_SKY,
	AssetKeyWildSkyRailroad:     monopoly_deal_schema.AssetKey_ASSET_KEY_WILD_SKY_RAILROAD,
	AssetKeyWildPinkOrange:      monopoly_deal_schema.AssetKey_ASSET_KEY_WILD_PINK_ORANGE,
	AssetKeyWildRedYellow:       monopoly_deal_schema.AssetKey_ASSET_KEY_WILD_RED_YELLOW,
	AssetKeyWildGreenBlue:       monopoly_deal_schema.AssetKey_ASSET_KEY_WILD_GREEN_BLUE,
	AssetKeyWildGreenRailroad:   monopoly_deal_schema.AssetKey_ASSET_KEY_WILD_GREEN_RAILROAD,
	AssetKeyWildUtilityRailroad: monopoly_deal_schema.AssetKey_ASSET_KEY_WILD_UTILITY_RAILROAD,
	AssetKeyWildWild:            monopoly_deal_schema.AssetKey_ASSET_KEY_WILD_WILD,
	AssetKeyMoney10:             monopoly_deal_schema.AssetKey_ASSET_KEY_MONEY_10,
	AssetKeyMoney5:              monopoly_deal_schema.AssetKey_ASSET_KEY_MONEY_5,
	AssetKeyMoney4:              monopoly_deal_schema.AssetKey_ASSET_KEY_MONEY_4,
	AssetKeyMoney3:              monopoly_deal_schema.AssetKey_ASSET_KEY_MONEY_3,
	AssetKeyMoney2:              monopoly_deal_schema.AssetKey_ASSET_KEY_MONEY_2,
	AssetKeyMoney1:              monopoly_deal_schema.AssetKey_ASSET_KEY_MONEY_1,
	AssetKeyDealBreaker:         monopoly_deal_schema.AssetKey_ASSET_KEY_DEAL_BREAKER,
	AssetKeyJustSayNo:           monopoly_deal_schema.AssetKey_ASSET_KEY_JUST_SAY_NO,
	AssetKeyHotel:               monopoly_deal_schema.AssetKey_ASSET_KEY_HOTEL,
	AssetKeyDebtCollector:       monopoly_deal_schema.AssetKey_ASSET_KEY_DEBT_COLLECTOR,
	AssetKeyForcedDeal:          monopoly_deal_schema.AssetKey_ASSET_KEY_FORCED_DEAL,
	AssetKeySlyDeal:             monopoly_deal_schema.AssetKey_ASSET_KEY_SLY_DEAL,
	AssetKeyHouse:               monopoly_deal_schema.AssetKey_ASSET_KEY_HOUSE,
	AssetKeyItsMyBirthday:       monopoly_deal_schema.AssetKey_ASSET_KEY_ITS_MY_BIRTHDAY,
	AssetKeyDoubleTheRent:       monopoly_deal_schema.AssetKey_ASSET_KEY_DOUBLE_THE_RENT,
	AssetKeyPassGo:              monopoly_deal_schema.AssetKey_ASSET_KEY_PASS_GO,
	AssetKeyRentWild:            monopoly_deal_schema.AssetKey_ASSET_KEY_RENT_WILD,
	AssetKeyRentBrownSky:        monopoly_deal_schema.AssetKey_ASSET_KEY_RENT_BROWN_SKY,
	AssetKeyRentPinkOrange:      monopoly_deal_schema.AssetKey_ASSET_KEY_RENT_PINK_ORANGE,
	AssetKeyRentRedYellow:       monopoly_deal_schema.AssetKey_ASSET_KEY_RENT_RED_YELLOW,
	AssetKeyRentGreenBlue:       monopoly_deal_schema.AssetKey_ASSET_KEY_RENT_GREEN_BLUE,
	AssetKeyRentUtilityRailroad: monopoly_deal_schema.AssetKey_ASSET_KEY_RENT_UTILITY_RAILROAD,
}

func AllAssetKeys() []AssetKey {
	return []AssetKey{AssetKeyBrown1, AssetKeyBrown2, AssetKeySky1, AssetKeySky2, AssetKeySky3, AssetKeyPink1, AssetKeyPink2, AssetKeyPink3, AssetKeyOrange1, AssetKeyOrange2, AssetKeyOrange3, AssetKeyRed1, AssetKeyRed2, AssetKeyRed3, AssetKeyYellow1, AssetKeyYellow2, AssetKeyYellow3, AssetKeyGreen1, AssetKeyGreen2, AssetKeyGreen3, AssetKeyBlue1, AssetKeyBlue2, AssetKeyUtil1, AssetKeyUtil2, AssetKeyRail1, AssetKeyRail2, AssetKeyRail3, AssetKeyRail4, AssetKeyWildBrownSky, AssetKeyWildSkyRailroad, AssetKeyWildPinkOrange, AssetKeyWildRedYellow, AssetKeyWildGreenBlue, AssetKeyWildGreenRailroad, AssetKeyWildUtilityRailroad, AssetKeyWildWild, AssetKeyMoney10, AssetKeyMoney5, AssetKeyMoney4, AssetKeyMoney3, AssetKeyMoney2, AssetKeyMoney1, AssetKeyDealBreaker, AssetKeyJustSayNo, AssetKeyHotel, AssetKeyDebtCollector, AssetKeyForcedDeal, AssetKeySlyDeal, AssetKeyHouse, AssetKeyItsMyBirthday, AssetKeyDoubleTheRent, AssetKeyPassGo, AssetKeyRentWild, AssetKeyRentBrownSky, AssetKeyRentPinkOrange, AssetKeyRentRedYellow, AssetKeyRentGreenBlue, AssetKeyRentUtilityRailroad}
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
	CategoryUnspecified Category = iota
	CategoryPureProperty
	CategoryWildProperty
	CategoryMoney
	CategoryAction
)

var CategoryProtoMap = map[Category]monopoly_deal_schema.Category{
	CategoryUnspecified:  monopoly_deal_schema.Category_CATEGORY_UNSPECIFIED,
	CategoryPureProperty: monopoly_deal_schema.Category_CATEGORY_PURE_PROPERTY,
	CategoryWildProperty: monopoly_deal_schema.Category_CATEGORY_WILD_PROPERTY,
	CategoryMoney:        monopoly_deal_schema.Category_CATEGORY_MONEY,
	CategoryAction:       monopoly_deal_schema.Category_CATEGORY_ACTION,
}

func (c Category) Proto() monopoly_deal_schema.Category {
	return CategoryProtoMap[c]
}

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

var ProtoColorMap = map[monopoly_deal_schema.Color]Color{
	monopoly_deal_schema.Color_COLOR_UNSPECIFIED: ColorUnspecified,
	monopoly_deal_schema.Color_COLOR_BROWN:       ColorBrown,
	monopoly_deal_schema.Color_COLOR_SKY:         ColorSky,
	monopoly_deal_schema.Color_COLOR_PINK:        ColorPink,
	monopoly_deal_schema.Color_COLOR_ORANGE:      ColorOrange,
	monopoly_deal_schema.Color_COLOR_RED:         ColorRed,
	monopoly_deal_schema.Color_COLOR_YELLOW:      ColorYellow,
	monopoly_deal_schema.Color_COLOR_GREEN:       ColorGreen,
	monopoly_deal_schema.Color_COLOR_BLUE:        ColorBlue,
	monopoly_deal_schema.Color_COLOR_UTILITY:     ColorUtility,
	monopoly_deal_schema.Color_COLOR_RAILROAD:    ColorRailroad,
}

func ColorFromProto(c monopoly_deal_schema.Color) Color {
	return ProtoColorMap[c]
}

var ColorProtoMap = map[Color]monopoly_deal_schema.Color{
	ColorUnspecified: monopoly_deal_schema.Color_COLOR_UNSPECIFIED,
	ColorBrown:       monopoly_deal_schema.Color_COLOR_BROWN,
	ColorSky:         monopoly_deal_schema.Color_COLOR_SKY,
	ColorPink:        monopoly_deal_schema.Color_COLOR_PINK,
	ColorOrange:      monopoly_deal_schema.Color_COLOR_ORANGE,
	ColorRed:         monopoly_deal_schema.Color_COLOR_RED,
	ColorYellow:      monopoly_deal_schema.Color_COLOR_YELLOW,
	ColorGreen:       monopoly_deal_schema.Color_COLOR_GREEN,
	ColorBlue:        monopoly_deal_schema.Color_COLOR_BLUE,
	ColorUtility:     monopoly_deal_schema.Color_COLOR_UTILITY,
	ColorRailroad:    monopoly_deal_schema.Color_COLOR_RAILROAD,
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
	for _, col := range c.Colors {
		if col == color {
			return true
		}
	}
	return false
}

type Cards []Card

func (c *Cards) Proto() []*monopoly_deal_schema.Card {
	if c == nil {
		return nil
	}

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

func (c *Cards) Value() int {
	value := 0
	for _, card := range *c {
		value += card.Value
	}
	return value
}
