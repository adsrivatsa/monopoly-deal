package deal_no_mercy

import (
	"slices"

	"the-deal/internal/errors"

	"github.com/google/uuid"
)

// Every aggressive action funnels through the demand system so that NAH!
// can deny it and the service layer can auto-resolve timeouts.
// Multi-target actions (heist, market crash, property raid, rent) create
// one demand per target. Where the ATTACKER chooses (heist bank card,
// market crash property), the choice is made at play time and carried in
// the demand; where the TARGET chooses (repo man, tax day), the choice is
// carried in the compliance call.

// PlaySetSnatcher steals a complete set (no houses/hotels exist in this
// game; an attached shack travels with the set).
func (g *Game) PlaySetSnatcher(playerID, targetID uuid.UUID, cardID, setID Identifier) (*ActionDemandsCreated, error) {
	card, err := g.checkTargetedActionPlay(playerID, targetID, cardID, AssetKeySetSnatcher)
	if err != nil {
		return nil, err
	}

	targetProperties := g.Properties[targetID]
	setIdx := targetProperties.IndexBySetID(setID)
	if setIdx == -1 {
		return nil, errors.PropertySetDoesntExist
	}
	if !targetProperties[setIdx].IsComplete() {
		return nil, errors.PropertySetIsNotComplete
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewPropertySetDemand(demandID, playerID, targetID, setID, DemandSourceSetSnatcher)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayDebtTrap takes a debt chip from the target (deniable by NAH!): once
// complied, they owe the player a settlement at the start of their next
// turn, exactly like a failed payment.
func (g *Game) PlayDebtTrap(playerID, targetID uuid.UUID, cardID Identifier) (*ActionDemandsCreated, error) {
	card, err := g.checkTargetedActionPlay(playerID, targetID, cardID, AssetKeyDebtTrap)
	if err != nil {
		return nil, err
	}

	if g.DebtChips[targetID] <= 0 {
		return nil, NoDebtChipsAvailable
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewDebtTrapDemand(demandID, playerID, targetID)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayHeist steals one attacker-chosen bank card from EVERY opponent.
// picks maps each opponent with a non-empty bank to the chosen card; every
// such opponent must be picked, opponents with empty banks must not be.
func (g *Game) PlayHeist(playerID uuid.UUID, cardID Identifier, picks map[uuid.UUID]Identifier) (*ActionDemandsCreated, error) {
	card, err := g.checkActionPlay(playerID, cardID, AssetKeyHeist)
	if err != nil {
		return nil, err
	}

	eligible := make([]uuid.UUID, 0, len(g.Players)-1)
	for _, targetID := range g.Players {
		if targetID == playerID {
			continue
		}
		if len(g.Money[targetID]) > 0 {
			eligible = append(eligible, targetID)
		}
	}

	if len(eligible) == 0 {
		return nil, NoValidTargets
	}

	if len(picks) != len(eligible) {
		return nil, InvalidTargetPicks
	}

	for _, targetID := range eligible {
		pickID, ok := picks[targetID]
		if !ok {
			return nil, InvalidTargetPicks
		}

		bank := g.Money[targetID]
		found := false
		for _, c := range bank {
			if c.ID == pickID {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.PlayerDoesNotHaveCard
		}
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandsMap := make(map[Identifier]Demand)
	demandsSlice := make([]Demand, 0, len(eligible))
	for _, targetID := range eligible {
		id := g.IDGenerator.New()
		demand := NewBankCardDemand(id, playerID, targetID, picks[targetID], DemandSourceHeist)
		demandsMap[id] = demand
		demandsSlice = append(demandsSlice, demand)
	}
	g.Demands = demandsMap

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demandsSlice...)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayMarketCrash steals one attacker-chosen property from EVERY opponent.
// Complete sets are NOT protected — this is No Mercy (see engine-notes.md).
// picks maps each opponent owning at least one property to the chosen card.
func (g *Game) PlayMarketCrash(playerID uuid.UUID, cardID Identifier, picks map[uuid.UUID]Identifier) (*ActionDemandsCreated, error) {
	card, err := g.checkActionPlay(playerID, cardID, AssetKeyMarketCrash)
	if err != nil {
		return nil, err
	}

	eligible := make([]uuid.UUID, 0, len(g.Players)-1)
	for _, targetID := range g.Players {
		if targetID == playerID {
			continue
		}

		hasProperty := false
		for _, set := range g.Properties[targetID] {
			if set.Cards.Len() > 0 {
				hasProperty = true
				break
			}
		}
		if hasProperty {
			eligible = append(eligible, targetID)
		}
	}

	if len(eligible) == 0 {
		return nil, NoValidTargets
	}

	if len(picks) != len(eligible) {
		return nil, InvalidTargetPicks
	}

	for _, targetID := range eligible {
		pickID, ok := picks[targetID]
		if !ok {
			return nil, InvalidTargetPicks
		}

		pick, err := g.checkCard(pickID, CategoryPureProperty, CategoryWildProperty)
		if err != nil {
			return nil, err
		}

		targetProperties := g.Properties[targetID]
		i, j := targetProperties.IndexByCardID(pick.ID)
		if i == -1 || j == -1 {
			return nil, errors.PlayerDoesNotHaveCard
		}
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandsMap := make(map[Identifier]Demand)
	demandsSlice := make([]Demand, 0, len(eligible))
	for _, targetID := range eligible {
		id := g.IDGenerator.New()
		demand := NewPropertyDemand(id, playerID, targetID, picks[targetID], DemandSourceMarketCrash)
		demandsMap[id] = demand
		demandsSlice = append(demandsSlice, demand)
	}
	g.Demands = demandsMap

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demandsSlice...)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayRepoMan targets one player: they keep 1 property of their choice and
// distribute the rest to players of their choice (choices carried in the
// compliance call).
func (g *Game) PlayRepoMan(playerID, targetID uuid.UUID, cardID Identifier) (*ActionDemandsCreated, error) {
	card, err := g.checkTargetedActionPlay(playerID, targetID, cardID, AssetKeyRepoMan)
	if err != nil {
		return nil, err
	}

	hasProperty := false
	for _, set := range g.Properties[targetID] {
		if set.Cards.Len() > 0 {
			hasProperty = true
			break
		}
	}
	if !hasProperty {
		return nil, NoValidTargets
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewRepoManDemand(demandID, playerID, targetID)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayPropertyRaid names a color and steals ALL sets of that color from
// every opponent — complete sets included, shacks travel with their sets.
func (g *Game) PlayPropertyRaid(playerID uuid.UUID, cardID Identifier, color Color) (*ActionDemandsCreated, error) {
	card, err := g.checkActionPlay(playerID, cardID, AssetKeyPropertyRaid)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(AllColors(), color) {
		return nil, InvalidColorForCard
	}

	eligible := make([]uuid.UUID, 0, len(g.Players)-1)
	for _, targetID := range g.Players {
		if targetID == playerID {
			continue
		}

		for _, set := range g.Properties[targetID] {
			if set.Color == color && set.Cards.Len() > 0 {
				eligible = append(eligible, targetID)
				break
			}
		}
	}

	if len(eligible) == 0 {
		return nil, NoValidTargets
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandsMap := make(map[Identifier]Demand)
	demandsSlice := make([]Demand, 0, len(eligible))
	for _, targetID := range eligible {
		id := g.IDGenerator.New()
		demand := NewColorPropertiesDemand(id, playerID, targetID, color, DemandSourcePropertyRaid)
		demandsMap[id] = demand
		demandsSlice = append(demandsSlice, demand)
	}
	g.Demands = demandsMap

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demandsSlice...)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayTaxDay targets one player: they keep 1 bank card of their choice and
// distribute the rest to players of their choice (choices carried in the
// compliance call).
func (g *Game) PlayTaxDay(playerID, targetID uuid.UUID, cardID Identifier) (*ActionDemandsCreated, error) {
	card, err := g.checkTargetedActionPlay(playerID, targetID, cardID, AssetKeyTaxDay)
	if err != nil {
		return nil, err
	}

	if len(g.Money[targetID]) == 0 {
		return nil, BankIsEmpty
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewTaxDayDemand(demandID, playerID, targetID)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayPickpocket names a category (property/money/action) and steals every
// hand card of that category from one target.
func (g *Game) PlayPickpocket(playerID, targetID uuid.UUID, cardID Identifier, category StealCategory) (*ActionDemandsCreated, error) {
	card, err := g.checkTargetedActionPlay(playerID, targetID, cardID, AssetKeyPickpocket)
	if err != nil {
		return nil, err
	}

	if category != StealCategoryProperty && category != StealCategoryMoney && category != StealCategoryAction {
		return nil, InvalidStealCategory
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewPickpocketDemand(demandID, playerID, targetID, category)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayBankSwap swaps entire banks with the target. Illegal when the
// player's own bank is empty.
func (g *Game) PlayBankSwap(playerID, targetID uuid.UUID, cardID Identifier) (*ActionDemandsCreated, error) {
	card, err := g.checkTargetedActionPlay(playerID, targetID, cardID, AssetKeyBankSwap)
	if err != nil {
		return nil, err
	}

	if len(g.Money[playerID]) == 0 {
		return nil, BankIsEmpty
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewBankSwapDemand(demandID, playerID, targetID)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayYoink demands Config.YoinkPayment (10M) from one target; no change
// given, debt chip on shortfall.
func (g *Game) PlayYoink(playerID, targetID uuid.UUID, cardID Identifier) (*ActionDemandsCreated, error) {
	card, err := g.checkTargetedActionPlay(playerID, targetID, cardID, AssetKeyYoink)
	if err != nil {
		return nil, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewPaymentDemand(demandID, playerID, targetID, g.Config.YoinkPayment, DemandSourceYoink)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayRent charges ALL opponents rent in the chosen color. Works for the
// five two-color rent cards, rent-any-color, and the double rent variants
// (multiplier 2). There is no separate pending-rent phase in No Mercy —
// the double-the-rent modifier does not exist, so demands are created
// immediately.
func (g *Game) PlayRent(playerID uuid.UUID, cardID Identifier, color Color) (*ActionDemandsCreated, error) {
	err := g.checkPlay(playerID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	allowed, multiplier, ok := rentCardSpec(card.AssetKey)
	if !ok {
		return nil, errors.InvalidCardForAction
	}

	if !slices.Contains(allowed, color) {
		return nil, InvalidColorForCard
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	properties := g.Properties[playerID]
	rent := properties.ColorRent(g.Config.ShackRentBonus, color) * multiplier

	demandsMap := make(map[Identifier]Demand)
	demandsSlice := make([]Demand, 0, len(g.Players)-1)
	for _, targetID := range g.Players {
		if targetID == playerID {
			continue
		}
		id := g.IDGenerator.New()
		demand := NewPaymentDemand(id, playerID, targetID, rent, DemandSourceRent)
		demandsMap[id] = demand
		demandsSlice = append(demandsSlice, demand)
	}
	g.Demands = demandsMap

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demandsSlice...)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}
