package main

import (
	"fmt"
	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"
	"the-deal/internal/schema"
	dnm "the-deal/internal/schema/deal_no_mercy_schema"

	"github.com/google/uuid"
)

// firstHandCardKey returns the id of the first hand card with the given asset
// key for a player, from authoritative state.
func (h *Harness) handCardID(g *GameCtx, pid uuid.UUID, key deal_no_mercy.AssetKey) string {
	game := h.reload(g.gameID)
	for _, c := range game.Hands[pid] {
		if c.AssetKey == key {
			return string(c.ID)
		}
	}
	return ""
}

// demandsFor returns the outstanding demands targeting a player.
func (h *Harness) demandsFor(g *GameCtx, target uuid.UUID) []deal_no_mercy.Demand {
	game := h.reload(g.gameID)
	var out []deal_no_mercy.Demand
	for _, d := range game.Demands {
		if d.TargetID == target {
			out = append(out, d)
		}
	}
	return out
}

func (h *Harness) numDemands(g *GameCtx) int {
	game := h.reload(g.gameID)
	return len(game.Demands)
}

// helpers for calling pointer-receiver engine methods on map values (Go can't
// take the address of a map element, so copy to a local first).
func handLenOf(cards deal_no_mercy.Cards) int    { return cards.Len() }
func bankValueOf(cards deal_no_mercy.Cards) int   { return cards.Value() }
func cardIndex(sets deal_no_mercy.PropertySets, id deal_no_mercy.Identifier) (int, int) {
	return sets.IndexByCardID(id)
}

// waitDemands blocks until the current-player's aggressive play has been
// applied and its demands persisted (or errors after a few tries).
func (h *Harness) awaitDemandsApplied(g *GameCtx, attacker uuid.UUID) bool {
	// The attacker's own socket gets the ActionDemandsCreated echo.
	_, err := awaitAction(g.sockFor(attacker), dnm.ActionKind_ACTION_KIND_DEMAND_CREATED)
	return err == nil
}

// scenarioMechanics runs at least one full flow of every new mechanic.
func (h *Harness) scenarioMechanics(g *GameCtx) {
	h.mechShack(g)
	h.mechBigPayday(g)
	h.mechGoAgain(g)
	h.mechYoink(g)
	h.mechRentAllDoubleVariant(g)
	h.mechHeist(g)
	h.mechMarketCrash(g)
	h.mechPropertyRaid(g)
	h.mechSetSnatcher(g)
	h.mechTaxDay(g)
	h.mechPickpocket(g)
	h.mechBankSwap(g)
	h.mechDebtTrap(g)
	h.mechNahCounter(g)
	h.mechDebtChipAndSettle(g)
}

// prepareTurn resets the game so `attacker` is on turn with full moves and no
// demands. Returns attacker and target ids.
func (h *Harness) prepareTurn(g *GameCtx) (attacker, target uuid.UUID) {
	attacker = h.currentPlayer(g)
	target = g.other(attacker)
	return attacker, target
}

// mechShack: attach a shack to a set and verify the +rent effect through a rent
// demand amount.
func (h *Harness) mechShack(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	var setID string
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.clearProperties(att)
		b.clearProperties(tgt)
		// attacker: complete brown set (2 cards) so rent is 2M, plus a shack in hand.
		set := b.addSet(att, deal_no_mercy.ColorBrown, false, deal_no_mercy.AssetKeyBrown1, deal_no_mercy.AssetKeyBrown2)
		setID = string(set.ID)
		b.setHand(att, deal_no_mercy.AssetKeyShack, deal_no_mercy.AssetKeyRentBrownSky)
		// give target plenty of bank to pay rent.
		b.setBank(tgt, deal_no_mercy.AssetKeyMoney10, deal_no_mercy.AssetKeyMoney5)
	})
	g.resync(h)

	shackID := h.handCardID(g, att, deal_no_mercy.AssetKeyShack)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayShack{
		PlayShack: &dnm.PlayShack{CardId: shackID, PropertySetId: setID},
	}}), "send shack")
	_, err := awaitAction(g.sockFor(att), dnm.ActionKind_ACTION_KIND_PLAY_SHACK)
	// Verify shack attached in authoritative state.
	game := h.reload(g.gameID)
	shackOK := false
	for _, s := range game.Properties[att] {
		if string(s.ID) == setID && s.HasShack() {
			shackOK = true
		}
	}
	h.record("shack attach", err == nil && shackOK, boolStr(shackOK, "shack on brown set", errStr(err)))

	// Now play rent brown/sky; the shack bonus (5) should raise brown rent 2 -> 7.
	rentID := h.handCardID(g, att, deal_no_mercy.AssetKeyRentBrownSky)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayRent{
		PlayRent: &dnm.PlayRent{CardId: rentID, Color: dnm.Color_COLOR_BROWN},
	}}), "send rent")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	rentOK := len(ds) == 1 && ds[0].Kind == deal_no_mercy.DemandKindPayment && ds[0].Payment.Amount == 7
	amt := -1
	if len(ds) == 1 {
		amt = ds[0].Payment.Amount
	}
	h.record("shack +rent effect (rent 2->7)", rentOK, fmt.Sprintf("payment demand amount=%d (want 7)", amt))

	// Comply with the rent to clear the demand.
	if len(ds) == 1 {
		h.payDemand(g, tgt, ds[0])
	}
	h.endTurn(g, att)
}

// payDemand pays a payment demand in full using the target's bank cards.
func (h *Harness) payDemand(g *GameCtx, target uuid.UUID, d deal_no_mercy.Demand) {
	game := h.reload(g.gameID)
	need := d.Payment.Amount
	var ids []string
	sum := 0
	for _, c := range game.Money[target] {
		if sum >= need {
			break
		}
		ids = append(ids, string(c.ID))
		sum += c.Value
	}
	must(g.send(target, &dnm.ClientMessage{Payload: &dnm.ClientMessage_ComplyPaymentDemand{
		ComplyPaymentDemand: &dnm.ComplyPaymentDemand{DemandId: string(d.ID), CardIds: ids},
	}}), "send comply payment")
	_, _ = awaitAction(g.sockFor(target), dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED)
	g.sockFor(g.other(target)).drain()
}

// endTurn forces a clean turn boundary: overwrite moves to 0, empty extra
// hand, then complete turn for the attacker so the loop can continue.
func (h *Harness) endTurn(g *GameCtx, pid uuid.UUID) {
	h.writeState(g, func(b *stateBuilder) {
		b.g.MovesLeft = 0
		// trim hand to <= max so complete turn is accepted; leave debts alone.
		hand := b.g.Hands[pid]
		if hand.Len() > b.g.Config.MaxHandSize {
			b.g.Hands[pid] = hand[:b.g.Config.MaxHandSize]
		}
	})
	g.resync(h)
	// only complete if no outstanding demands/debts
	game := h.reload(g.gameID)
	if len(game.Demands) != 0 || len(game.PlayerDebts(pid)) != 0 {
		return
	}
	_ = g.send(pid, &dnm.ClientMessage{Payload: &dnm.ClientMessage_CompleteTurn{CompleteTurn: &dnm.CompleteTurn{}}})
	g.sockA.drain()
	g.sockB.drain()
}

// mechBigPayday: draw to hand target (7).
func (h *Harness) mechBigPayday(g *GameCtx) {
	att := h.currentPlayer(g)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		// hand: just the big payday card (1 card) => should draw up to 7.
		b.setHand(att, deal_no_mercy.AssetKeyBigPayday)
		// ensure deck has enough cards on top.
		b.setDeckTop(
			deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1,
			deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1,
			deal_no_mercy.AssetKeyMoney1,
		)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyBigPayday)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayBigPayday{
		PlayBigPayday: &dnm.PlayBigPayday{CardId: cardID},
	}}), "send big payday")
	_, err := awaitAction(g.sockFor(att), dnm.ActionKind_ACTION_KIND_PLAY_BIG_PAYDAY)
	game := h.reload(g.gameID)
	handLen := handLenOf(game.Hands[att])
	h.record("big payday (draw to 7)", err == nil && handLen == 7, fmt.Sprintf("hand=%d (want 7)", handLen))
	h.endTurn(g, att)
}

// mechGoAgain: play go again as last card, verify same player takes next turn.
func (h *Harness) mechGoAgain(g *GameCtx) {
	att := h.currentPlayer(g)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeyGoAgain)
		b.setDeckTop(deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyGoAgain)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayGoAgain{
		PlayGoAgain: &dnm.PlayGoAgain{CardId: cardID},
	}}), "send go again")
	_, err := awaitAction(g.sockFor(att), dnm.ActionKind_ACTION_KIND_PLAY_GO_AGAIN)
	// GoAgainQueued should be set; complete turn keeps the same current player.
	game := h.reload(g.gameID)
	queued := game.GoAgainQueued
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_CompleteTurn{CompleteTurn: &dnm.CompleteTurn{}}}), "complete after go again")
	_, _ = awaitAction(g.sockFor(att), dnm.ActionKind_ACTION_KIND_START_TURN)
	after := h.currentPlayer(g)
	h.record("go again (same player next turn)", err == nil && queued && after == att,
		fmt.Sprintf("queued=%v currentAfter=%s want=%s", queued, short(after.String()), short(att.String())))
	h.endTurn(g, att)
}

// mechYoink: payment demand for the yoink amount (10 default).
func (h *Harness) mechYoink(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeyYoink)
		b.setBank(tgt, deal_no_mercy.AssetKeyMoney10, deal_no_mercy.AssetKeyMoney5)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyYoink)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayYoink{
		PlayYoink: &dnm.PlayYoink{CardId: cardID, TargetId: tgt.String()},
	}}), "send yoink")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	ok := len(ds) == 1 && ds[0].Kind == deal_no_mercy.DemandKindPayment && ds[0].Payment.Amount == 10
	amt := -1
	if len(ds) == 1 {
		amt = ds[0].Payment.Amount
	}
	h.record("yoink (payment demand = 10)", ok, fmt.Sprintf("amount=%d (want 10)", amt))
	if len(ds) == 1 {
		h.payDemand(g, tgt, ds[0])
	}
	h.endTurn(g, att)
}

// mechRentAllDoubleVariant: double rent charges all opponents at 2x.
func (h *Harness) mechRentAllDoubleVariant(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.clearProperties(att)
		// complete brown set: base rent 2, double => 4.
		b.addSet(att, deal_no_mercy.ColorBrown, false, deal_no_mercy.AssetKeyBrown1, deal_no_mercy.AssetKeyBrown2)
		b.setHand(att, deal_no_mercy.AssetKeyDoubleRentBrownSky)
		b.setBank(tgt, deal_no_mercy.AssetKeyMoney5, deal_no_mercy.AssetKeyMoney5)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyDoubleRentBrownSky)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayRent{
		PlayRent: &dnm.PlayRent{CardId: cardID, Color: dnm.Color_COLOR_BROWN},
	}}), "send double rent")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	ok := len(ds) == 1 && ds[0].Payment.Amount == 4
	amt := -1
	if len(ds) == 1 {
		amt = ds[0].Payment.Amount
	}
	h.record("rent-all double variant (2x, charges opponent)", ok, fmt.Sprintf("amount=%d (want 4=2x2)", amt))
	if len(ds) == 1 {
		h.payDemand(g, tgt, ds[0])
	}
	h.endTurn(g, att)
}

// mechHeist: attacker names one bank card per opponent; target complies.
func (h *Harness) mechHeist(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	var targetCardID string
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeyHeist)
		bank := b.setBank(tgt, deal_no_mercy.AssetKeyMoney5, deal_no_mercy.AssetKeyMoney2)
		targetCardID = string(bank[0].ID) // the 5M
		b.setBank(att)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyHeist)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayHeist{
		PlayHeist: &dnm.PlayHeist{CardId: cardID, Picks: []*dnm.PropertyPick{
			{TargetId: tgt.String(), CardId: targetCardID},
		}},
	}}), "send heist")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	createdOK := len(ds) == 1 && ds[0].Kind == deal_no_mercy.DemandKindBankCard
	// comply
	complied := false
	if createdOK {
		must(g.send(tgt, &dnm.ClientMessage{Payload: &dnm.ClientMessage_ComplyBankCardDemand{
			ComplyBankCardDemand: &dnm.ComplyBankCardDemand{DemandId: string(ds[0].ID)},
		}}), "comply heist")
		_, err := awaitAction(g.sockFor(tgt), dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED)
		g.sockFor(att).drain()
		game := h.reload(g.gameID)
		// the 5M should now be in attacker's bank
		for _, c := range game.Money[att] {
			if string(c.ID) == targetCardID {
				complied = true
			}
		}
		_ = err
	}
	h.record("heist (multi-target bank card pick + comply)", createdOK && complied, boolStr(createdOK && complied, "5M moved to attacker bank", "demand/comply failed"))
	h.endTurn(g, att)
}

// mechMarketCrash: steal a property (may be from a complete set); comply.
func (h *Harness) mechMarketCrash(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	var targetCardID string
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeyMarketCrash)
		b.clearProperties(tgt)
		set := b.addSet(tgt, deal_no_mercy.ColorBrown, false, deal_no_mercy.AssetKeyBrown1, deal_no_mercy.AssetKeyBrown2)
		targetCardID = string(set.Cards[0].ID)
		b.clearProperties(att)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyMarketCrash)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayMarketCrash{
		PlayMarketCrash: &dnm.PlayMarketCrash{CardId: cardID, Picks: []*dnm.PropertyPick{
			{TargetId: tgt.String(), CardId: targetCardID},
		}},
	}}), "send market crash")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	ok := len(ds) == 1 && ds[0].Kind == deal_no_mercy.DemandKindProperty
	moved := false
	if ok {
		must(g.send(tgt, &dnm.ClientMessage{Payload: &dnm.ClientMessage_ComplyPropertyDemand{
			ComplyPropertyDemand: &dnm.ComplyPropertyDemand{DemandId: string(ds[0].ID)},
		}}), "comply market crash")
		_, _ = awaitAction(g.sockFor(tgt), dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED)
		g.sockFor(att).drain()
		game := h.reload(g.gameID)
		i, j := cardIndex(game.Properties[att], deal_no_mercy.Identifier(targetCardID))
		moved = i != -1 && j != -1
	}
	h.record("market crash (multi-target property pick + comply)", ok && moved, boolStr(ok && moved, "property moved to attacker", "failed"))
	h.endTurn(g, att)
}

// mechPropertyRaid: steal all sets of a color; comply.
func (h *Harness) mechPropertyRaid(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeyPropertyRaid)
		b.clearProperties(tgt)
		b.addSet(tgt, deal_no_mercy.ColorPink, false, deal_no_mercy.AssetKeyPink1, deal_no_mercy.AssetKeyPink2)
		b.clearProperties(att)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyPropertyRaid)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayPropertyRaid{
		PlayPropertyRaid: &dnm.PlayPropertyRaid{CardId: cardID, Color: dnm.Color_COLOR_PINK},
	}}), "send property raid")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	ok := len(ds) == 1 && ds[0].Kind == deal_no_mercy.DemandKindColorProperties
	moved := false
	if ok {
		must(g.send(tgt, &dnm.ClientMessage{Payload: &dnm.ClientMessage_ComplyColorPropertiesDemand{
			ComplyColorPropertiesDemand: &dnm.ComplyColorPropertiesDemand{DemandId: string(ds[0].ID)},
		}}), "comply property raid")
		_, _ = awaitAction(g.sockFor(tgt), dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED)
		g.sockFor(att).drain()
		game := h.reload(g.gameID)
		cnt := 0
		for _, s := range game.Properties[att] {
			if s.Color == deal_no_mercy.ColorPink {
				cnt += s.Cards.Len()
			}
		}
		moved = cnt == 2
	}
	h.record("property raid (steal whole color) + comply", ok && moved, boolStr(ok && moved, "pink set moved to attacker", "failed"))
	h.endTurn(g, att)
}

// mechSetSnatcher: steal a complete set; comply.
func (h *Harness) mechSetSnatcher(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	var setID string
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeySetSnatcher)
		b.clearProperties(tgt)
		set := b.addSet(tgt, deal_no_mercy.ColorBrown, false, deal_no_mercy.AssetKeyBrown1, deal_no_mercy.AssetKeyBrown2)
		setID = string(set.ID)
		b.clearProperties(att)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeySetSnatcher)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlaySetSnatcher{
		PlaySetSnatcher: &dnm.PlaySetSnatcher{CardId: cardID, TargetId: tgt.String(), PropertySetId: setID},
	}}), "send set snatcher")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	ok := len(ds) == 1 && ds[0].Kind == deal_no_mercy.DemandKindPropertySet
	moved := false
	if ok {
		must(g.send(tgt, &dnm.ClientMessage{Payload: &dnm.ClientMessage_ComplyPropertySetDemand{
			ComplyPropertySetDemand: &dnm.ComplyPropertySetDemand{DemandId: string(ds[0].ID)},
		}}), "comply set snatcher")
		_, _ = awaitAction(g.sockFor(tgt), dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED)
		g.sockFor(att).drain()
		game := h.reload(g.gameID)
		for _, s := range game.Properties[att] {
			if s.Color == deal_no_mercy.ColorBrown && s.Cards.Len() == 2 {
				moved = true
			}
		}
	}
	h.record("set snatcher (steal complete set) + comply", ok && moved, boolStr(ok && moved, "brown set moved to attacker", "failed"))
	h.endTurn(g, att)
}

// mechTaxDay: target keeps one card, distributes the rest (comply).
func (h *Harness) mechTaxDay(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeyTaxDay)
		b.clearProperties(tgt)
		b.setBank(tgt, deal_no_mercy.AssetKeyMoney5, deal_no_mercy.AssetKeyMoney2)
		b.clearProperties(att)
		b.setBank(att)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyTaxDay)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayTaxDay{
		PlayTaxDay: &dnm.PlayTaxDay{CardId: cardID, TargetId: tgt.String()},
	}}), "send tax day")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	ok := len(ds) == 1 && ds[0].Kind == deal_no_mercy.DemandKindTaxDay
	distributed := false
	if ok {
		game := h.reload(g.gameID)
		bank := game.Money[tgt]
		// keep the first bank card, distribute the rest to the attacker.
		keep := string(bank[0].ID)
		var dist []*dnm.DistributionAssignment
		for _, c := range bank[1:] {
			dist = append(dist, &dnm.DistributionAssignment{CardId: string(c.ID), RecipientId: att.String()})
		}
		must(g.send(tgt, &dnm.ClientMessage{Payload: &dnm.ClientMessage_ComplyTaxDayDemand{
			ComplyTaxDayDemand: &dnm.ComplyTaxDayDemand{DemandId: string(ds[0].ID), KeepCardId: keep, Distribution: dist},
		}}), "comply tax day")
		_, err := awaitAction(g.sockFor(tgt), dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED)
		g.sockFor(att).drain()
		after := h.reload(g.gameID)
		distributed = err == nil && bankValueOf(after.Money[att]) == 2 && bankValueOf(after.Money[tgt]) == 5
	}
	h.record("tax day distribution comply", ok && distributed, boolStr(ok && distributed, "kept 5M, gave 2M to attacker", "failed"))
	h.endTurn(g, att)
}

// mechPickpocket: steal all hand cards of a category; verify attacker & target
// see the real transfer (2-player: no bystander masking to test, but confirm
// the stolen card actually moved hand->hand).
func (h *Harness) mechPickpocket(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	var stolenID string
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeyPickpocket)
		// target hand has one money card to be pickpocketed.
		hand := b.setHand(tgt, deal_no_mercy.AssetKeyMoney3, deal_no_mercy.AssetKeyBrown1)
		stolenID = string(hand[0].ID) // the money3
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyPickpocket)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayPickpocket{
		PlayPickpocket: &dnm.PlayPickpocket{CardId: cardID, TargetId: tgt.String(), Category: dnm.StealCategory_STEAL_CATEGORY_MONEY},
	}}), "send pickpocket")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	ok := len(ds) == 1 && ds[0].Kind == deal_no_mercy.DemandKindPickpocket
	moved := false
	if ok {
		must(g.send(tgt, &dnm.ClientMessage{Payload: &dnm.ClientMessage_ComplyPickpocketDemand{
			ComplyPickpocketDemand: &dnm.ComplyPickpocketDemand{DemandId: string(ds[0].ID)},
		}}), "comply pickpocket")
		_, _ = awaitAction(g.sockFor(tgt), dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED)
		g.sockFor(att).drain()
		game := h.reload(g.gameID)
		for _, c := range game.Hands[att] {
			if string(c.ID) == stolenID {
				moved = true
			}
		}
	}
	h.record("pickpocket (hand->hand, category money) + comply", ok && moved, boolStr(ok && moved, "3M moved from target hand to attacker hand", "failed"))
	h.endTurn(g, att)
}

// mechBankSwap: attacker swaps banks with target; comply.
func (h *Harness) mechBankSwap(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeyBankSwap)
		b.setBank(att, deal_no_mercy.AssetKeyMoney1)
		b.setBank(tgt, deal_no_mercy.AssetKeyMoney10)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyBankSwap)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayBankSwap{
		PlayBankSwap: &dnm.PlayBankSwap{CardId: cardID, TargetId: tgt.String()},
	}}), "send bank swap")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	ok := len(ds) == 1 && ds[0].Kind == deal_no_mercy.DemandKindBankSwap
	swapped := false
	if ok {
		must(g.send(tgt, &dnm.ClientMessage{Payload: &dnm.ClientMessage_ComplyBankSwapDemand{
			ComplyBankSwapDemand: &dnm.ComplyBankSwapDemand{DemandId: string(ds[0].ID)},
		}}), "comply bank swap")
		_, _ = awaitAction(g.sockFor(tgt), dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED)
		g.sockFor(att).drain()
		game := h.reload(g.gameID)
		swapped = bankValueOf(game.Money[att]) == 10 && bankValueOf(game.Money[tgt]) == 1
	}
	h.record("bank swap + comply", ok && swapped, boolStr(ok && swapped, "banks swapped (10<->1)", "failed"))
	h.endTurn(g, att)
}

// mechDebtTrap: attacker plays debt trap; comply issues a debt chip.
func (h *Harness) mechDebtTrap(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeyDebtTrap)
	})
	g.resync(h)
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyDebtTrap)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayDebtTrap{
		PlayDebtTrap: &dnm.PlayDebtTrap{CardId: cardID, TargetId: tgt.String()},
	}}), "send debt trap")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	ok := len(ds) == 1 && ds[0].Kind == deal_no_mercy.DemandKindDebtTrap
	trapped := false
	if ok {
		must(g.send(tgt, &dnm.ClientMessage{Payload: &dnm.ClientMessage_ComplyDebtTrapDemand{
			ComplyDebtTrapDemand: &dnm.ComplyDebtTrapDemand{DemandId: string(ds[0].ID)},
		}}), "comply debt trap")
		_, _ = awaitAction(g.sockFor(tgt), dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED)
		g.sockFor(att).drain()
		game := h.reload(g.gameID)
		trapped = len(game.PlayerDebts(tgt)) == 1
	}
	h.record("debt trap issues obligation on comply", ok && trapped, boolStr(ok && trapped, "target now owes 1 debt", "failed"))
	// Clear the debt so the loop continues cleanly: give target a hand card and
	// have them settle at the start of their turn (also exercises settlement).
	h.endTurn(g, att)
	h.settleTargetDebt(g, tgt)
}

// settleTargetDebt: at the target's own turn, settle any outstanding debt with
// a hand card (mandatory settlement consuming a play).
func (h *Harness) settleTargetDebt(g *GameCtx, debtor uuid.UUID) {
	game := h.reload(g.gameID)
	if len(game.PlayerDebts(debtor)) == 0 {
		return
	}
	// Make it the debtor's turn with a hand card to settle with.
	h.writeState(g, func(b *stateBuilder) {
		b.setTurnKeepDemands(debtor)
		if handLenOf(b.g.Hands[debtor]) == 0 {
			b.setHand(debtor, deal_no_mercy.AssetKeyMoney1)
		}
	})
	g.resync(h)
	game = h.reload(g.gameID)
	debts := game.PlayerDebts(debtor)
	if len(debts) == 0 {
		return
	}
	// A regular play should be blocked while debt is outstanding.
	handID := string(game.Hands[debtor][0].ID)
	must(g.send(debtor, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayMoney{
		PlayMoney: &dnm.PlayMoney{CardId: handID},
	}}), "attempt play under debt")
	blockErr, _ := awaitError(g.sockFor(debtor))
	h.record("mandatory settlement blocks other plays", blockErr != nil, boolStr(blockErr != nil, msgOf(blockErr), "play was NOT blocked"))

	// Settle the debt (consumes a play).
	must(g.send(debtor, &dnm.ClientMessage{Payload: &dnm.ClientMessage_SettleDebt{
		SettleDebt: &dnm.SettleDebt{DebtId: string(debts[0].ID), CardId: handID},
	}}), "settle debt")
	_, err := awaitAction(g.sockFor(debtor), dnm.ActionKind_ACTION_KIND_DEBT_SETTLED)
	after := h.reload(g.gameID)
	settled := err == nil && len(after.PlayerDebts(debtor)) == 0
	h.record("debt settlement consumes a play (chip returned)", settled, boolStr(settled, "debt cleared", errStr(err)))
	h.endTurn(g, debtor)
}

// mechNahCounter: attacker plays yoink, target denies with NAH!, and the demand
// is re-issued back at the attacker (deny + counter-flow).
func (h *Harness) mechNahCounter(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.setHand(att, deal_no_mercy.AssetKeyYoink)
		b.setHand(tgt, deal_no_mercy.AssetKeyNah)
		b.setBank(att, deal_no_mercy.AssetKeyMoney10, deal_no_mercy.AssetKeyMoney5)
		b.setBank(tgt, deal_no_mercy.AssetKeyMoney5)
	})
	g.resync(h)
	yoinkID := h.handCardID(g, att, deal_no_mercy.AssetKeyYoink)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayYoink{
		PlayYoink: &dnm.PlayYoink{CardId: yoinkID, TargetId: tgt.String()},
	}}), "send yoink (nah)")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	if len(ds) != 1 {
		h.record("NAH! deny + counter-flow", false, "no demand to deny")
		return
	}
	nahID := h.handCardID(g, tgt, deal_no_mercy.AssetKeyNah)
	origDemandID := string(ds[0].ID)
	// Drain the target socket so the only response we observe is the deny's.
	g.sockFor(tgt).drain()
	if err := g.send(tgt, &dnm.ClientMessage{Payload: &dnm.ClientMessage_DenyDemand{
		DenyDemand: &dnm.DenyDemand{DemandId: origDemandID, CardId: nahID},
	}}); err != nil {
		must(err, "send nah")
	}
	// The deny either echoes a DEMAND_CREATED (success) or an Error (rejection).
	m, werr := g.sockFor(tgt).nextMatching(waitT, func(sm *schema.ServerMessage) bool {
		a := dnmAction(sm)
		return (a != nil && a.GetKind() == dnm.ActionKind_ACTION_KIND_DEMAND_CREATED) || dnmErr(sm) != nil
	})
	denyErr := ""
	if werr == nil {
		if e := dnmErr(m); e != nil {
			denyErr = e.GetMessage()
		}
	}
	// After deny, the demand is re-issued targeting the original attacker
	// (source/target swapped). Find any active demand not equal to the original
	// id that now targets att.
	reissued := h.demandsFor(g, att)
	ok := denyErr == "" && len(reissued) == 1 && reissued[0].Kind == deal_no_mercy.DemandKindPayment && string(reissued[0].ID) != origDemandID
	detail := "denial re-issued the payment demand at the original attacker (inactive; counter-flow)"
	if !ok {
		detail = fmt.Sprintf("denyErr=%q reissuedToAtt=%d", trunc(denyErr, 40), len(reissued))
	}
	h.record("NAH! deny re-issues demand at attacker", ok, detail)
	// Attacker now pays (counter-flow resolves).
	if ok {
		h.payDemand(g, att, reissued[0])
	}
	h.endTurn(g, att)
}

// mechDebtChipAndSettle: short payment issues a debt chip, and the mandatory
// settlement next turn consumes a play. (Covers debt chip issue on short
// payment; settlement path is also exercised in mechDebtTrap.)
func (h *Harness) mechDebtChipAndSettle(g *GameCtx) {
	att, tgt := h.prepareTurn(g)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.clearProperties(att)
		// attacker completes a green set (base rent 4).
		b.addSet(att, deal_no_mercy.ColorGreen, false,
			deal_no_mercy.AssetKeyGreen1, deal_no_mercy.AssetKeyGreen2, deal_no_mercy.AssetKeyGreen3)
		b.setHand(att, deal_no_mercy.AssetKeyRentGreenBlue)
		// target can only pay 1M against a 4 rent => short by 3 => chip issued.
		b.setBank(tgt, deal_no_mercy.AssetKeyMoney1)
		b.clearProperties(tgt)
	})
	g.resync(h)
	rentID := h.handCardID(g, att, deal_no_mercy.AssetKeyRentGreenBlue)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayRent{
		PlayRent: &dnm.PlayRent{CardId: rentID, Color: dnm.Color_COLOR_GREEN},
	}}), "send rent for short pay")
	h.awaitDemandsApplied(g, att)
	ds := h.demandsFor(g, tgt)
	if len(ds) != 1 {
		h.record("debt chip on short payment", false, "no rent demand created")
		return
	}
	// Pay everything available (1M) — still short, so a chip is issued.
	game := h.reload(g.gameID)
	var ids []string
	for _, c := range game.Money[tgt] {
		ids = append(ids, string(c.ID))
	}
	must(g.send(tgt, &dnm.ClientMessage{Payload: &dnm.ClientMessage_ComplyPaymentDemand{
		ComplyPaymentDemand: &dnm.ComplyPaymentDemand{DemandId: string(ds[0].ID), CardIds: ids},
	}}), "comply short payment")
	_, err := awaitAction(g.sockFor(tgt), dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED)
	g.sockFor(att).drain()
	after := h.reload(g.gameID)
	chipIssued := err == nil && len(after.PlayerDebts(tgt)) == 1 && after.DebtChips[tgt] == 2
	h.record("debt chip issued on short payment", chipIssued,
		fmt.Sprintf("debts=%d availableChips=%d (want 1 debt, 2 chips)", len(after.PlayerDebts(tgt)), after.DebtChips[tgt]))
	h.endTurn(g, att)
	// Mandatory settlement next turn.
	h.settleTargetDebt(g, tgt)
}
