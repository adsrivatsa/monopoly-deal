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
	nAction := cfg.PaydayAmount + cfg.DoubleTheRentAmount + cfg.PartyBillAmount +
		cfg.HouseAmount + cfg.PropertyStealAmount + cfg.PropertySwapAmount + cfg.SettleUpAmount +
		cfg.HotelAmount + cfg.NahAmount + cfg.SetSnatcherAmount
	nRent := cfg.RentAmount*5 + cfg.WildRentAmount
	n := 28 + 11 + 20 + nAction + nRent
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
	addCopies(AssetKeyBrown1, 1)
	addCopies(AssetKeyBrown2, 1)
	addCopies(AssetKeySky1, 1)
	addCopies(AssetKeySky2, 1)
	addCopies(AssetKeySky3, 1)
	addCopies(AssetKeyPink1, 1)
	addCopies(AssetKeyPink2, 1)
	addCopies(AssetKeyPink3, 1)
	addCopies(AssetKeyOrange1, 1)
	addCopies(AssetKeyOrange2, 1)
	addCopies(AssetKeyOrange3, 1)
	addCopies(AssetKeyRed1, 1)
	addCopies(AssetKeyRed2, 1)
	addCopies(AssetKeyRed3, 1)
	addCopies(AssetKeyYellow1, 1)
	addCopies(AssetKeyYellow2, 1)
	addCopies(AssetKeyYellow3, 1)
	addCopies(AssetKeyGreen1, 1)
	addCopies(AssetKeyGreen2, 1)
	addCopies(AssetKeyGreen3, 1)
	addCopies(AssetKeyBlue1, 1)
	addCopies(AssetKeyBlue2, 1)
	addCopies(AssetKeyUtil1, 1)
	addCopies(AssetKeyUtil2, 1)
	addCopies(AssetKeyRail1, 1)
	addCopies(AssetKeyRail2, 1)
	addCopies(AssetKeyRail3, 1)
	addCopies(AssetKeyRail4, 1)

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
	addCopies(AssetKeyPassGo, cfg.PaydayAmount)
	addCopies(AssetKeyDoubleTheRent, cfg.DoubleTheRentAmount)
	addCopies(AssetKeyItsMyBirthday, cfg.PartyBillAmount)
	addCopies(AssetKeyHouse, cfg.HouseAmount)
	addCopies(AssetKeySlyDeal, cfg.PropertyStealAmount)
	addCopies(AssetKeyForcedDeal, cfg.PropertySwapAmount)
	addCopies(AssetKeyDebtCollector, cfg.SettleUpAmount)
	addCopies(AssetKeyHotel, cfg.HotelAmount)
	addCopies(AssetKeyJustSayNo, cfg.NahAmount)
	addCopies(AssetKeyDealBreaker, cfg.SetSnatcherAmount)

	// Rent cards (13)
	addCopies(AssetKeyRentBrownSky, cfg.RentAmount)
	addCopies(AssetKeyRentPinkOrange, cfg.RentAmount)
	addCopies(AssetKeyRentRedYellow, cfg.RentAmount)
	addCopies(AssetKeyRentGreenBlue, cfg.RentAmount)
	addCopies(AssetKeyRentUtilityRailroad, cfg.RentAmount)
	addCopies(AssetKeyRentWild, cfg.WildRentAmount)

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
