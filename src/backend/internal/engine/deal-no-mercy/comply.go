package deal_no_mercy

import (
	"math/rand/v2"
	"slices"

	"the-deal/internal/errors"

	"github.com/google/uuid"
)

func (g *Game) checkDemand(playerID uuid.UUID, demandID Identifier, kind DemandKind) (Demand, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return Demand{}, err
	}

	demand, ok := g.Demands[demandID]
	if !ok {
		return Demand{}, errors.DemandDoesNotExist
	}

	if playerID != demand.TargetID {
		return Demand{}, errors.PlayerNotDemandTarget
	}

	if demand.Kind != kind {
		return Demand{}, errors.DemandDoesNotExist
	}

	return demand, nil
}

// complyInactive dismisses a demand that was denied via NAH!: the original
// attacker accepts the denial and nothing is transferred.
func (g *Game) complyInactive(playerID uuid.UUID, demand Demand) *ActionDemandComplied {
	delete(g.Demands, demand.ID)
	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)
	g.SequenceNum++
	return action
}

// ComplyPaymentDemand pays a payment demand (rent, yoink). If the selected
// cards do not cover the amount they must be everything payable the player
// owns; the shortfall then costs the debtor one debt chip (if any remain —
// with none left the shortfall is written off).
func (g *Game) ComplyPaymentDemand(playerID uuid.UUID, demandID Identifier, cardIDs ...Identifier) (*ActionDemandComplied, error) {
	demand, err := g.checkDemand(playerID, demandID, DemandKindPayment)
	if err != nil {
		return nil, err
	}

	if !demand.IsActive {
		return g.complyInactive(playerID, demand), nil
	}

	cards := make(Cards, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		card, ok := g.Cards[cardID]
		if !ok {
			return nil, errors.CardDoesNotExist
		}
		cards = append(cards, card)
	}

	var paid int
	for _, card := range cards {
		paid += card.Value
	}

	if paid < demand.Payment.Amount {
		selectedIDs := make([]Identifier, 0, len(cards))
		for _, card := range cards {
			selectedIDs = append(selectedIDs, card.ID)
		}

		payableCards := g.getPayableCards(playerID)
		payableIDs := make([]Identifier, 0, len(payableCards))
		for _, card := range payableCards {
			payableIDs = append(payableIDs, card.ID)
		}

		slices.Sort(selectedIDs)
		slices.Sort(payableIDs)
		if !slices.Equal(selectedIDs, payableIDs) {
			return nil, errors.PaymentDoesNotCoverAmount
		}
	}

	transferredCards, transferredSets, err := g.transferCards(playerID, demand.SourceID, cardIDs...)
	if err != nil {
		return nil, err
	}

	// Short payment: surrender was total, now a debt chip goes to the
	// creditor. NAH! cannot prevent debt — only the original demand.
	var debt *DebtObligation
	if paid < demand.Payment.Amount && g.DebtChips[playerID] > 0 {
		g.DebtChips[playerID]--
		obligation := DebtObligation{
			ID:         g.IDGenerator.New(),
			DebtorID:   playerID,
			CreditorID: demand.SourceID,
		}
		g.Debts = append(g.Debts, obligation)
		debt = &obligation
	}

	sourceProperty := g.Properties[demand.SourceID]
	sourceMonies := g.Money[demand.SourceID]
	targetProperty := g.Properties[demand.TargetID]
	targetMonies := g.Money[demand.TargetID]

	delete(g.Demands, demandID)

	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)
	action.TransferCards = &TransferCards{
		SourceID:     playerID,
		TargetID:     demand.SourceID,
		Cards:        transferredCards,
		PropertySets: transferredSets,
		SourceSets:   sourceProperty.CompleteCount(),
		SourceMoney:  sourceMonies.Value(),
		TargetSets:   targetProperty.CompleteCount(),
		TargetMoney:  targetMonies.Value(),
		Debt:         debt,
	}

	g.SequenceNum++

	return action, nil
}

// ComplyPropertyDemand hands over the attacker-chosen property card
// (market crash). The choice was made at play time and lives in the demand.
func (g *Game) ComplyPropertyDemand(playerID uuid.UUID, demandID Identifier) (*ActionDemandComplied, error) {
	demand, err := g.checkDemand(playerID, demandID, DemandKindProperty)
	if err != nil {
		return nil, err
	}

	if !demand.IsActive {
		return g.complyInactive(playerID, demand), nil
	}

	card, set, err := g.transferProperty(demand.TargetID, demand.SourceID, demand.Property.TargetCardID)
	if err != nil {
		return nil, err
	}

	delete(g.Demands, demandID)

	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)
	action.TransferProperty = &TransferProperty{
		SourceID:    playerID,
		TargetID:    demand.SourceID,
		Card:        card,
		PropertySet: set,
	}

	g.SequenceNum++

	return action, nil
}

// ComplyPropertySetDemand hands over the complete set (set snatcher). The
// shack travels with the set.
func (g *Game) ComplyPropertySetDemand(playerID uuid.UUID, demandID Identifier) (*ActionDemandComplied, error) {
	demand, err := g.checkDemand(playerID, demandID, DemandKindPropertySet)
	if err != nil {
		return nil, err
	}

	if !demand.IsActive {
		return g.complyInactive(playerID, demand), nil
	}

	targetProperties := g.Properties[demand.TargetID]
	setIdx := targetProperties.IndexBySetID(demand.PropertySet.PropertySetID)
	if setIdx == -1 {
		return nil, errors.PropertySetDoesntExist
	}
	if !targetProperties[setIdx].IsComplete() {
		return nil, errors.PropertySetIsNotComplete
	}

	set, err := g.transferSet(demand.TargetID, demand.SourceID, setIdx)
	if err != nil {
		return nil, err
	}

	delete(g.Demands, demandID)

	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)
	action.TransferPropertySets = &TransferPropertySets{
		SourceID:     playerID,
		TargetID:     demand.SourceID,
		PropertySets: PropertySets{set},
	}

	g.SequenceNum++

	return action, nil
}

// ComplyColorPropertiesDemand hands over every set of the raided color
// (property raid), complete sets included, shacks travelling along.
func (g *Game) ComplyColorPropertiesDemand(playerID uuid.UUID, demandID Identifier) (*ActionDemandComplied, error) {
	demand, err := g.checkDemand(playerID, demandID, DemandKindColorProperties)
	if err != nil {
		return nil, err
	}

	if !demand.IsActive {
		return g.complyInactive(playerID, demand), nil
	}

	color := demand.ColorProperties.Color

	var transferred PropertySets
	for {
		targetProperties := g.Properties[demand.TargetID]
		setIdx := -1
		for i, set := range targetProperties {
			if set.Color == color {
				setIdx = i
				break
			}
		}
		if setIdx == -1 {
			break
		}

		set, err := g.transferSet(demand.TargetID, demand.SourceID, setIdx)
		if err != nil {
			return nil, err
		}
		transferred.Add(set)
	}

	delete(g.Demands, demandID)

	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)
	action.TransferPropertySets = &TransferPropertySets{
		SourceID:     playerID,
		TargetID:     demand.SourceID,
		PropertySets: transferred,
	}

	g.SequenceNum++

	return action, nil
}

// ComplyBankCardDemand hands over the attacker-chosen bank card (heist).
func (g *Game) ComplyBankCardDemand(playerID uuid.UUID, demandID Identifier) (*ActionDemandComplied, error) {
	demand, err := g.checkDemand(playerID, demandID, DemandKindBankCard)
	if err != nil {
		return nil, err
	}

	if !demand.IsActive {
		return g.complyInactive(playerID, demand), nil
	}

	card, err := g.transferMoney(demand.TargetID, demand.SourceID, demand.BankCard.TargetCardID)
	if err != nil {
		return nil, err
	}

	delete(g.Demands, demandID)

	targetMoney := g.Money[demand.TargetID]
	sourceMoney := g.Money[demand.SourceID]
	targetProperties := g.Properties[demand.TargetID]
	sourceProperties := g.Properties[demand.SourceID]

	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)
	action.TransferCards = &TransferCards{
		SourceID:    playerID,
		TargetID:    demand.SourceID,
		Cards:       Cards{card},
		SourceMoney: targetMoney.Value(),
		TargetMoney: sourceMoney.Value(),
		SourceSets:  targetProperties.CompleteCount(),
		TargetSets:  sourceProperties.CompleteCount(),
	}

	g.SequenceNum++

	return action, nil
}

// ComplyBankSwapDemand swaps the two banks wholesale.
func (g *Game) ComplyBankSwapDemand(playerID uuid.UUID, demandID Identifier) (*ActionDemandComplied, error) {
	demand, err := g.checkDemand(playerID, demandID, DemandKindBankSwap)
	if err != nil {
		return nil, err
	}

	if !demand.IsActive {
		return g.complyInactive(playerID, demand), nil
	}

	g.Money[demand.SourceID], g.Money[demand.TargetID] = g.Money[demand.TargetID], g.Money[demand.SourceID]

	delete(g.Demands, demandID)

	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)
	action.TransferBankSwap = &TransferBankSwap{
		SourceID:   playerID,
		TargetID:   demand.SourceID,
		SourceBank: g.Money[demand.TargetID],
		TargetBank: g.Money[demand.SourceID],
	}

	g.SequenceNum++

	return action, nil
}

// ComplyDebtTrapDemand moves a debt chip from the target's pool into an
// outstanding obligation towards the attacker. If the target ran out of
// chips in the meantime the trap fizzles (nothing to take).
func (g *Game) ComplyDebtTrapDemand(playerID uuid.UUID, demandID Identifier) (*ActionDemandComplied, error) {
	demand, err := g.checkDemand(playerID, demandID, DemandKindDebtTrap)
	if err != nil {
		return nil, err
	}

	if !demand.IsActive {
		return g.complyInactive(playerID, demand), nil
	}

	delete(g.Demands, demandID)

	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)

	if g.DebtChips[demand.TargetID] > 0 {
		g.DebtChips[demand.TargetID]--
		obligation := DebtObligation{
			ID:         g.IDGenerator.New(),
			DebtorID:   demand.TargetID,
			CreditorID: demand.SourceID,
		}
		g.Debts = append(g.Debts, obligation)
		action.TransferDebtChip = &TransferDebtChip{Debt: obligation}
	}

	g.SequenceNum++

	return action, nil
}

// ComplyRepoManDemand: the target keeps keepCardID (one of their property
// cards) and distributes every other property card per distribution
// (card id → recipient). Recipients may be any player except the target.
func (g *Game) ComplyRepoManDemand(playerID uuid.UUID, demandID Identifier, keepCardID Identifier, distribution map[Identifier]uuid.UUID) (*ActionDemandComplied, error) {
	demand, err := g.checkDemand(playerID, demandID, DemandKindRepoMan)
	if err != nil {
		return nil, err
	}

	if !demand.IsActive {
		return g.complyInactive(playerID, demand), nil
	}

	targetProperties := g.Properties[demand.TargetID]
	owned := make([]Identifier, 0)
	for _, set := range targetProperties {
		for _, c := range set.Cards {
			owned = append(owned, c.ID)
		}
	}

	if !slices.Contains(owned, keepCardID) {
		return nil, errors.PlayerDoesNotHaveCard
	}

	keepCard := g.Cards[keepCardID]

	if len(distribution) != len(owned)-1 {
		return nil, InvalidDistribution
	}

	for _, cardID := range owned {
		if cardID == keepCardID {
			continue
		}
		recipientID, ok := distribution[cardID]
		if !ok {
			return nil, InvalidDistribution
		}
		if recipientID == demand.TargetID {
			return nil, InvalidDistribution
		}
		if err := g.checkPlayer(recipientID); err != nil {
			return nil, err
		}
	}

	entries := make([]DistributionEntry, 0, len(distribution))
	for _, cardID := range owned {
		if cardID == keepCardID {
			continue
		}

		recipientID := distribution[cardID]
		card, set, err := g.transferProperty(demand.TargetID, recipientID, cardID)
		if err != nil {
			return nil, err
		}
		setCopy := set
		entries = append(entries, DistributionEntry{
			RecipientID: recipientID,
			Card:        card,
			PropertySet: &setCopy,
		})
	}

	delete(g.Demands, demandID)

	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)
	action.TransferDistribution = &TransferDistribution{
		SourceID: playerID,
		KeptCard: keepCard,
		Entries:  entries,
	}

	g.SequenceNum++

	return action, nil
}

// ComplyTaxDayDemand: the target keeps keepCardID (one of their bank cards)
// and distributes every other bank card per distribution (card id →
// recipient's bank). Recipients may be any player except the target.
func (g *Game) ComplyTaxDayDemand(playerID uuid.UUID, demandID Identifier, keepCardID Identifier, distribution map[Identifier]uuid.UUID) (*ActionDemandComplied, error) {
	demand, err := g.checkDemand(playerID, demandID, DemandKindTaxDay)
	if err != nil {
		return nil, err
	}

	if !demand.IsActive {
		return g.complyInactive(playerID, demand), nil
	}

	bank := g.Money[demand.TargetID]
	owned := make([]Identifier, 0, bank.Len())
	for _, c := range bank {
		owned = append(owned, c.ID)
	}

	if !slices.Contains(owned, keepCardID) {
		return nil, errors.PlayerDoesNotHaveCard
	}

	keepCard := g.Cards[keepCardID]

	if len(distribution) != len(owned)-1 {
		return nil, InvalidDistribution
	}

	for _, cardID := range owned {
		if cardID == keepCardID {
			continue
		}
		recipientID, ok := distribution[cardID]
		if !ok {
			return nil, InvalidDistribution
		}
		if recipientID == demand.TargetID {
			return nil, InvalidDistribution
		}
		if err := g.checkPlayer(recipientID); err != nil {
			return nil, err
		}
	}

	entries := make([]DistributionEntry, 0, len(distribution))
	for _, cardID := range owned {
		if cardID == keepCardID {
			continue
		}

		recipientID := distribution[cardID]
		card, err := g.transferMoney(demand.TargetID, recipientID, cardID)
		if err != nil {
			return nil, err
		}
		entries = append(entries, DistributionEntry{
			RecipientID: recipientID,
			Card:        card,
		})
	}

	delete(g.Demands, demandID)

	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)
	action.TransferDistribution = &TransferDistribution{
		SourceID: playerID,
		KeptCard: keepCard,
		Entries:  entries,
	}

	g.SequenceNum++

	return action, nil
}

// ComplyPickpocketDemand moves every hand card of the named category from
// the target's hand into the attacker's hand.
func (g *Game) ComplyPickpocketDemand(playerID uuid.UUID, demandID Identifier) (*ActionDemandComplied, error) {
	demand, err := g.checkDemand(playerID, demandID, DemandKindPickpocket)
	if err != nil {
		return nil, err
	}

	if !demand.IsActive {
		return g.complyInactive(playerID, demand), nil
	}

	category := demand.Pickpocket.Category

	targetHand := g.Hands[demand.TargetID]
	var stolen Cards
	var remaining Cards
	for _, card := range targetHand {
		if category.Matches(card) {
			stolen.Add(card)
			continue
		}
		remaining.Add(card)
	}
	g.Hands[demand.TargetID] = remaining

	sourceHand := g.Hands[demand.SourceID]
	sourceHand.Add(stolen...)
	g.Hands[demand.SourceID] = sourceHand

	delete(g.Demands, demandID)

	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID)
	action.TransferPickpocket = &TransferPickpocket{
		SourceID: playerID,
		TargetID: demand.SourceID,
		Category: category,
		Cards:    stolen,
	}

	g.SequenceNum++

	return action, nil
}

// DenyDemand plays NAH! against a demand targeting the player. The demand
// is re-issued under a fresh ID with source/target swapped and IsActive
// flipped, so the original attacker may accept (comply) or counter-deny.
func (g *Game) DenyDemand(playerID uuid.UUID, demandID Identifier, cardID Identifier) (*ActionDemandsCreated, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	isCurrentPlayer := g.Players[g.CurrPlayerIdx] == playerID
	if isCurrentPlayer && g.Config.NahConsumesMove {
		err = g.checkMoves()
		if err != nil {
			return nil, err
		}
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	if card.AssetKey != AssetKeyNah {
		return nil, errors.InvalidCardForAction
	}

	demand, ok := g.Demands[demandID]
	if !ok {
		return nil, errors.DemandDoesNotExist
	}

	if playerID != demand.TargetID {
		return nil, errors.PlayerNotDemandTarget
	}

	deniedDemand := demand

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	id := g.IDGenerator.New()
	demand.Deny(id)
	delete(g.Demands, demandID)
	g.Demands[id] = demand

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, &deniedDemand, demand)

	g.LastAction = card
	if isCurrentPlayer && g.Config.NahConsumesMove {
		g.CompleteMove()
	}
	g.SequenceNum++

	return action, nil
}

// SettleDebt settles one outstanding debt chip at the start of the debtor's
// turn: the debtor gives the creditor any one hand card, which the creditor
// receives as if they had played it themselves (property → table, anything
// else → bank). Consumes one of the turn's plays and returns the chip to
// the debtor's pool. Debt settlement is mandatory before any other play
// and cannot be denied by NAH!.
func (g *Game) SettleDebt(playerID uuid.UUID, debtID Identifier, cardID Identifier) (*ActionDebtSettled, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkDemands()
	if err != nil {
		return nil, err
	}

	err = g.checkMoves()
	if err != nil {
		return nil, err
	}

	debtIdx := -1
	for i, d := range g.Debts {
		if d.ID == debtID && d.DebtorID == playerID {
			debtIdx = i
			break
		}
	}
	if debtIdx == -1 {
		return nil, DebtDoesNotExist
	}
	debt := g.Debts[debtIdx]

	card, err := g.removeFromHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	var setPtr *PropertySet
	switch card.Category {
	case CategoryPureProperty, CategoryWildProperty:
		set := g.placeProperty(debt.CreditorID, card)
		setPtr = &set
	default:
		money := g.Money[debt.CreditorID]
		money.Add(card)
		g.Money[debt.CreditorID] = money
	}

	g.Debts = slices.Delete(g.Debts, debtIdx, debtIdx+1)
	g.DebtChips[playerID]++

	action := NewActionDebtSettled(g.SequenceNum, playerID, debt, card, setPtr, false)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// DefaultMove is the timeout fallback for the current player: settle any
// outstanding debts with arbitrary hand cards, merge invalid duplicate
// sets, then discard down to the hand limit.
func (g *Game) DefaultMove(playerID uuid.UUID) ([]Action, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	actions := make([]Action, 0)

	for _, debt := range g.PlayerDebts(playerID) {
		hand := g.Hands[playerID]
		if hand.Len() == 0 || g.MovesLeft <= 0 {
			break
		}

		action, err := g.SettleDebt(playerID, debt.ID, hand[0].ID)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	actions = append(actions, g.defaultRearrangeProperties(playerID)...)

	cards := Cards{}
	for {
		hand := g.Hands[playerID]
		n := hand.Len()
		if n <= g.Config.MaxHandSize {
			break
		}

		r := rand.IntN(n)
		card := hand[r]
		card, err = g.discardHand(playerID, card.ID)
		if err != nil {
			return nil, err
		}

		cards = append(cards, card)
	}

	action := NewActionDiscardCards(g.SequenceNum, playerID, cards...)
	actions = append(actions, action)

	g.SequenceNum++

	return actions, nil
}

func (g *Game) defaultRearrangeProperties(playerID uuid.UUID) []Action {
	properties := g.Properties[playerID]
	if properties.Valid() {
		return nil
	}

	actions := make([]Action, 0)

	targetIncomplete := make(map[Color]int)

	for i := 0; i < len(properties); i++ {
		set := properties[i]
		if set.IsComplete() {
			continue
		}

		targetIdx, ok := targetIncomplete[set.Color]
		if !ok {
			targetIncomplete[set.Color] = i
			continue
		}

		targetSet := properties[targetIdx]
		for _, card := range set.Cards {
			card.ActiveColor = targetSet.Color
			targetSet.Cards.Add(card)
			actions = append(actions, NewActionRearrangeCard(g.SequenceNum, playerID, &card, targetSet, properties.CompleteCount()))
			g.SequenceNum++
		}

		properties[targetIdx] = targetSet
		properties[i].Cards = Cards{}
	}

	g.Properties[playerID] = properties
	g.cleanProperties(playerID)

	return actions
}

type paymentCardChoice struct {
	card          Card
	isCompleteSet bool
}

type paymentSubsetChoice struct {
	cardIDs []Identifier
	paid    int
}

func betterPaymentSubset(candidate, current paymentSubsetChoice, amount int) bool {
	if len(current.cardIDs) == 0 {
		return true
	}

	candidateOverpay := candidate.paid - amount
	currentOverpay := current.paid - amount
	if candidateOverpay != currentOverpay {
		return candidateOverpay < currentOverpay
	}

	if len(candidate.cardIDs) != len(current.cardIDs) {
		return len(candidate.cardIDs) < len(current.cardIDs)
	}

	c := append([]Identifier(nil), candidate.cardIDs...)
	b := append([]Identifier(nil), current.cardIDs...)
	slices.Sort(c)
	slices.Sort(b)
	for i := range c {
		if c[i] == b[i] {
			continue
		}
		return c[i] < b[i]
	}

	return false
}

func chooseBestPaymentSubset(cards Cards, amount int) paymentSubsetChoice {
	best := paymentSubsetChoice{}
	currentIDs := make([]Identifier, 0, len(cards))

	var dfs func(idx, total int)
	dfs = func(idx, total int) {
		if total >= amount {
			candidate := paymentSubsetChoice{
				cardIDs: append([]Identifier(nil), currentIDs...),
				paid:    total,
			}
			if betterPaymentSubset(candidate, best, amount) {
				best = candidate
			}
			return
		}

		if idx >= len(cards) {
			return
		}

		currentIDs = append(currentIDs, cards[idx].ID)
		dfs(idx+1, total+cards[idx].Value)
		currentIDs = currentIDs[:len(currentIDs)-1]

		dfs(idx+1, total)
	}

	dfs(0, 0)
	return best
}

func (g *Game) defaultPaymentCardIDs(playerID uuid.UUID, amount int) []Identifier {
	money := append(Cards(nil), g.Money[playerID]...)
	moneyTotal := money.Value()

	properties := make([]paymentCardChoice, 0)
	propertyTotal := 0
	for _, set := range g.Properties[playerID] {
		for _, card := range set.Cards {
			properties = append(properties, paymentCardChoice{
				card:          card,
				isCompleteSet: set.IsComplete(),
			})
			propertyTotal += card.Value
		}
	}

	totalPayable := moneyTotal + propertyTotal
	if totalPayable < amount {
		ids := make([]Identifier, 0, len(money)+len(properties))
		for _, card := range money {
			ids = append(ids, card.ID)
		}
		for _, prop := range properties {
			ids = append(ids, prop.card.ID)
		}
		return ids
	}

	if moneyTotal >= amount {
		best := chooseBestPaymentSubset(money, amount)
		return best.cardIDs
	}

	selected := make([]Identifier, 0, len(money)+len(properties))
	for _, card := range money {
		selected = append(selected, card.ID)
	}

	remaining := amount - moneyTotal
	slices.SortFunc(properties, func(a, b paymentCardChoice) int {
		if a.isCompleteSet != b.isCompleteSet {
			if !a.isCompleteSet {
				return -1
			}
			return 1
		}
		if a.card.Value != b.card.Value {
			return a.card.Value - b.card.Value
		}
		if a.card.ID < b.card.ID {
			return -1
		}
		if a.card.ID > b.card.ID {
			return 1
		}
		return 0
	})

	paidByProperty := 0
	for _, prop := range properties {
		selected = append(selected, prop.card.ID)
		paidByProperty += prop.card.Value
		if paidByProperty >= remaining {
			break
		}
	}

	return selected
}

// defaultHighestValueCard picks the card with the highest value (ties by
// lowest id for determinism).
func defaultHighestValueCard(cards Cards) (Card, bool) {
	if len(cards) == 0 {
		return Card{}, false
	}

	best := cards[0]
	for _, card := range cards[1:] {
		if card.Value > best.Value || (card.Value == best.Value && card.ID < best.ID) {
			best = card
		}
	}
	return best, true
}

// DefaultDemand is the timeout fallback for a demand target: comply with a
// reasonable default choice.
func (g *Game) DefaultDemand(playerID uuid.UUID, demandID Identifier) (Action, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	demand, ok := g.Demands[demandID]
	if !ok {
		return nil, errors.DemandDoesNotExist
	}

	switch demand.Kind {
	case DemandKindPayment:
		cardIDs := g.defaultPaymentCardIDs(playerID, demand.Payment.Amount)
		return g.ComplyPaymentDemand(playerID, demandID, cardIDs...)
	case DemandKindProperty:
		return g.ComplyPropertyDemand(playerID, demandID)
	case DemandKindPropertySet:
		return g.ComplyPropertySetDemand(playerID, demandID)
	case DemandKindColorProperties:
		return g.ComplyColorPropertiesDemand(playerID, demandID)
	case DemandKindBankCard:
		return g.ComplyBankCardDemand(playerID, demandID)
	case DemandKindBankSwap:
		return g.ComplyBankSwapDemand(playerID, demandID)
	case DemandKindDebtTrap:
		return g.ComplyDebtTrapDemand(playerID, demandID)
	case DemandKindRepoMan:
		if !demand.IsActive {
			return g.ComplyRepoManDemand(playerID, demandID, NullIdentifier, nil)
		}

		properties := g.Properties[playerID]
		var owned Cards
		for _, set := range properties {
			owned = append(owned, set.Cards...)
		}
		keep, ok := defaultHighestValueCard(owned)
		if !ok {
			return nil, errors.PlayerDoesNotHaveCard
		}

		distribution := make(map[Identifier]uuid.UUID, len(owned)-1)
		for _, card := range owned {
			if card.ID == keep.ID {
				continue
			}
			distribution[card.ID] = demand.SourceID
		}
		return g.ComplyRepoManDemand(playerID, demandID, keep.ID, distribution)
	case DemandKindTaxDay:
		if !demand.IsActive {
			return g.ComplyTaxDayDemand(playerID, demandID, NullIdentifier, nil)
		}

		bank := g.Money[playerID]
		keep, ok := defaultHighestValueCard(bank)
		if !ok {
			return nil, errors.PlayerDoesNotHaveCard
		}

		distribution := make(map[Identifier]uuid.UUID, bank.Len()-1)
		for _, card := range bank {
			if card.ID == keep.ID {
				continue
			}
			distribution[card.ID] = demand.SourceID
		}
		return g.ComplyTaxDayDemand(playerID, demandID, keep.ID, distribution)
	case DemandKindPickpocket:
		return g.ComplyPickpocketDemand(playerID, demandID)
	default:
		return nil, errors.DemandDoesNotExist
	}
}
