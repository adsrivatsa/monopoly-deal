package deal_no_mercy

import (
	"testing"

	"github.com/google/uuid"
)

func TestPlayMoneyDoesNotDuplicateCard(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	card := handCardByCategory(t, g, a, CategoryMoney)

	_, err := g.PlayMoney(a, card.ID)
	if err != nil {
		t.Fatal(err)
	}

	if g.MovesLeft != g.Config.MovesPerTurn-1 {
		t.Fatalf("moves left %d, want %d", g.MovesLeft, g.Config.MovesPerTurn-1)
	}

	assertZoneIntegrity(t, g)
}

func TestPlayPropertyDoesNotDuplicateCard(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	card := handCardByCategory(t, g, a, CategoryPureProperty)

	action, err := g.PlayProperty(a, card.ID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if action.PropertySet.Cards.Len() != 1 {
		t.Fatalf("expected 1 card in new set, got %d", action.PropertySet.Cards.Len())
	}

	assertZoneIntegrity(t, g)
}

func TestPlayShackAttachRules(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	// incomplete railroad set: shacks attach to ANY set, complete or not
	set := giveSet(t, g, a, ColorRailroad, 1)

	shack := handCardByKey(t, g, a, AssetKeyShack)
	action, err := g.PlayShack(a, shack.ID, set.ID)
	if err != nil {
		t.Fatal(err)
	}
	if action.PropertySet.Shack == nil {
		t.Fatal("shack not attached")
	}

	properties := g.Properties[a]
	idx := properties.IndexBySetID(set.ID)
	if idx == -1 || properties[idx].Shack == nil || properties[idx].Shack.ID != shack.ID {
		t.Fatal("shack not on the set")
	}

	// the shack stays on the table: it must NOT have been discarded to the deck
	for _, c := range g.Deck.Cards {
		if c.ID == shack.ID {
			t.Fatal("played shack was returned to the deck")
		}
	}

	// max 1 per set
	shack2 := handCardByKey(t, g, a, AssetKeyShack)
	_, err = g.PlayShack(a, shack2.ID, set.ID)
	if err != SetAlreadyHasShack {
		t.Fatalf("expected SetAlreadyHasShack, got %v", err)
	}

	assertZoneIntegrity(t, g)
}

func TestShackRaisesRent(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	set := giveSet(t, g, a, ColorGreen, 1) // base rent 2
	properties := g.Properties[a]
	if rent := properties.ColorRent(g.Config.ShackRentBonus, ColorGreen); rent != 2 {
		t.Fatalf("base rent %d, want 2", rent)
	}

	giveShack(t, g, a, set.ID)
	properties = g.Properties[a]
	if rent := properties.ColorRent(g.Config.ShackRentBonus, ColorGreen); rent != 7 {
		t.Fatalf("shacked rent %d, want 7", rent)
	}
}

func TestPlayBigPaydayDrawsToHandTarget(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	// stage a 5-card hand plus the big payday card
	for i := 0; i < 5; i++ {
		handCardByCategory(t, g, a, CategoryMoney)
	}
	card := handCardByKey(t, g, a, AssetKeyBigPayday)

	action, err := g.PlayBigPayday(a, card.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got := len(g.Hands[a]); got != g.Config.BigPaydayHandTarget {
		t.Fatalf("hand size %d, want %d", got, g.Config.BigPaydayHandTarget)
	}

	if len(action.Cards) != 2 {
		t.Fatalf("drew %d cards, want 2", len(action.Cards))
	}

	assertZoneIntegrity(t, g)
}

func TestPlayBigPaydayAtTargetDrawsNothing(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	// pad hand to the target before playing
	for len(g.Hands[a]) < g.Config.BigPaydayHandTarget {
		handCardByCategory(t, g, a, CategoryMoney)
	}
	card := handCardByKey(t, g, a, AssetKeyBigPayday) // now above target

	action, err := g.PlayBigPayday(a, card.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(action.Cards) != 0 {
		t.Fatalf("drew %d cards, want 0", len(action.Cards))
	}

	assertZoneIntegrity(t, g)
}

func TestGoAgainTurnOrder(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b := players[0], players[1]

	card := handCardByKey(t, g, a, AssetKeyGoAgain)

	action, err := g.PlayGoAgain(a, card.ID)
	if err != nil {
		t.Fatal(err)
	}
	if action.LastPlayedCard.ID != card.ID {
		t.Fatal("wrong card recorded")
	}

	// go again is the last play of the turn
	if g.MovesLeft != 0 {
		t.Fatalf("moves left %d, want 0", g.MovesLeft)
	}
	if !g.GoAgainQueued {
		t.Fatal("go again not queued")
	}

	turn, err := g.CompleteTurn(a)
	if err != nil {
		t.Fatal(err)
	}

	// same player takes another full turn
	if turn.PlayerID != a || !turn.GoAgain {
		t.Fatalf("expected go-again turn for a, got %+v", turn)
	}
	if g.Players[g.CurrPlayerIdx] != a {
		t.Fatal("turn advanced past a")
	}
	if g.MovesLeft != g.Config.MovesPerTurn {
		t.Fatalf("moves not reset: %d", g.MovesLeft)
	}
	if g.GoAgainQueued {
		t.Fatal("go again flag not cleared")
	}

	// ending the bonus turn advances normally
	turn2, err := g.CompleteTurn(a)
	if err != nil {
		t.Fatal(err)
	}
	if turn2.PlayerID != b || turn2.GoAgain {
		t.Fatalf("expected normal turn for b, got %+v", turn2)
	}

	assertZoneIntegrity(t, g)
}

func TestCompleteTurnEnforcesHandLimit(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	for len(g.Hands[a]) <= g.Config.MaxHandSize {
		handCardByCategory(t, g, a, CategoryMoney)
	}

	_, err := g.CompleteTurn(a)
	if err == nil {
		t.Fatal("expected hand limit error")
	}

	// discard down and retry
	hand := g.Hands[a]
	over := len(hand) - g.Config.MaxHandSize
	ids := make([]Identifier, 0, over)
	for i := 0; i < over; i++ {
		ids = append(ids, hand[i].ID)
	}
	_, err = g.DiscardCards(a, ids...)
	if err != nil {
		t.Fatal(err)
	}

	_, err = g.CompleteTurn(a)
	if err != nil {
		t.Fatal(err)
	}

	assertZoneIntegrity(t, g)
}

func TestStartTurnDrawsFiveOnEmptyHand(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	// empty b's hand back into the deck
	for _, c := range g.Hands[b] {
		g.Deck.Add(c)
	}
	g.Hands[b] = Cards{}

	_, err := g.CompleteTurn(a)
	if err != nil {
		t.Fatal(err)
	}

	if got := len(g.Hands[b]); got != g.Config.StartNumCards {
		t.Fatalf("empty-hand draw got %d cards, want %d", got, g.Config.StartNumCards)
	}

	assertZoneIntegrity(t, g)
}

func TestWinAtThreeCompleteSets(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	giveSet(t, g, a, ColorBrown, 2)
	giveSet(t, g, a, ColorBlue, 2)

	_, _, won, err := g.CheckWinConditions(a)
	if err != nil {
		t.Fatal(err)
	}
	if won {
		t.Fatal("won with only 2 complete sets")
	}

	giveSet(t, g, a, ColorUtility, 2)

	sets, _, won, err := g.CheckWinConditions(a)
	if err != nil {
		t.Fatal(err)
	}
	if !won || sets != 3 {
		t.Fatalf("expected win with 3 sets, got won=%v sets=%d", won, sets)
	}
}

func TestZoneIntegrityAfterScriptedSequence(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b, c := players[0], players[1], players[2]

	// a: bank a money card, play a property, shack it
	money := handCardByCategory(t, g, a, CategoryMoney)
	if _, err := g.PlayMoney(a, money.ID); err != nil {
		t.Fatal(err)
	}
	assertZoneIntegrity(t, g)

	prop := handCardByCategory(t, g, a, CategoryPureProperty)
	action, err := g.PlayProperty(a, prop.ID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertZoneIntegrity(t, g)

	shack := handCardByKey(t, g, a, AssetKeyShack)
	if _, err := g.PlayShack(a, shack.ID, action.PropertySet.ID); err != nil {
		t.Fatal(err)
	}
	assertZoneIntegrity(t, g)

	if err := func() error { _, err := g.CompleteTurn(a); return err }(); err != nil {
		t.Fatal(err)
	}
	assertZoneIntegrity(t, g)

	// b: heist a's bank
	setMoves(g, g.Config.MovesPerTurn)
	heist := handCardByKey(t, g, b, AssetKeyHeist)
	heistAction, err := g.PlayHeist(b, heist.ID, map[uuid.UUID]Identifier{a: money.ID})
	if err != nil {
		t.Fatal(err)
	}
	assertZoneIntegrity(t, g)

	if _, err := g.ComplyBankCardDemand(a, heistAction.Demands[0].ID); err != nil {
		t.Fatal(err)
	}
	assertZoneIntegrity(t, g)

	// b: market crash a's shacked single-card set → shack orphaned to deck
	crash := handCardByKey(t, g, b, AssetKeyMarketCrash)
	crashAction, err := g.PlayMarketCrash(b, crash.ID, map[uuid.UUID]Identifier{a: prop.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.ComplyPropertyDemand(a, crashAction.Demands[0].ID); err != nil {
		t.Fatal(err)
	}
	assertZoneIntegrity(t, g)

	// the orphaned shack must be back in the deck
	found := false
	for _, cd := range g.Deck.Cards {
		if cd.ID == shack.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("orphaned shack was not discarded to the deck")
	}

	// hand-limit discards for b, then pass around to c
	for len(g.Hands[b]) > g.Config.MaxHandSize {
		if _, err := g.DiscardCards(b, g.Hands[b][0].ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := g.CompleteTurn(b); err != nil {
		t.Fatal(err)
	}
	assertZoneIntegrity(t, g)

	if g.Players[g.CurrPlayerIdx] != c {
		t.Fatal("turn order broken")
	}
}
