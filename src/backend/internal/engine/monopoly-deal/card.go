package monopoly_deal

type CardKey string

const (
	CardKeyBalticAve            CardKey = "baltic-ave"
	CardKeyMediterraneanAve     CardKey = "mediterranean-ave"
	CardKeyConnecticutAve       CardKey = "connecticut-ave"
	CardKeyOrientalAve          CardKey = "oriental-ave"
	CardKeyVermontAve           CardKey = "vermont-ave"
	CardKeyStCharlesPlace       CardKey = "st-charles-place"
	CardKeyVirginiaAve          CardKey = "virginia-ave"
	CardKeyStateAve             CardKey = "state-ave"
	CardKeyNewYorkAve           CardKey = "new-york-ave"
	CardKeyStJamesPlace         CardKey = "st-james-place"
	CardKeyTennesseeAve         CardKey = "tennessee-ave"
	CardKeyKentuckyAve          CardKey = "kentucky-ave"
	CardKeyIndianaAve           CardKey = "indiana-ave"
	CardKeyIllinoisAve          CardKey = "illinois-ave"
	CardKeyVentnorAve           CardKey = "ventnor-ave"
	CardKeyMarvinGardens        CardKey = "marvin-gardens"
	CardKeyAtlanticAve          CardKey = "atlantic-ave"
	CardKeyNorthCarolinaAve     CardKey = "north-carolina-ave"
	CardKeyPacificAve           CardKey = "pacific-ave"
	CardKeyPennsylvaniaAve      CardKey = "pennsylvania-ave"
	CardKeyBoardwalk            CardKey = "boardwalk"
	CardKeyParkPlace            CardKey = "park-place"
	CardKeyWaterWorks           CardKey = "water-works"
	CardKeyElectricCompany      CardKey = "electric-company"
	CardKeyShortLine            CardKey = "short-line"
	CardKeyBandORailRoad        CardKey = "b-and-o-railroad"
	CardKeyReadingRailroad      CardKey = "reading-railroad"
	CardKeyPennsylvaniaRailroad CardKey = "pennsylvania-railroad"
	CardKeyWildBrownSky         CardKey = "wild-brown-sky"
	CardKeyWildSkyRailroad      CardKey = "wild-sky-railroad"
	CardKeyWildPinkOrange       CardKey = "wild-pink-orange"
	CardKeyWildRedYellow        CardKey = "wild-red-yellow"
	CardKeyWildGreenBlue        CardKey = "wild-green-blue"
	CardKeyWildGreenRailroad    CardKey = "wild-green-railroad"
	CardKeyWildUtilityRailroad  CardKey = "wild-utility-railroad"
	CardKeyWildWild             CardKey = "wild-wild"
	CardKeyMoney10              CardKey = "money-10"
	CardKeyMoney5               CardKey = "money-5"
	CardKeyMoney4               CardKey = "money-4"
	CardKeyMoney3               CardKey = "money-3"
	CardKeyMoney2               CardKey = "money-2"
	CardKeyMoney1               CardKey = "money-1"
	CardKeyDealBreaker          CardKey = "deal-breaker"
	CardKeyJustSayNo            CardKey = "just-say-no"
	CardKeyHotel                CardKey = "hotel"
	CardKeyDebtCollector        CardKey = "debt-collector"
	CardKeyForcedDeal           CardKey = "forced-deal"
	CardKeySlyDeal              CardKey = "sly-deal"
	CardKeyHouse                CardKey = "house"
	CardKeyItsMyBirthday        CardKey = "its-my-birthday"
	CardKeyDoubleTheRent        CardKey = "double-the-rent"
	CardKeyPassGo               CardKey = "pass-go"
	CardKeyRentWild             CardKey = "rent-wild"
	CardKeyRentBrownSky         CardKey = "rent-brown-sky"
	CardKeyRentPinkOrange       CardKey = "rent-pink-orange"
	CardKeyRentRedYellow        CardKey = "rent-red-yellow"
	CardKeyRentGreenBlue        CardKey = "rent-green-blue"
	CardKeyRentUtilityRailroad  CardKey = "rent-utility-railroad"
)

//type CardKey int
//
//const (
//	CardKeyBalticAve CardKey = iota
//	CardKeyMediterraneanAve
//	CardKeyConnecticutAve
//	CardKeyOrientalAve
//	CardKeyVermontAve
//	CardKeyStCharlesPlace
//	CardKeyVirginiaAve
//	CardKeyStateAve
//	CardKeyNewYorkAve
//	CardKeyStJamesPlace
//	CardKeyTennesseeAve
//	CardKeyKentuckyAve
//	CardKeyIndianaAve
//	CardKeyIllinoisAve
//	CardKeyVentnorAve
//	CardKeyMarvinGardens
//	CardKeyAtlanticAve
//	CardKeyNorthCarolinaAve
//	CardKeyPacificAve
//	CardKeyPennsylvaniaAve
//	CardKeyBoardwalk
//	CardKeyParkPlace
//	CardKeyWaterWorks
//	CardKeyElectricCompany
//	CardKeyShortLine
//	CardKeyBandORailRoad
//	CardKeyReadingRailroad
//	CardKeyPennsylvaniaRailroad
//	CardKeyWildBrownSky
//	CardKeyWildSkyRailroad
//	CardKeyWildPinkOrange
//	CardKeyWildRedYellow
//	CardKeyWildGreenBlue
//	CardKeyWildGreenRailroad
//	CardKeyWildUtilityRailroad
//	CardKeyWildWild
//	CardKeyMoney10
//	CardKeyMoney5
//	CardKeyMoney4
//	CardKeyMoney3
//	CardKeyMoney2
//	CardKeyMoney1
//	CardKeyDealBreaker
//	CardKeyJustSayNo
//	CardKeyHotel
//	CardKeyDebtCollector
//	CardKeyForcedDeal
//	CardKeySlyDeal
//	CardKeyHouse
//	CardKeyItsMyBirthday
//	CardKeyDoubleTheRent
//	CardKeyPassGo
//	CardKeyRentWild
//	CardKeyRentBrownSky
//	CardKeyRentPinkOrange
//	CardKeyRentRedYellow
//	CardKeyRentGreenBlue
//	CardKeyRentUtilityRailroad
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
	Category    Category `json:"category"`
	CardKey     CardKey  `json:"card_key"`
	Type        int      `json:"type"`
	Value       int      `json:"value"`
	Colors      []Color  `json:"colors"`
	ActiveColor Color    `json:"active_color"`
}

func NewCard(c Category, ck CardKey, t int, v int, colors ...Color) Card {
	activeColor := ColorNone
	if len(colors) > 0 {
		activeColor = colors[0]
	}

	return Card{
		Category:    c,
		CardKey:     ck,
		Type:        t,
		Value:       v,
		Colors:      append([]Color(nil), colors...),
		ActiveColor: activeColor,
	}
}

var CardByKey = map[CardKey]Card{
	CardKeyBalticAve:            NewCard(CategoryPureProperty, CardKeyBalticAve, int(PurePropertyCardBrown), 1, ColorBrown),
	CardKeyMediterraneanAve:     NewCard(CategoryPureProperty, CardKeyMediterraneanAve, int(PurePropertyCardBrown), 1, ColorBrown),
	CardKeyConnecticutAve:       NewCard(CategoryPureProperty, CardKeyConnecticutAve, int(PurePropertyCardSky), 1, ColorSky),
	CardKeyOrientalAve:          NewCard(CategoryPureProperty, CardKeyOrientalAve, int(PurePropertyCardSky), 1, ColorSky),
	CardKeyVermontAve:           NewCard(CategoryPureProperty, CardKeyVermontAve, int(PurePropertyCardSky), 1, ColorSky),
	CardKeyStCharlesPlace:       NewCard(CategoryPureProperty, CardKeyStCharlesPlace, int(PurePropertyCardPink), 2, ColorPink),
	CardKeyVirginiaAve:          NewCard(CategoryPureProperty, CardKeyVirginiaAve, int(PurePropertyCardPink), 2, ColorPink),
	CardKeyStateAve:             NewCard(CategoryPureProperty, CardKeyStateAve, int(PurePropertyCardPink), 2, ColorPink),
	CardKeyNewYorkAve:           NewCard(CategoryPureProperty, CardKeyNewYorkAve, int(PurePropertyCardOrange), 2, ColorOrange),
	CardKeyStJamesPlace:         NewCard(CategoryPureProperty, CardKeyStJamesPlace, int(PurePropertyCardOrange), 2, ColorOrange),
	CardKeyTennesseeAve:         NewCard(CategoryPureProperty, CardKeyTennesseeAve, int(PurePropertyCardOrange), 2, ColorOrange),
	CardKeyKentuckyAve:          NewCard(CategoryPureProperty, CardKeyKentuckyAve, int(PurePropertyCardRed), 3, ColorRed),
	CardKeyIndianaAve:           NewCard(CategoryPureProperty, CardKeyIndianaAve, int(PurePropertyCardRed), 3, ColorRed),
	CardKeyIllinoisAve:          NewCard(CategoryPureProperty, CardKeyIllinoisAve, int(PurePropertyCardRed), 3, ColorRed),
	CardKeyVentnorAve:           NewCard(CategoryPureProperty, CardKeyVentnorAve, int(PurePropertyCardYellow), 3, ColorYellow),
	CardKeyMarvinGardens:        NewCard(CategoryPureProperty, CardKeyMarvinGardens, int(PurePropertyCardYellow), 3, ColorYellow),
	CardKeyAtlanticAve:          NewCard(CategoryPureProperty, CardKeyAtlanticAve, int(PurePropertyCardYellow), 3, ColorYellow),
	CardKeyNorthCarolinaAve:     NewCard(CategoryPureProperty, CardKeyNorthCarolinaAve, int(PurePropertyCardGreen), 4, ColorGreen),
	CardKeyPacificAve:           NewCard(CategoryPureProperty, CardKeyPacificAve, int(PurePropertyCardGreen), 4, ColorGreen),
	CardKeyPennsylvaniaAve:      NewCard(CategoryPureProperty, CardKeyPennsylvaniaAve, int(PurePropertyCardGreen), 4, ColorGreen),
	CardKeyBoardwalk:            NewCard(CategoryPureProperty, CardKeyBoardwalk, int(PurePropertyCardBlue), 4, ColorBlue),
	CardKeyParkPlace:            NewCard(CategoryPureProperty, CardKeyParkPlace, int(PurePropertyCardBlue), 4, ColorBlue),
	CardKeyWaterWorks:           NewCard(CategoryPureProperty, CardKeyWaterWorks, int(PurePropertyCardUtility), 2, ColorUtility),
	CardKeyElectricCompany:      NewCard(CategoryPureProperty, CardKeyElectricCompany, int(PurePropertyCardUtility), 2, ColorUtility),
	CardKeyShortLine:            NewCard(CategoryPureProperty, CardKeyShortLine, int(PurePropertyCardRailroad), 2, ColorRailroad),
	CardKeyBandORailRoad:        NewCard(CategoryPureProperty, CardKeyBandORailRoad, int(PurePropertyCardRailroad), 2, ColorRailroad),
	CardKeyReadingRailroad:      NewCard(CategoryPureProperty, CardKeyReadingRailroad, int(PurePropertyCardRailroad), 2, ColorRailroad),
	CardKeyPennsylvaniaRailroad: NewCard(CategoryPureProperty, CardKeyPennsylvaniaRailroad, int(PurePropertyCardRailroad), 2, ColorRailroad),

	CardKeyWildBrownSky:        NewCard(CategoryWildProperty, CardKeyWildBrownSky, int(WildPropertyCardBrownSky), 1, ColorBrown, ColorSky),
	CardKeyWildSkyRailroad:     NewCard(CategoryWildProperty, CardKeyWildSkyRailroad, int(WildPropertyCardSkyRailroad), 4, ColorSky, ColorRailroad),
	CardKeyWildPinkOrange:      NewCard(CategoryWildProperty, CardKeyWildPinkOrange, int(WildPropertyCardPinkOrange), 2, ColorPink, ColorOrange),
	CardKeyWildRedYellow:       NewCard(CategoryWildProperty, CardKeyWildRedYellow, int(WildPropertyCardRedYellow), 3, ColorRed, ColorYellow),
	CardKeyWildGreenBlue:       NewCard(CategoryWildProperty, CardKeyWildGreenBlue, int(WildPropertyCardGreenBlue), 4, ColorGreen, ColorBlue),
	CardKeyWildGreenRailroad:   NewCard(CategoryWildProperty, CardKeyWildGreenRailroad, int(WildPropertyCardGreenRailroad), 4, ColorGreen, ColorRailroad),
	CardKeyWildUtilityRailroad: NewCard(CategoryWildProperty, CardKeyWildUtilityRailroad, int(WildPropertyCardUtilityRailroad), 2, ColorUtility, ColorRailroad),
	CardKeyWildWild:            NewCard(CategoryWildProperty, CardKeyWildWild, int(WildPropertyCardWild), 0, ColorBrown, ColorSky, ColorPink, ColorOrange, ColorRed, ColorYellow, ColorGreen, ColorBlue, ColorUtility, ColorRailroad),

	CardKeyMoney1:  NewCard(CategoryMoney, CardKeyMoney1, int(MoneyCard1), 1),
	CardKeyMoney2:  NewCard(CategoryMoney, CardKeyMoney2, int(MoneyCard2), 2),
	CardKeyMoney3:  NewCard(CategoryMoney, CardKeyMoney3, int(MoneyCard3), 3),
	CardKeyMoney4:  NewCard(CategoryMoney, CardKeyMoney4, int(MoneyCard4), 4),
	CardKeyMoney5:  NewCard(CategoryMoney, CardKeyMoney5, int(MoneyCard5), 5),
	CardKeyMoney10: NewCard(CategoryMoney, CardKeyMoney10, int(MoneyCard10), 10),

	CardKeyPassGo:              NewCard(CategoryAction, CardKeyPassGo, int(ActionCardPassGo), 1),
	CardKeyDoubleTheRent:       NewCard(CategoryAction, CardKeyDoubleTheRent, int(ActionCardDoubleTheRent), 1),
	CardKeyItsMyBirthday:       NewCard(CategoryAction, CardKeyItsMyBirthday, int(ActionCardItsMyBirthday), 2),
	CardKeyHouse:               NewCard(CategoryAction, CardKeyHouse, int(ActionCardHouse), 3),
	CardKeySlyDeal:             NewCard(CategoryAction, CardKeySlyDeal, int(ActionCardSlyDeal), 3),
	CardKeyForcedDeal:          NewCard(CategoryAction, CardKeyForcedDeal, int(ActionCardForcedDeal), 3),
	CardKeyDebtCollector:       NewCard(CategoryAction, CardKeyDebtCollector, int(ActionCardDebtCollector), 3),
	CardKeyHotel:               NewCard(CategoryAction, CardKeyHotel, int(ActionCardHotel), 4),
	CardKeyJustSayNo:           NewCard(CategoryAction, CardKeyJustSayNo, int(ActionCardJustSayNo), 4),
	CardKeyDealBreaker:         NewCard(CategoryAction, CardKeyDealBreaker, int(ActionCardDealBreaker), 5),
	CardKeyRentBrownSky:        NewCard(CategoryAction, CardKeyRentBrownSky, int(ActionCardRentBrownSky), 1),
	CardKeyRentPinkOrange:      NewCard(CategoryAction, CardKeyRentPinkOrange, int(ActionCardRentPinkOrange), 1),
	CardKeyRentRedYellow:       NewCard(CategoryAction, CardKeyRentRedYellow, int(ActionCardRentRedYellow), 1),
	CardKeyRentGreenBlue:       NewCard(CategoryAction, CardKeyRentGreenBlue, int(ActionCardRentGreenBlue), 1),
	CardKeyRentUtilityRailroad: NewCard(CategoryAction, CardKeyRentUtilityRailroad, int(ActionCardRentUtilityRailroad), 1),
	CardKeyRentWild:            NewCard(CategoryAction, CardKeyRentWild, int(ActionCardRentWild), 3),
}

func (c Card) PureProperty() PurePropertyCardType {
	return PurePropertyCardType(c.Type)
}

func (c Card) WildProperty() WildPropertyCardType {
	return WildPropertyCardType(c.Type)
}

func (c Card) Money() MoneyCardType {
	return MoneyCardType(c.Type)
}

func (c Card) Action() ActionCardType {
	return ActionCardType(c.Type)
}

func (c Card) String() string {
	return string(c.CardKey)
}

func (c Card) HasColor(color Color) bool {
	// TODO - sort the colors while initialization and use slices.BinarySearch

	for _, c := range c.Colors {
		if c == color {
			return true
		}
	}
	return false
}

type PurePropertyCardType int

const (
	PurePropertyCardBrown PurePropertyCardType = iota
	PurePropertyCardSky
	PurePropertyCardPink
	PurePropertyCardOrange
	PurePropertyCardRed
	PurePropertyCardYellow
	PurePropertyCardGreen
	PurePropertyCardBlue
	PurePropertyCardUtility
	PurePropertyCardRailroad
)

type WildPropertyCardType int

const (
	WildPropertyCardBrownSky WildPropertyCardType = iota
	WildPropertyCardSkyRailroad
	WildPropertyCardPinkOrange
	WildPropertyCardRedYellow
	WildPropertyCardGreenBlue
	WildPropertyCardGreenRailroad
	WildPropertyCardUtilityRailroad
	WildPropertyCardWild
)

type MoneyCardType int

const (
	MoneyCard1 MoneyCardType = iota
	MoneyCard2
	MoneyCard3
	MoneyCard4
	MoneyCard5
	MoneyCard10
)

type ActionCardType int

const (
	ActionCardPassGo ActionCardType = iota
	ActionCardDoubleTheRent
	ActionCardItsMyBirthday
	ActionCardHouse
	ActionCardSlyDeal
	ActionCardForcedDeal
	ActionCardDebtCollector
	ActionCardHotel
	ActionCardJustSayNo
	ActionCardDealBreaker
	ActionCardRentBrownSky
	ActionCardRentPinkOrange
	ActionCardRentRedYellow
	ActionCardRentGreenBlue
	ActionCardRentUtilityRailroad
	ActionCardRentWild
)

type Cards []Card

func (c *Cards) Len() int {
	return len(*c)
}

func (c *Cards) Add(card ...Card) {
	*c = append(*c, card...)
}

func (c *Cards) Remove(ck CardKey) (Card, bool) {
	if c == nil {
		return Card{}, false
	}

	cards := *c
	for i, c1 := range cards {
		if c1.CardKey == ck {
			*c = append(cards[:i], cards[i+1:]...)
			return c1, true
		}
	}

	return Card{}, false
}
