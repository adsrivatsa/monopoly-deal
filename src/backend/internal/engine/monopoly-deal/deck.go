package monopoly_deal

import (
	"math/rand"
	"slices"
)

type Deck struct {
	Cards []Card
}

func NewDeck(cfg Settings, shuffle bool) Deck {
	cards := make([]Card, 0, 110)

	addCopies := func(ck CardKey, count int) {
		card, ok := CardByKey[ck]
		if !ok {
			panic("missing card definition for key: " + string(ck))
		}

		for i := 0; i < count*int(cfg.NumDecks); i++ {
			cards = append(cards, card)
		}
	}

	// Pure property cards (28)
	addCopies(CardKeyBalticAve, 1)
	addCopies(CardKeyMediterraneanAve, 1)
	addCopies(CardKeyConnecticutAve, 1)
	addCopies(CardKeyOrientalAve, 1)
	addCopies(CardKeyVermontAve, 1)
	addCopies(CardKeyStCharlesPlace, 1)
	addCopies(CardKeyVirginiaAve, 1)
	addCopies(CardKeyStateAve, 1)
	addCopies(CardKeyNewYorkAve, 1)
	addCopies(CardKeyStJamesPlace, 1)
	addCopies(CardKeyTennesseeAve, 1)
	addCopies(CardKeyKentuckyAve, 1)
	addCopies(CardKeyIndianaAve, 1)
	addCopies(CardKeyIllinoisAve, 1)
	addCopies(CardKeyVentnorAve, 1)
	addCopies(CardKeyMarvinGardens, 1)
	addCopies(CardKeyAtlanticAve, 1)
	addCopies(CardKeyNorthCarolinaAve, 1)
	addCopies(CardKeyPacificAve, 1)
	addCopies(CardKeyPennsylvaniaAve, 1)
	addCopies(CardKeyBoardwalk, 1)
	addCopies(CardKeyParkPlace, 1)
	addCopies(CardKeyWaterWorks, 1)
	addCopies(CardKeyElectricCompany, 1)
	addCopies(CardKeyShortLine, 1)
	addCopies(CardKeyBandORailRoad, 1)
	addCopies(CardKeyReadingRailroad, 1)
	addCopies(CardKeyPennsylvaniaRailroad, 1)

	// Wild property cards (12)
	addCopies(CardKeyWildBrownSky, 1)
	addCopies(CardKeyWildSkyRailroad, 1)
	addCopies(CardKeyWildPinkOrange, 2)
	addCopies(CardKeyWildRedYellow, 2)
	addCopies(CardKeyWildGreenBlue, 2)
	addCopies(CardKeyWildGreenRailroad, 1)
	addCopies(CardKeyWildUtilityRailroad, 1)
	addCopies(CardKeyWildWild, 2)

	// Money cards (20)
	addCopies(CardKeyMoney1, 6)
	addCopies(CardKeyMoney2, 5)
	addCopies(CardKeyMoney3, 3)
	addCopies(CardKeyMoney4, 3)
	addCopies(CardKeyMoney5, 2)
	addCopies(CardKeyMoney10, 1)

	// Action cards (50)
	addCopies(CardKeyPassGo, 10)
	addCopies(CardKeyDoubleTheRent, 2)
	addCopies(CardKeyItsMyBirthday, 3)
	addCopies(CardKeyHouse, 3)
	addCopies(CardKeySlyDeal, 3)
	addCopies(CardKeyForcedDeal, 3)
	addCopies(CardKeyDebtCollector, 3)
	addCopies(CardKeyHotel, 2)
	addCopies(CardKeyJustSayNo, 3)
	addCopies(CardKeyDealBreaker, 2)
	addCopies(CardKeyRentBrownSky, 2)
	addCopies(CardKeyRentPinkOrange, 2)
	addCopies(CardKeyRentRedYellow, 2)
	addCopies(CardKeyRentGreenBlue, 2)
	addCopies(CardKeyRentUtilityRailroad, 2)
	addCopies(CardKeyRentWild, 3)

	d := Deck{Cards: cards}

	if shuffle {
		d.Shuffle()
	}

	return d
}

func (d *Deck) Shuffle() {
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r := rand.New(rand.NewSource(43))
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
