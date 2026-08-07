package deal_no_mercy

import (
	"math/rand"
	"slices"
	"time"
)

// The deck doubles as the discard pile: consumed action cards are Add-ed to
// the back and Draw takes from the front. Cards that stay on the table when
// played (money, properties, shacks) must NEVER be added back here while
// they are in play — otherwise the same card id ends up in two zones.
type Deck struct {
	Cards Cards `json:"cards" msgpack:"a"`
}

// NewDeck builds the No Mercy deck. With DefaultSettings and NumDecks=1 it
// contains exactly 120 cards:
//
//	28 pure properties + 11 wilds + 23 money + 37 actions + 5 NAH! + 16 rents
func NewDeck(cfg Settings, gen *IdentifierGenerator) (Deck, map[Identifier]Card) {
	nAction := cfg.SetSnatcherAmount + cfg.DebtTrapAmount + cfg.GoAgainAmount +
		cfg.HeistAmount + cfg.MarketCrashAmount + cfg.BigPaydayAmount + cfg.RepoManAmount +
		cfg.ShackAmount + cfg.PropertyRaidAmount + cfg.TaxDayAmount + cfg.PickpocketAmount +
		cfg.BankSwapAmount + cfg.YoinkAmount + cfg.NahAmount
	nRent := cfg.RentAmount*5 + cfg.DoubleRentAmount*5 + cfg.WildRentAmount + cfg.DoubleRentWildAmount
	n := 28 + 11 + 23 + nAction + nRent
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

	// Pure property cards (28) — same roster as classic
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

	// Wild property cards (11): 9 two-color (2x pink/orange, 2x red/yellow,
	// 1x each other pair) + 2 any-color
	addCopies(AssetKeyWildBrownSky, 1)
	addCopies(AssetKeyWildSkyRailroad, 1)
	addCopies(AssetKeyWildPinkOrange, 2)
	addCopies(AssetKeyWildRedYellow, 2)
	addCopies(AssetKeyWildGreenBlue, 1)
	addCopies(AssetKeyWildGreenRailroad, 1)
	addCopies(AssetKeyWildUtilityRailroad, 1)
	addCopies(AssetKeyWildWild, 2)

	// Money cards (23) incl. the new 15M.
	// The official per-denomination split is unpublished; card-list.md
	// proposes 18 money cards, which leaves the 120-card box total only
	// reachable by counting its 5 reference cards. The engine has no
	// reference cards, so the money pile absorbs those 5 slots (see
	// docs/deal-no-mercy/engine-notes.md).
	addCopies(AssetKeyMoney1, 6)
	addCopies(AssetKeyMoney2, 5)
	addCopies(AssetKeyMoney3, 4)
	addCopies(AssetKeyMoney4, 3)
	addCopies(AssetKeyMoney5, 2)
	addCopies(AssetKeyMoney10, 2)
	addCopies(AssetKeyMoney15, 1)

	// Action cards (42 with defaults, incl. NAH!)
	addCopies(AssetKeySetSnatcher, cfg.SetSnatcherAmount)
	addCopies(AssetKeyDebtTrap, cfg.DebtTrapAmount)
	addCopies(AssetKeyGoAgain, cfg.GoAgainAmount)
	addCopies(AssetKeyHeist, cfg.HeistAmount)
	addCopies(AssetKeyMarketCrash, cfg.MarketCrashAmount)
	addCopies(AssetKeyBigPayday, cfg.BigPaydayAmount)
	addCopies(AssetKeyRepoMan, cfg.RepoManAmount)
	addCopies(AssetKeyShack, cfg.ShackAmount)
	addCopies(AssetKeyPropertyRaid, cfg.PropertyRaidAmount)
	addCopies(AssetKeyTaxDay, cfg.TaxDayAmount)
	addCopies(AssetKeyPickpocket, cfg.PickpocketAmount)
	addCopies(AssetKeyBankSwap, cfg.BankSwapAmount)
	addCopies(AssetKeyYoink, cfg.YoinkAmount)
	addCopies(AssetKeyNah, cfg.NahAmount)

	// Rent cards (16 with defaults)
	addCopies(AssetKeyRentBrownSky, cfg.RentAmount)
	addCopies(AssetKeyRentPinkOrange, cfg.RentAmount)
	addCopies(AssetKeyRentRedYellow, cfg.RentAmount)
	addCopies(AssetKeyRentGreenBlue, cfg.RentAmount)
	addCopies(AssetKeyRentUtilityRailroad, cfg.RentAmount)
	addCopies(AssetKeyRentWild, cfg.WildRentAmount)
	addCopies(AssetKeyDoubleRentBrownSky, cfg.DoubleRentAmount)
	addCopies(AssetKeyDoubleRentPinkOrange, cfg.DoubleRentAmount)
	addCopies(AssetKeyDoubleRentRedYellow, cfg.DoubleRentAmount)
	addCopies(AssetKeyDoubleRentGreenBlue, cfg.DoubleRentAmount)
	addCopies(AssetKeyDoubleRentUtilityRailroad, cfg.DoubleRentAmount)
	addCopies(AssetKeyDoubleRentWild, cfg.DoubleRentWildAmount)

	d := Deck{Cards: cards}

	d.Shuffle()

	return d, cardMap
}

func (d *Deck) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
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
