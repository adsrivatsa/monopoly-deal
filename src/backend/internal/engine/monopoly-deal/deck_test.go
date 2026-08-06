package monopoly_deal

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewDeckUniqueIDs(t *testing.T) {
	for _, numDecks := range []int{1, 2, 3} {
		cfg := DefaultSettings()
		cfg.NumDecks = numDecks

		gen := NewIdentifierGenerator()
		d, cardMap := NewDeck(cfg, gen)

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

func TestGameSnapshotRoundTripPreservesIDGenerator(t *testing.T) {
	cfg := DefaultSettings()
	cfg.NumDecks = 3

	players := []uuid.UUID{uuid.New(), uuid.New()}
	g := NewGame(cfg, players)

	next := g.IDGenerator.Next

	buf, err := g.EncodeMsgpack()
	if err != nil {
		t.Fatal(err)
	}

	g2, err := DecodeMsgpack(buf)
	if err != nil {
		t.Fatal(err)
	}

	if g2.IDGenerator.Next != next {
		t.Fatalf("id generator not preserved: want %d, got %d", next, g2.IDGenerator.Next)
	}

	// new ids issued after a reload must not collide with existing cards
	id := g2.IDGenerator.New()
	if _, ok := g2.Cards[id]; ok {
		t.Fatalf("newly issued id %q collides with an existing card", id)
	}
}
