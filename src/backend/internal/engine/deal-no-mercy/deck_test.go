package deal_no_mercy

import (
	"testing"
)

// TestDeckComposition asserts the exact per-card counts of the No Mercy
// deck (docs/deal-no-mercy/card-list.md) and that the default deck sums to
// exactly 120 cards.
func TestDeckComposition(t *testing.T) {
	cfg := DefaultSettings()
	gen := NewIdentifierGenerator()
	d, cardMap := NewDeck(cfg, gen)

	if len(d.Cards) != 120 {
		t.Fatalf("deck has %d cards, want 120", len(d.Cards))
	}

	if len(cardMap) != 120 {
		t.Fatalf("card map has %d cards, want 120", len(cardMap))
	}

	expected := map[AssetKey]int{
		// 28 pure properties
		AssetKeyBrown1: 1, AssetKeyBrown2: 1,
		AssetKeySky1: 1, AssetKeySky2: 1, AssetKeySky3: 1,
		AssetKeyPink1: 1, AssetKeyPink2: 1, AssetKeyPink3: 1,
		AssetKeyOrange1: 1, AssetKeyOrange2: 1, AssetKeyOrange3: 1,
		AssetKeyRed1: 1, AssetKeyRed2: 1, AssetKeyRed3: 1,
		AssetKeyYellow1: 1, AssetKeyYellow2: 1, AssetKeyYellow3: 1,
		AssetKeyGreen1: 1, AssetKeyGreen2: 1, AssetKeyGreen3: 1,
		AssetKeyBlue1: 1, AssetKeyBlue2: 1,
		AssetKeyUtil1: 1, AssetKeyUtil2: 1,
		AssetKeyRail1: 1, AssetKeyRail2: 1, AssetKeyRail3: 1, AssetKeyRail4: 1,
		// 11 wilds (9 two-color + 2 any-color)
		AssetKeyWildBrownSky: 1, AssetKeyWildSkyRailroad: 1,
		AssetKeyWildPinkOrange: 2, AssetKeyWildRedYellow: 2,
		AssetKeyWildGreenBlue: 1, AssetKeyWildGreenRailroad: 1,
		AssetKeyWildUtilityRailroad: 1, AssetKeyWildWild: 2,
		// 23 money (incl. the new 15M)
		AssetKeyMoney1: 6, AssetKeyMoney2: 5, AssetKeyMoney3: 4,
		AssetKeyMoney4: 3, AssetKeyMoney5: 2, AssetKeyMoney10: 2, AssetKeyMoney15: 1,
		// 42 actions
		AssetKeySetSnatcher: 3, AssetKeyDebtTrap: 3, AssetKeyGoAgain: 3,
		AssetKeyHeist: 3, AssetKeyMarketCrash: 3, AssetKeyBigPayday: 4,
		AssetKeyRepoMan: 2, AssetKeyShack: 3, AssetKeyPropertyRaid: 3,
		AssetKeyTaxDay: 2, AssetKeyPickpocket: 3, AssetKeyBankSwap: 2,
		AssetKeyYoink: 3, AssetKeyNah: 5,
		// 16 rents
		AssetKeyRentBrownSky: 1, AssetKeyRentPinkOrange: 1, AssetKeyRentRedYellow: 1,
		AssetKeyRentGreenBlue: 1, AssetKeyRentUtilityRailroad: 1, AssetKeyRentWild: 3,
		AssetKeyDoubleRentBrownSky: 1, AssetKeyDoubleRentPinkOrange: 1,
		AssetKeyDoubleRentRedYellow: 1, AssetKeyDoubleRentGreenBlue: 1,
		AssetKeyDoubleRentUtilityRailroad: 1, AssetKeyDoubleRentWild: 3,
	}

	total := 0
	for _, n := range expected {
		total += n
	}
	if total != 120 {
		t.Fatalf("expected composition sums to %d, want 120", total)
	}

	counts := make(map[AssetKey]int)
	for _, c := range d.Cards {
		counts[c.AssetKey]++
	}

	for key, want := range expected {
		if counts[key] != want {
			t.Errorf("asset %q: got %d copies, want %d", key, counts[key], want)
		}
	}
	for key := range counts {
		if _, ok := expected[key]; !ok {
			t.Errorf("unexpected asset %q in deck", key)
		}
	}
}

func TestNewDeckUniqueIDs(t *testing.T) {
	for _, numDecks := range []int{1, 2, 3} {
		cfg := DefaultSettings()
		cfg.NumDecks = numDecks

		gen := NewIdentifierGenerator()
		d, cardMap := NewDeck(cfg, gen)

		if len(d.Cards) != 120*numDecks {
			t.Fatalf("numDecks=%d: deck has %d cards, want %d", numDecks, len(d.Cards), 120*numDecks)
		}

		if len(d.Cards) != len(cardMap) {
			t.Fatalf("numDecks=%d: deck has %d cards but card map has %d", numDecks, len(d.Cards), len(cardMap))
		}

		seen := make(map[Identifier]Card)
		for _, c := range d.Cards {
			if dup, ok := seen[c.ID]; ok {
				t.Fatalf("numDecks=%d: duplicate card id %q (%s and %s)", numDecks, c.ID, dup.AssetKey, c.AssetKey)
			}
			seen[c.ID] = c
		}
	}
}

func TestSettingsDecodeDefaultsAndTimeouts(t *testing.T) {
	base := DefaultSettings()
	buf, err := base.Encode()
	if err != nil {
		t.Fatal(err)
	}

	var cfg Settings
	if err := cfg.Decode(buf); err != nil {
		t.Fatal(err)
	}

	if cfg.DebtChipsPerPlayer != 3 || cfg.YoinkPayment != 10 || cfg.ShackRentBonus != 5 || cfg.BigPaydayHandTarget != 7 {
		t.Fatalf("no mercy defaults not preserved: %+v", cfg)
	}

	if cfg.MoveTimeout == 0 || cfg.DemandTimeout == 0 {
		t.Fatal("timeouts not derived from speed")
	}

	if cfg.DemandTimeout != cfg.MoveTimeout*2/3 {
		t.Fatalf("demand timeout %v not derived from move timeout %v", cfg.DemandTimeout, cfg.MoveTimeout)
	}
}
