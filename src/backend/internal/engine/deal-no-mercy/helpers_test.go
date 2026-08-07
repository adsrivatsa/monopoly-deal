package deal_no_mercy

import (
	"testing"

	"github.com/google/uuid"
)

// newTestGame builds a game and returns the randomly dealt opening hands to
// the deck, leaving every hand empty. Tests then stage exactly the cards
// they need via handCardByKey / giveBank / giveSet, keeping every card of
// the 120 available and the scenario deterministic.
func newTestGame(t *testing.T, numPlayers int) (*Game, []uuid.UUID) {
	t.Helper()

	cfg := DefaultSettings()
	players := make([]uuid.UUID, numPlayers)
	for i := range players {
		players[i] = uuid.New()
	}

	g := NewGame(cfg, players)
	for _, pid := range players {
		for _, c := range g.Hands[pid] {
			g.Deck.Add(c)
		}
		g.Hands[pid] = Cards{}
	}

	return g, players
}

// takeFromDeck removes and returns the first deck card matching pred.
func takeFromDeck(t *testing.T, g *Game, pred func(Card) bool) Card {
	t.Helper()

	for i, c := range g.Deck.Cards {
		if pred(c) {
			g.Deck.Cards = append(g.Deck.Cards[:i], g.Deck.Cards[i+1:]...)
			return c
		}
	}

	t.Fatal("no matching card left in deck")
	return Card{}
}

// handCardByKey pulls a specific action/money card from the deck into the
// player's hand so tests can play it deterministically.
func handCardByKey(t *testing.T, g *Game, playerID uuid.UUID, key AssetKey) Card {
	t.Helper()

	card := takeFromDeck(t, g, func(c Card) bool { return c.AssetKey == key })
	hand := g.Hands[playerID]
	hand.Add(card)
	g.Hands[playerID] = hand
	return card
}

// handCardByCategory pulls the first deck card of one of the categories
// into the player's hand.
func handCardByCategory(t *testing.T, g *Game, playerID uuid.UUID, categories ...Category) Card {
	t.Helper()

	card := takeFromDeck(t, g, func(c Card) bool {
		for _, cat := range categories {
			if c.Category == cat {
				return true
			}
		}
		return false
	})
	hand := g.Hands[playerID]
	hand.Add(card)
	g.Hands[playerID] = hand
	return card
}

// giveBank moves deck cards with the given asset keys straight into the
// player's bank.
func giveBank(t *testing.T, g *Game, playerID uuid.UUID, keys ...AssetKey) Cards {
	t.Helper()

	var given Cards
	for _, key := range keys {
		card := takeFromDeck(t, g, func(c Card) bool { return c.AssetKey == key })
		money := g.Money[playerID]
		money.Add(card)
		g.Money[playerID] = money
		given.Add(card)
	}
	return given
}

// giveSet builds a property set of n pure-property cards of the color
// directly on the player's table.
func giveSet(t *testing.T, g *Game, playerID uuid.UUID, color Color, n int) PropertySet {
	t.Helper()

	set := NewPropertySet(g.IDGenerator.New(), color)
	for i := 0; i < n; i++ {
		card := takeFromDeck(t, g, func(c Card) bool {
			return c.Category == CategoryPureProperty && c.ActiveColor == color
		})
		set.Cards.Add(card)
	}

	properties := g.Properties[playerID]
	properties.Add(set)
	g.Properties[playerID] = properties
	return set
}

// giveShack attaches a shack from the deck to one of the player's sets.
func giveShack(t *testing.T, g *Game, playerID uuid.UUID, setID Identifier) Card {
	t.Helper()

	card := takeFromDeck(t, g, func(c Card) bool { return c.AssetKey == AssetKeyShack })
	properties := g.Properties[playerID]
	idx := properties.IndexBySetID(setID)
	if idx == -1 {
		t.Fatal("set not found")
	}
	properties[idx].Shack = &card
	g.Properties[playerID] = properties
	return card
}

// assertZoneIntegrity walks every zone (deck, hands, banks, property set
// cards, shacks) and fails if any card id appears in more than one zone or
// if any card of the game is missing from all zones (conservation).
func assertZoneIntegrity(t *testing.T, g *Game) {
	t.Helper()

	seen := make(map[Identifier]string)
	record := func(zone string, cards ...Card) {
		for _, c := range cards {
			if prev, ok := seen[c.ID]; ok {
				t.Fatalf("card %s (%s) exists in both %s and %s", c.ID, c.AssetKey, prev, zone)
			}
			seen[c.ID] = zone
		}
	}

	record("deck", g.Deck.Cards...)
	for pid, hand := range g.Hands {
		record("hand:"+pid.String(), hand...)
	}
	for pid, money := range g.Money {
		record("money:"+pid.String(), money...)
	}
	for pid, sets := range g.Properties {
		for _, set := range sets {
			record("properties:"+pid.String(), set.Cards...)
			if set.Shack != nil {
				record("shack:"+pid.String(), *set.Shack)
			}
		}
	}

	for id, card := range g.Cards {
		if _, ok := seen[id]; !ok {
			t.Fatalf("card %s (%s) is missing from every zone", id, card.AssetKey)
		}
	}

	if len(seen) != len(g.Cards) {
		t.Fatalf("zones hold %d cards but the game knows %d", len(seen), len(g.Cards))
	}
}

// startFreshTurn hands the turn to the current player index with full
// moves, mirroring the controller-driven turn start without drawing.
func setMoves(g *Game, n int) {
	g.MovesLeft = n
}
