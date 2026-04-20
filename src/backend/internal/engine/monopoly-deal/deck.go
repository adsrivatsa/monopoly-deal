package monopoly_deal

import (
	"math/rand"
	"slices"
	"time"
)

type Deck struct {
	Cards Cards `json:"cards" msgpack:"a"`
}

func NewDeck(cfg Settings, gen *IdentifierGenerator) (Deck, map[Identifier]Card) {
	n := 28 + 11 + 20 + 34 + 13
	cards := make([]Card, 0, n*cfg.NumDecks)
	cardMap := make(map[Identifier]Card)

	addCopies := func(ck AssetKey, count int) {
		card, ok := CardByAssetKey[ck]
		if !ok {
			panic("missing card definition for key: " + string(ck))
		}

		for i := 0; i < count*cfg.NumDecks; i++ {
			id := gen.New()
			card = NewCard(id, card.Category, card.AssetKey, card.Value, card.Colors...)
			cardMap[id] = card
			cards = append(cards, card)
		}
	}

	// Pure property cards (28)
	addCopies(AssetKeyBalticAve, 1)
	addCopies(AssetKeyMediterraneanAve, 1)
	addCopies(AssetKeyConnecticutAve, 1)
	addCopies(AssetKeyOrientalAve, 1)
	addCopies(AssetKeyVermontAve, 1)
	addCopies(AssetKeyStCharlesPlace, 1)
	addCopies(AssetKeyVirginiaAve, 1)
	addCopies(AssetKeyStatesAve, 1)
	addCopies(AssetKeyNewYorkAve, 1)
	addCopies(AssetKeyStJamesPlace, 1)
	addCopies(AssetKeyTennesseeAve, 1)
	addCopies(AssetKeyKentuckyAve, 1)
	addCopies(AssetKeyIndianaAve, 1)
	addCopies(AssetKeyIllinoisAve, 1)
	addCopies(AssetKeyVentnorAve, 1)
	addCopies(AssetKeyMarvinGardens, 1)
	addCopies(AssetKeyAtlanticAve, 1)
	addCopies(AssetKeyNorthCarolinaAve, 1)
	addCopies(AssetKeyPacificAve, 1)
	addCopies(AssetKeyPennsylvaniaAve, 1)
	addCopies(AssetKeyBoardwalk, 1)
	addCopies(AssetKeyParkPlace, 1)
	addCopies(AssetKeyWaterWorks, 1)
	addCopies(AssetKeyElectricCompany, 1)
	addCopies(AssetKeyShortLine, 1)
	addCopies(AssetKeyBandORailRoad, 1)
	addCopies(AssetKeyReadingRailroad, 1)
	addCopies(AssetKeyPennsylvaniaRailroad, 1)

	// Wild property cards (11)
	addCopies(AssetKeyWildBrownSky, 1)
	addCopies(AssetKeyWildSkyRailroad, 1)
	addCopies(AssetKeyWildPinkOrange, 2)
	addCopies(AssetKeyWildRedYellow, 2)
	addCopies(AssetKeyWildGreenBlue, 1)
	addCopies(AssetKeyWildGreenRailroad, 1)
	addCopies(AssetKeyWildUtilityRailroad, 1)
	addCopies(AssetKeyWildWild, 2)

	// Money cards (20)
	addCopies(AssetKeyMoney1, 6)
	addCopies(AssetKeyMoney2, 5)
	addCopies(AssetKeyMoney3, 3)
	addCopies(AssetKeyMoney4, 3)
	addCopies(AssetKeyMoney5, 2)
	addCopies(AssetKeyMoney10, 1)

	// Action cards (34)
	addCopies(AssetKeyPassGo, 10)
	addCopies(AssetKeyDoubleTheRent, 2)
	addCopies(AssetKeyItsMyBirthday, 3)
	addCopies(AssetKeyHouse, 3)
	addCopies(AssetKeySlyDeal, 3)
	addCopies(AssetKeyForcedDeal, 3)
	addCopies(AssetKeyDebtCollector, 3)
	addCopies(AssetKeyHotel, 2)
	addCopies(AssetKeyJustSayNo, 3)
	addCopies(AssetKeyDealBreaker, 2)

	// Rent cards (13)
	addCopies(AssetKeyRentBrownSky, 2)
	addCopies(AssetKeyRentPinkOrange, 2)
	addCopies(AssetKeyRentRedYellow, 2)
	addCopies(AssetKeyRentGreenBlue, 2)
	addCopies(AssetKeyRentUtilityRailroad, 2)
	addCopies(AssetKeyRentWild, 3)

	d := Deck{Cards: cards}

	d.Shuffle()

	return d, cardMap
}

func (d *Deck) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//r := rand.New(rand.NewSource(5))
	r.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

func (d *Deck) Draw(n int) Cards {
	if n <= 0 || len(d.Cards) == 0 {
		return []Card{}
	}

	if n > len(d.Cards) {
		n = len(d.Cards)
	}

	drawn := slices.Clone(d.Cards[:n])
	d.Cards = d.Cards[n:]

	return drawn
}

func (d *Deck) Add(c Card) {
	d.Cards = append(d.Cards, c)
}
