package deal_no_mercy

import (
	"testing"
)

// A yoink against a target with nothing payable and no NAH! auto-resolves at
// creation: the demand is completed (empty comply) and a debt chip is issued.
func TestAutoResolveForcedDemandEmptyTargetIssuesChip(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	yoink := handCardByKey(t, g, a, AssetKeyYoink)
	// b has no bank, no property, no NAH!.

	action, err := g.PlayYoink(a, b, yoink.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(action.Demands) != 1 {
		t.Fatalf("expected 1 demand created, got %d", len(action.Demands))
	}

	complies, err := g.AutoResolveForcedDemands()
	if err != nil {
		t.Fatal(err)
	}
	if len(complies) != 1 {
		t.Fatalf("expected 1 auto-comply, got %d", len(complies))
	}
	if len(g.Demands) != 0 {
		t.Fatalf("expected demand to be resolved, %d remain", len(g.Demands))
	}
	if complies[0].TransferCards == nil || complies[0].TransferCards.Debt == nil {
		t.Fatal("empty comply should issue a debt chip")
	}
	if g.DebtChips[b] != g.Config.DebtChipsPerPlayer-1 {
		t.Fatalf("debtor chips %d, want %d", g.DebtChips[b], g.Config.DebtChipsPerPlayer-1)
	}
	if len(g.PlayerDebts(b)) != 1 {
		t.Fatal("obligation not recorded")
	}
	assertZoneIntegrity(t, g)
}

// A target who CAN pay keeps the choice: no auto-resolution.
func TestAutoResolveLeavesPayableTargetOpen(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	yoink := handCardByKey(t, g, a, AssetKeyYoink)
	giveBank(t, g, b, AssetKeyMoney2)

	if _, err := g.PlayYoink(a, b, yoink.ID); err != nil {
		t.Fatal(err)
	}

	complies, err := g.AutoResolveForcedDemands()
	if err != nil {
		t.Fatal(err)
	}
	if len(complies) != 0 {
		t.Fatalf("payable target should not auto-resolve, got %d complies", len(complies))
	}
	if len(g.Demands) != 1 {
		t.Fatalf("demand should stay open, %d remain", len(g.Demands))
	}
}

// A target holding a NAH! keeps the choice even with nothing payable.
func TestAutoResolveLeavesNahHolderOpen(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	yoink := handCardByKey(t, g, a, AssetKeyYoink)
	handCardByKey(t, g, b, AssetKeyNah) // b can't pay but holds NAH!

	if _, err := g.PlayYoink(a, b, yoink.ID); err != nil {
		t.Fatal(err)
	}

	complies, err := g.AutoResolveForcedDemands()
	if err != nil {
		t.Fatal(err)
	}
	if len(complies) != 0 {
		t.Fatalf("NAH! holder should not auto-resolve, got %d complies", len(complies))
	}
	if len(g.Demands) != 1 {
		t.Fatalf("demand should stay open, %d remain", len(g.Demands))
	}
}

// Multi-target rent: forced (empty, no NAH!) targets auto-resolve, payable /
// NAH!-holding targets stay open in the same pass.
func TestAutoResolveMixedTargets(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b, c := players[0], players[1], players[2]

	// a has a rent-eligible set so rent > 0. Use a two-color rent card.
	rent := handCardByKey(t, g, a, AssetKeyRentBrownSky)
	giveSet(t, g, a, ColorBrown, 1)

	// b: empty, no NAH! → forced. c: has bank → open.
	giveBank(t, g, c, AssetKeyMoney2)

	if _, err := g.PlayRent(a, rent.ID, ColorUnspecified); err != nil {
		t.Fatal(err)
	}
	if len(g.Demands) != 2 {
		t.Fatalf("expected 2 rent demands, got %d", len(g.Demands))
	}

	complies, err := g.AutoResolveForcedDemands()
	if err != nil {
		t.Fatal(err)
	}
	if len(complies) != 1 {
		t.Fatalf("expected exactly 1 forced auto-comply (b), got %d", len(complies))
	}
	if len(g.Demands) != 1 {
		t.Fatalf("expected c's demand to stay open, %d remain", len(g.Demands))
	}
	// The surviving demand must be c's.
	for _, d := range g.Demands {
		if d.TargetID != c {
			t.Fatalf("surviving demand targets %s, want c %s", d.TargetID, c)
		}
	}
	_ = b
	assertZoneIntegrity(t, g)
}

// Go Again is valid only as the final play of the turn (movesLeft == 1) and is
// rejected while moves remain.
func TestGoAgainMustBeLastPlay(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	// Rejected with moves remaining (> 1).
	setMoves(g, 3)
	goAgain := handCardByKey(t, g, a, AssetKeyGoAgain)
	if _, err := g.PlayGoAgain(a, goAgain.ID); err != GoAgainNotLastPlay {
		t.Fatalf("expected GoAgainNotLastPlay with moves remaining, got %v", err)
	}
	// Card must not have been consumed on rejection.
	if g.GoAgainQueued {
		t.Fatal("GoAgainQueued should be false after a rejected play")
	}

	// Valid as the last play (movesLeft == 1).
	setMoves(g, 1)
	if _, err := g.PlayGoAgain(a, goAgain.ID); err != nil {
		t.Fatalf("go again as last play should succeed, got %v", err)
	}
	if !g.GoAgainQueued {
		t.Fatal("GoAgainQueued should be set after a valid play")
	}
	if g.MovesLeft != 0 {
		t.Fatalf("moves left %d, want 0", g.MovesLeft)
	}
}
