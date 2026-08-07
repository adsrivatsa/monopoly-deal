package deal_no_mercy

import (
	"testing"

	"github.com/google/uuid"
)

func findDemandForTarget(t *testing.T, demands []Demand, targetID uuid.UUID) Demand {
	t.Helper()
	for _, d := range demands {
		if d.TargetID == targetID {
			return d
		}
	}
	t.Fatalf("no demand targeting %s", targetID)
	return Demand{}
}

func TestYoinkPaymentAndDebtChipLifecycle(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	yoink := handCardByKey(t, g, a, AssetKeyYoink)
	bank := giveBank(t, g, b, AssetKeyMoney2)

	action, err := g.PlayYoink(a, b, yoink.ID)
	if err != nil {
		t.Fatal(err)
	}

	demand := action.Demands[0]
	if demand.Payment.Amount != g.Config.YoinkPayment {
		t.Fatalf("yoink amount %d, want %d", demand.Payment.Amount, g.Config.YoinkPayment)
	}

	// b cannot cover 10M: must surrender everything payable, then a chip
	// goes to the creditor
	res, err := g.ComplyPaymentDemand(b, demand.ID, bank[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if res.TransferCards == nil || res.TransferCards.Debt == nil {
		t.Fatal("expected a debt chip to be issued")
	}
	if g.DebtChips[b] != g.Config.DebtChipsPerPlayer-1 {
		t.Fatalf("debtor chips %d, want %d", g.DebtChips[b], g.Config.DebtChipsPerPlayer-1)
	}
	if len(g.PlayerDebts(b)) != 1 {
		t.Fatal("obligation not recorded")
	}
	if v, _ := g.CountMoney(a); v != 2 {
		t.Fatalf("creditor received %dM, want 2M", v)
	}
	assertZoneIntegrity(t, g)

	// a finishes; b's turn starts with the settlement due
	if _, err := g.CompleteTurn(a); err != nil {
		t.Fatal(err)
	}

	// every regular play is blocked until the debt is settled
	someCard := g.Hands[b][0]
	if _, err := g.PlayMoney(b, someCard.ID); err != OutstandingDebtExists {
		t.Fatalf("expected OutstandingDebtExists, got %v", err)
	}
	if _, err := g.CompleteTurn(b); err != OutstandingDebtExists {
		t.Fatalf("turn end must be blocked too, got %v", err)
	}

	// settling consumes one of the 3 plays and returns the chip
	debt := g.PlayerDebts(b)[0]
	settled, err := g.SettleDebt(b, debt.ID, someCard.ID)
	if err != nil {
		t.Fatal(err)
	}
	if settled.Debt.ID != debt.ID {
		t.Fatal("wrong debt settled")
	}
	if g.MovesLeft != g.Config.MovesPerTurn-1 {
		t.Fatalf("settlement must consume a play: moves left %d", g.MovesLeft)
	}
	if g.DebtChips[b] != g.Config.DebtChipsPerPlayer {
		t.Fatal("chip did not return to the debtor")
	}
	if len(g.PlayerDebts(b)) != 0 {
		t.Fatal("obligation not cleared")
	}

	// the creditor received the card as if they played it themselves
	found := false
	for _, c := range g.Money[a] {
		if c.ID == someCard.ID {
			found = true
		}
	}
	for _, set := range g.Properties[a] {
		for _, c := range set.Cards {
			if c.ID == someCard.ID {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("creditor did not receive the settlement card")
	}

	// b can play normally now
	if _, err := g.PlayMoney(b, g.Hands[b][0].ID); err != nil && err != OutstandingDebtExists {
		// any error other than debt-blocking is fine (e.g. wrong category);
		// debt blocking must be gone
		if err == OutstandingDebtExists {
			t.Fatal("still debt-blocked after settlement")
		}
	}

	assertZoneIntegrity(t, g)
}

func TestFailedPaymentWithNoChipsWritesOffShortfall(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	g.DebtChips[b] = 0

	yoink := handCardByKey(t, g, a, AssetKeyYoink)
	if _, err := g.PlayYoink(a, b, yoink.ID); err != nil {
		t.Fatal(err)
	}

	var demand Demand
	for _, d := range g.Demands {
		demand = d
	}

	// b owns nothing payable: comply with empty payment
	res, err := g.ComplyPaymentDemand(b, demand.ID)
	if err != nil {
		t.Fatal(err)
	}

	if res.TransferCards.Debt != nil {
		t.Fatal("no chip should be issued when the debtor has none left")
	}
	if len(g.Debts) != 0 {
		t.Fatal("no obligation should exist")
	}

	assertZoneIntegrity(t, g)
}

func TestDebtTrapIssueAndNahDenial(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	// happy path: chip moves immediately on compliance
	trap := handCardByKey(t, g, a, AssetKeyDebtTrap)
	action, err := g.PlayDebtTrap(a, b, trap.ID)
	if err != nil {
		t.Fatal(err)
	}

	res, err := g.ComplyDebtTrapDemand(b, action.Demands[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.TransferDebtChip == nil {
		t.Fatal("expected chip transfer")
	}
	if g.DebtChips[b] != g.Config.DebtChipsPerPlayer-1 || len(g.PlayerDebts(b)) != 1 {
		t.Fatal("chip not moved to obligation")
	}

	// denial path: NAH! blocks debt trap (but not debt itself)
	trap2 := handCardByKey(t, g, a, AssetKeyDebtTrap)
	action2, err := g.PlayDebtTrap(a, b, trap2.ID)
	if err != nil {
		t.Fatal(err)
	}

	nah := handCardByKey(t, g, b, AssetKeyNah)
	denied, err := g.DenyDemand(b, action2.Demands[0].ID, nah.ID)
	if err != nil {
		t.Fatal(err)
	}

	flipped := denied.Demands[0]
	if flipped.IsActive || flipped.TargetID != a || flipped.ID == action2.Demands[0].ID {
		t.Fatalf("denial did not flip the demand: %+v", flipped)
	}

	// attacker accepts the denial: no chip taken
	if _, err := g.ComplyDebtTrapDemand(a, flipped.ID); err != nil {
		t.Fatal(err)
	}
	if g.DebtChips[b] != g.Config.DebtChipsPerPlayer-1 || len(g.PlayerDebts(b)) != 1 {
		t.Fatal("denied trap must not take a second chip")
	}

	assertZoneIntegrity(t, g)
}

func TestDebtTrapRejectedWhenNoChips(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	g.DebtChips[b] = 0
	trap := handCardByKey(t, g, a, AssetKeyDebtTrap)
	if _, err := g.PlayDebtTrap(a, b, trap.ID); err != NoDebtChipsAvailable {
		t.Fatalf("expected NoDebtChipsAvailable, got %v", err)
	}
}

func TestRentChargesAllOpponentsWithShackAndDenial(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b, c := players[0], players[1], players[2]

	set := giveSet(t, g, a, ColorGreen, 1) // rent 2
	giveShack(t, g, a, set.ID)             // rent 7

	rent := handCardByKey(t, g, a, AssetKeyRentGreenBlue)
	action, err := g.PlayRent(a, rent.ID, ColorGreen)
	if err != nil {
		t.Fatal(err)
	}

	if len(action.Demands) != 2 {
		t.Fatalf("rent must fan out to all opponents, got %d demands", len(action.Demands))
	}
	for _, d := range action.Demands {
		if d.Payment.Amount != 7 {
			t.Fatalf("rent amount %d, want 7 (2 base + 5 shack)", d.Payment.Amount)
		}
	}

	// b denies with NAH!
	bDemand := findDemandForTarget(t, action.Demands, b)
	nah := handCardByKey(t, g, b, AssetKeyNah)
	denied, err := g.DenyDemand(b, bDemand.ID, nah.ID)
	if err != nil {
		t.Fatal(err)
	}

	// c pays in full — no change given
	cDemand := findDemandForTarget(t, action.Demands, c)
	cBank := giveBank(t, g, c, AssetKeyMoney10)
	res, err := g.ComplyPaymentDemand(c, cDemand.ID, cBank[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.TransferCards.Debt != nil {
		t.Fatal("full payment must not issue a chip")
	}
	if v, _ := g.CountMoney(a); v != 10 {
		t.Fatalf("creditor bank %dM, want 10M", v)
	}

	// a accepts b's denial
	if _, err := g.ComplyPaymentDemand(a, denied.Demands[0].ID); err != nil {
		t.Fatal(err)
	}
	if len(g.Demands) != 0 {
		t.Fatal("demands should be resolved")
	}

	assertZoneIntegrity(t, g)
}

func TestCounterNahReactivatesDemand(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	yoink := handCardByKey(t, g, a, AssetKeyYoink)
	action, err := g.PlayYoink(a, b, yoink.ID)
	if err != nil {
		t.Fatal(err)
	}

	// b denies, a counter-denies
	bNah := handCardByKey(t, g, b, AssetKeyNah)
	denied, err := g.DenyDemand(b, action.Demands[0].ID, bNah.ID)
	if err != nil {
		t.Fatal(err)
	}

	aNah := handCardByKey(t, g, a, AssetKeyNah)
	counter, err := g.DenyDemand(a, denied.Demands[0].ID, aNah.ID)
	if err != nil {
		t.Fatal(err)
	}

	reactivated := counter.Demands[0]
	if !reactivated.IsActive || reactivated.TargetID != b {
		t.Fatalf("counter-nah must reactivate the demand against b: %+v", reactivated)
	}

	// b must now pay (has nothing → surrender nothing + chip)
	res, err := g.ComplyPaymentDemand(b, reactivated.ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.TransferCards.Debt == nil {
		t.Fatal("expected chip on failed payment")
	}

	assertZoneIntegrity(t, g)
}

func TestDoubleRentDoubles(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	giveSet(t, g, a, ColorGreen, 3) // complete: rent 7

	// pair variant refuses colors outside its pair (checked before any
	// state change, so nothing is consumed)
	card2 := handCardByKey(t, g, a, AssetKeyDoubleRentBrownSky)
	if _, err := g.PlayRent(a, card2.ID, ColorGreen); err != InvalidColorForCard {
		t.Fatalf("expected InvalidColorForCard, got %v", err)
	}

	card := handCardByKey(t, g, a, AssetKeyDoubleRentWild)
	action, err := g.PlayRent(a, card.ID, ColorGreen)
	if err != nil {
		t.Fatal(err)
	}

	if action.Demands[0].Payment.Amount != 14 {
		t.Fatalf("double rent %d, want 14", action.Demands[0].Payment.Amount)
	}
}

func TestSetSnatcherStealsCompleteSetWithShack(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	bSet := giveSet(t, g, b, ColorUtility, 2) // complete
	shack := giveShack(t, g, b, bSet.ID)

	// incomplete sets cannot be snatched
	bIncomplete := giveSet(t, g, b, ColorRed, 1)
	snatcher := handCardByKey(t, g, a, AssetKeySetSnatcher)
	if _, err := g.PlaySetSnatcher(a, b, snatcher.ID, bIncomplete.ID); err == nil {
		t.Fatal("snatching an incomplete set must fail")
	}

	action, err := g.PlaySetSnatcher(a, b, snatcher.ID, bSet.ID)
	if err != nil {
		t.Fatal(err)
	}

	res, err := g.ComplyPropertySetDemand(b, action.Demands[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	got := res.TransferPropertySets.PropertySets[0]
	if got.ID != bSet.ID {
		t.Fatal("wrong set stolen")
	}

	aProps := g.Properties[a]
	idx := aProps.IndexBySetID(bSet.ID)
	if idx == -1 {
		t.Fatal("set did not arrive at the attacker")
	}
	if aProps[idx].Shack == nil || aProps[idx].Shack.ID != shack.ID {
		t.Fatal("shack did not travel with the stolen set")
	}

	assertZoneIntegrity(t, g)
}

func TestMarketCrashStealsFromCompleteSetsAndOrphansShacks(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b, c := players[0], players[1], players[2]

	// b: complete brown set with a shack — complete sets are NOT protected
	bSet := giveSet(t, g, b, ColorBrown, 2)
	bShack := giveShack(t, g, b, bSet.ID)

	// c: single-card set with a shack — stealing it orphans the shack
	cSet := giveSet(t, g, c, ColorRed, 1)
	cShack := giveShack(t, g, c, cSet.ID)

	crash := handCardByKey(t, g, a, AssetKeyMarketCrash)
	picks := map[uuid.UUID]Identifier{
		b: bSet.Cards[0].ID,
		c: cSet.Cards[0].ID,
	}

	action, err := g.PlayMarketCrash(a, crash.ID, picks)
	if err != nil {
		t.Fatal(err)
	}
	if len(action.Demands) != 2 {
		t.Fatalf("market crash must target every opponent, got %d", len(action.Demands))
	}

	bDemand := findDemandForTarget(t, action.Demands, b)
	if _, err := g.ComplyPropertyDemand(b, bDemand.ID); err != nil {
		t.Fatal(err)
	}

	// b keeps a 1-card set with its shack intact
	bProps := g.Properties[b]
	if len(bProps) != 1 || bProps[0].Cards.Len() != 1 || bProps[0].Shack == nil || bProps[0].Shack.ID != bShack.ID {
		t.Fatalf("b's remaining set wrong: %+v", bProps)
	}

	cDemand := findDemandForTarget(t, action.Demands, c)
	if _, err := g.ComplyPropertyDemand(c, cDemand.ID); err != nil {
		t.Fatal(err)
	}

	// c's set emptied → shack discarded to deck
	if len(g.Properties[c]) != 0 {
		t.Fatal("c's emptied set should be gone")
	}
	foundShack := false
	for _, cd := range g.Deck.Cards {
		if cd.ID == cShack.ID {
			foundShack = true
			break
		}
	}
	if !foundShack {
		t.Fatal("orphaned shack must be discarded to the deck")
	}

	// a received both properties
	if len(g.Properties[a]) != 2 {
		t.Fatalf("attacker has %d sets, want 2", len(g.Properties[a]))
	}

	assertZoneIntegrity(t, g)
}

func TestMarketCrashDenialLeavesPropertyInPlace(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	bSet := giveSet(t, g, b, ColorSky, 2)

	crash := handCardByKey(t, g, a, AssetKeyMarketCrash)
	action, err := g.PlayMarketCrash(a, crash.ID, map[uuid.UUID]Identifier{b: bSet.Cards[0].ID})
	if err != nil {
		t.Fatal(err)
	}

	nah := handCardByKey(t, g, b, AssetKeyNah)
	denied, err := g.DenyDemand(b, action.Demands[0].ID, nah.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := g.ComplyPropertyDemand(a, denied.Demands[0].ID); err != nil {
		t.Fatal(err)
	}

	if g.Properties[b][0].Cards.Len() != 2 {
		t.Fatal("denied market crash must leave the property with the target")
	}
	if len(g.Properties[a]) != 0 {
		t.Fatal("attacker must not gain anything on denial")
	}

	assertZoneIntegrity(t, g)
}

func TestHeistStealsOneBankCardFromEveryOpponent(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b, c := players[0], players[1], players[2]

	bBank := giveBank(t, g, b, AssetKeyMoney3)
	cBank := giveBank(t, g, c, AssetKeyMoney4, AssetKeyMoney1)

	heist := handCardByKey(t, g, a, AssetKeyHeist)

	// picks must cover exactly the opponents with non-empty banks
	if _, err := g.PlayHeist(a, heist.ID, map[uuid.UUID]Identifier{b: bBank[0].ID}); err != InvalidTargetPicks {
		t.Fatalf("expected InvalidTargetPicks, got %v", err)
	}

	action, err := g.PlayHeist(a, heist.ID, map[uuid.UUID]Identifier{b: bBank[0].ID, c: cBank[0].ID})
	if err != nil {
		t.Fatal(err)
	}

	for _, d := range action.Demands {
		if _, err := g.ComplyBankCardDemand(d.TargetID, d.ID); err != nil {
			t.Fatal(err)
		}
	}

	if v, _ := g.CountMoney(a); v != 7 {
		t.Fatalf("attacker bank %dM, want 7M", v)
	}
	if v, _ := g.CountMoney(b); v != 0 {
		t.Fatalf("b bank %dM, want 0M", v)
	}
	if v, _ := g.CountMoney(c); v != 1 {
		t.Fatalf("c bank %dM, want 1M", v)
	}

	assertZoneIntegrity(t, g)
}

func TestHeistWithNoBanksRejected(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	heist := handCardByKey(t, g, a, AssetKeyHeist)
	if _, err := g.PlayHeist(a, heist.ID, nil); err != NoValidTargets {
		t.Fatalf("expected NoValidTargets, got %v", err)
	}
}

func TestPropertyRaidTakesWholeColorIncludingCompleteSets(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b := players[0], players[1]

	bSet := giveSet(t, g, b, ColorSky, 3) // complete
	shack := giveShack(t, g, b, bSet.ID)

	raid := handCardByKey(t, g, a, AssetKeyPropertyRaid)
	action, err := g.PlayPropertyRaid(a, raid.ID, ColorSky)
	if err != nil {
		t.Fatal(err)
	}

	// only b holds sky sets
	if len(action.Demands) != 1 {
		t.Fatalf("expected 1 demand, got %d", len(action.Demands))
	}

	res, err := g.ComplyColorPropertiesDemand(b, action.Demands[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.TransferPropertySets.PropertySets) != 1 {
		t.Fatal("expected the whole sky set to transfer")
	}

	aProps := g.Properties[a]
	idx := aProps.IndexBySetID(bSet.ID)
	if idx == -1 || !aProps[idx].IsComplete() || aProps[idx].Shack == nil || aProps[idx].Shack.ID != shack.ID {
		t.Fatal("complete set (with shack) did not arrive intact")
	}
	if len(g.Properties[b]) != 0 {
		t.Fatal("b should have lost all sky sets")
	}

	assertZoneIntegrity(t, g)
}

func TestPropertyRaidNoTargetsRejected(t *testing.T) {
	g, players := newTestGame(t, 2)
	a := players[0]

	raid := handCardByKey(t, g, a, AssetKeyPropertyRaid)
	if _, err := g.PlayPropertyRaid(a, raid.ID, ColorSky); err != NoValidTargets {
		t.Fatalf("expected NoValidTargets, got %v", err)
	}
}

func TestRepoManDistribution(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b, c := players[0], players[1], players[2]

	bSet := giveSet(t, g, b, ColorBrown, 2)
	bSet2 := giveSet(t, g, b, ColorRed, 1)

	repo := handCardByKey(t, g, a, AssetKeyRepoMan)
	action, err := g.PlayRepoMan(a, b, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	demandID := action.Demands[0].ID

	keep := bSet.Cards[0]
	give1 := bSet.Cards[1]
	give2 := bSet2.Cards[0]

	// incomplete distribution rejected
	if _, err := g.ComplyRepoManDemand(b, demandID, keep.ID, map[Identifier]uuid.UUID{give1.ID: a}); err != InvalidDistribution {
		t.Fatalf("expected InvalidDistribution, got %v", err)
	}

	// giving a card back to the target rejected
	if _, err := g.ComplyRepoManDemand(b, demandID, keep.ID, map[Identifier]uuid.UUID{give1.ID: a, give2.ID: b}); err != InvalidDistribution {
		t.Fatalf("expected InvalidDistribution, got %v", err)
	}

	res, err := g.ComplyRepoManDemand(b, demandID, keep.ID, map[Identifier]uuid.UUID{give1.ID: a, give2.ID: c})
	if err != nil {
		t.Fatal(err)
	}

	if res.TransferDistribution.KeptCard.ID != keep.ID || len(res.TransferDistribution.Entries) != 2 {
		t.Fatalf("distribution record wrong: %+v", res.TransferDistribution)
	}

	// b keeps exactly the kept card
	bProps := g.Properties[b]
	if len(bProps) != 1 || bProps[0].Cards.Len() != 1 || bProps[0].Cards[0].ID != keep.ID {
		t.Fatalf("b's remaining properties wrong: %+v", bProps)
	}

	// recipients received their cards as properties
	aProps := g.Properties[a]
	if _, j := aProps.IndexByCardID(give1.ID); j == -1 {
		t.Fatal("a did not receive the distributed property")
	}
	cProps := g.Properties[c]
	if _, j := cProps.IndexByCardID(give2.ID); j == -1 {
		t.Fatal("c did not receive the distributed property")
	}

	assertZoneIntegrity(t, g)
}

func TestTaxDayDistribution(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b, c := players[0], players[1], players[2]

	bank := giveBank(t, g, b, AssetKeyMoney1, AssetKeyMoney2, AssetKeyMoney5)

	tax := handCardByKey(t, g, a, AssetKeyTaxDay)
	action, err := g.PlayTaxDay(a, b, tax.ID)
	if err != nil {
		t.Fatal(err)
	}

	keep := bank[2] // keeps the 5M
	res, err := g.ComplyTaxDayDemand(b, action.Demands[0].ID, keep.ID, map[Identifier]uuid.UUID{
		bank[0].ID: a,
		bank[1].ID: c,
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.TransferDistribution.KeptCard.ID != keep.ID {
		t.Fatal("wrong kept card")
	}
	if v, _ := g.CountMoney(b); v != 5 {
		t.Fatalf("b bank %dM, want 5M", v)
	}
	if v, _ := g.CountMoney(a); v != 1 {
		t.Fatalf("a bank %dM, want 1M", v)
	}
	if v, _ := g.CountMoney(c); v != 2 {
		t.Fatalf("c bank %dM, want 2M", v)
	}

	assertZoneIntegrity(t, g)
}

func TestTaxDayRejectedOnEmptyBank(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	tax := handCardByKey(t, g, a, AssetKeyTaxDay)
	if _, err := g.PlayTaxDay(a, b, tax.ID); err != BankIsEmpty {
		t.Fatalf("expected BankIsEmpty, got %v", err)
	}
}

func TestPickpocketStealsCategoryFromHand(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	// make b's hand contents deterministic-ish: count their money cards
	handCardByKey(t, g, b, AssetKeyMoney5)
	wantStolen := 0
	for _, c := range g.Hands[b] {
		if StealCategoryMoney.Matches(c) {
			wantStolen++
		}
	}

	aHandBefore := len(g.Hands[a])

	pick := handCardByKey(t, g, a, AssetKeyPickpocket)
	action, err := g.PlayPickpocket(a, b, pick.ID, StealCategoryMoney)
	if err != nil {
		t.Fatal(err)
	}

	res, err := g.ComplyPickpocketDemand(b, action.Demands[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.TransferPickpocket.Cards) != wantStolen {
		t.Fatalf("stole %d cards, want %d", len(res.TransferPickpocket.Cards), wantStolen)
	}

	for _, c := range g.Hands[b] {
		if StealCategoryMoney.Matches(c) {
			t.Fatal("b still holds money cards")
		}
	}
	if len(g.Hands[a]) != aHandBefore+wantStolen {
		t.Fatal("stolen cards did not arrive in the attacker's hand")
	}

	assertZoneIntegrity(t, g)
}

func TestBankSwap(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	swap := handCardByKey(t, g, a, AssetKeyBankSwap)

	// own empty bank is rejected
	if _, err := g.PlayBankSwap(a, b, swap.ID); err != BankIsEmpty {
		t.Fatalf("expected BankIsEmpty, got %v", err)
	}

	giveBank(t, g, a, AssetKeyMoney1)
	giveBank(t, g, b, AssetKeyMoney10, AssetKeyMoney5)

	action, err := g.PlayBankSwap(a, b, swap.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := g.ComplyBankSwapDemand(b, action.Demands[0].ID); err != nil {
		t.Fatal(err)
	}

	if v, _ := g.CountMoney(a); v != 15 {
		t.Fatalf("a bank %dM, want 15M", v)
	}
	if v, _ := g.CountMoney(b); v != 1 {
		t.Fatalf("b bank %dM, want 1M", v)
	}

	assertZoneIntegrity(t, g)
}

func TestRentAllPlayersCanIssueMultipleDebtChips(t *testing.T) {
	g, players := newTestGame(t, 3)
	a, b, c := players[0], players[1], players[2]

	giveSet(t, g, a, ColorBlue, 2) // complete blue: rent 8

	rent := handCardByKey(t, g, a, AssetKeyRentWild)
	action, err := g.PlayRent(a, rent.ID, ColorBlue)
	if err != nil {
		t.Fatal(err)
	}

	giveBank(t, g, b, AssetKeyMoney1)

	// b surrenders everything and owes a chip
	bDemand := findDemandForTarget(t, action.Demands, b)
	resB, err := g.ComplyPaymentDemand(b, bDemand.ID, g.Money[b][0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if resB.TransferCards.Debt == nil {
		t.Fatal("b should owe a chip")
	}

	// c owns nothing at all and still owes a chip
	cDemand := findDemandForTarget(t, action.Demands, c)
	resC, err := g.ComplyPaymentDemand(c, cDemand.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resC.TransferCards.Debt == nil {
		t.Fatal("c should owe a chip")
	}

	if len(g.Debts) != 2 {
		t.Fatalf("expected 2 obligations, got %d", len(g.Debts))
	}

	assertZoneIntegrity(t, g)
}

func TestDefaultDemandPaysAndDefaultMoveSettles(t *testing.T) {
	g, players := newTestGame(t, 2)
	a, b := players[0], players[1]

	giveSet(t, g, a, ColorBlue, 2) // rent 8
	giveBank(t, g, b, AssetKeyMoney5, AssetKeyMoney4)

	rent := handCardByKey(t, g, a, AssetKeyRentWild)
	action, err := g.PlayRent(a, rent.ID, ColorBlue)
	if err != nil {
		t.Fatal(err)
	}

	// timeout: default demand pays from b's table
	if _, err := g.DefaultDemand(b, action.Demands[0].ID); err != nil {
		t.Fatal(err)
	}
	if v, _ := g.CountMoney(a); v < 8 {
		t.Fatalf("default payment too small: %dM", v)
	}

	if _, err := g.CompleteTurn(a); err != nil {
		t.Fatal(err)
	}

	// give b a debt then let the default move settle it
	g.DebtChips[b]--
	g.Debts = append(g.Debts, DebtObligation{ID: g.IDGenerator.New(), DebtorID: b, CreditorID: a})

	if _, err := g.DefaultMove(b); err != nil {
		t.Fatal(err)
	}
	if len(g.PlayerDebts(b)) != 0 {
		t.Fatal("default move must settle outstanding debts")
	}
	if len(g.Hands[b]) > g.Config.MaxHandSize {
		t.Fatal("default move must discard to the hand limit")
	}

	assertZoneIntegrity(t, g)
}
