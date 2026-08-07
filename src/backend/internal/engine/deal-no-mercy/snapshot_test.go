package deal_no_mercy

import (
	"reflect"
	"testing"
)

// TestSnapshotRoundTrip builds a rich game state — debts, spent chips, an
// attached shack, live demands, a queued go-again — and asserts that the
// msgpack round trip preserves all of it, including the ID generator
// counter so post-reload ids never collide with existing card ids.
func TestSnapshotRoundTrip(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b := players[0], players[1]

	set := giveSet(t, g, a, ColorGreen, 2)
	giveShack(t, g, a, set.ID)
	giveBank(t, g, b, AssetKeyMoney5, AssetKeyMoney1)

	// outstanding debt: b owes a
	g.DebtChips[b] = g.Config.DebtChipsPerPlayer - 1
	g.Debts = append(g.Debts, DebtObligation{ID: g.IDGenerator.New(), DebtorID: b, CreditorID: a})

	// queued extra turn + live demand
	g.GoAgainQueued = true
	demandID := g.IDGenerator.New()
	g.Demands[demandID] = NewPaymentDemand(demandID, a, b, 7, DemandSourceRent)

	g.SequenceNum = 42
	g.MovesLeft = 1

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
	g2.IDGenerator.Next = next

	if g2.SequenceNum != 42 || g2.MovesLeft != 1 || g2.CurrPlayerIdx != g.CurrPlayerIdx {
		t.Fatalf("turn state not preserved: %+v", g2)
	}

	if !g2.GoAgainQueued {
		t.Fatal("go-again flag not preserved")
	}

	if g2.DebtChips[b] != g.Config.DebtChipsPerPlayer-1 {
		t.Fatalf("debt chips not preserved: got %d", g2.DebtChips[b])
	}

	if !reflect.DeepEqual(g2.Debts, g.Debts) {
		t.Fatalf("debts not preserved: %+v vs %+v", g2.Debts, g.Debts)
	}

	if !reflect.DeepEqual(g2.Demands, g.Demands) {
		t.Fatalf("demands not preserved: %+v vs %+v", g2.Demands, g.Demands)
	}

	if !reflect.DeepEqual(g2.Players, g.Players) {
		t.Fatal("players not preserved")
	}

	for _, pid := range players {
		if !cardsEqual(g2.Hands[pid], g.Hands[pid]) {
			t.Fatalf("hand of %s not preserved", pid)
		}
		if !cardsEqual(g2.Money[pid], g.Money[pid]) {
			t.Fatalf("bank of %s not preserved", pid)
		}
		if !propertySetsEqual(g2.Properties[pid], g.Properties[pid]) {
			t.Fatalf("properties of %s not preserved", pid)
		}
	}

	if !cardsEqual(g2.Deck.Cards, g.Deck.Cards) {
		t.Fatal("deck not preserved (order matters: it doubles as the discard pile)")
	}

	if len(g2.Cards) != len(g.Cards) {
		t.Fatalf("card map not preserved: %d vs %d", len(g2.Cards), len(g.Cards))
	}

	// shack survived on the right set
	sets := g2.Properties[a]
	idx := sets.IndexBySetID(set.ID)
	if idx == -1 || sets[idx].Shack == nil || sets[idx].Shack.AssetKey != AssetKeyShack {
		t.Fatal("shack attachment not preserved")
	}

	assertZoneIntegrity(t, g2)
}

// TestSnapshotIsDeepCopy ensures mutating the restored game does not leak
// into the original.
func TestSnapshotIsDeepCopy(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	handCardByCategory(t, g, a, CategoryMoney)
	set := giveSet(t, g, a, ColorBrown, 1)
	giveShack(t, g, a, set.ID)

	snap, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	g2, err := NewGameFromSnapshot(snap)
	if err != nil {
		t.Fatal(err)
	}

	g2.Hands[a][0].AssetKey = "mutated"
	props := g2.Properties[a]
	props[0].Shack.AssetKey = "mutated"
	props[0].Cards[0].AssetKey = "mutated"
	g2.Deck.Cards[0].AssetKey = "mutated"

	if g.Hands[a][0].AssetKey == "mutated" {
		t.Fatal("hand aliased between snapshot copies")
	}
	if g.Properties[a][0].Shack.AssetKey == "mutated" {
		t.Fatal("shack aliased between snapshot copies")
	}
	if g.Properties[a][0].Cards[0].AssetKey == "mutated" {
		t.Fatal("set cards aliased between snapshot copies")
	}
	if g.Deck.Cards[0].AssetKey == "mutated" {
		t.Fatal("deck aliased between snapshot copies")
	}
}

func cardsEqual(a, b Cards) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].AssetKey != b[i].AssetKey || a[i].ActiveColor != b[i].ActiveColor || a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}

func propertySetsEqual(a, b PropertySets) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Color != b[i].Color || !cardsEqual(a[i].Cards, b[i].Cards) {
			return false
		}
		aShack, bShack := a[i].Shack, b[i].Shack
		if (aShack == nil) != (bShack == nil) {
			return false
		}
		if aShack != nil && aShack.ID != bShack.ID {
			return false
		}
	}
	return true
}
