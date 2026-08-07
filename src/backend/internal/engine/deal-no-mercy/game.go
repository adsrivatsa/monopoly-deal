package deal_no_mercy

import (
	"slices"

	"the-deal/internal/errors"

	"github.com/google/uuid"
)

type Game struct {
	Config        Settings                   `json:"config" msgpack:"a"`
	IDGenerator   *IdentifierGenerator       `json:"id_generator" msgpack:"b"`
	Deck          Deck                       `json:"deck" msgpack:"d"`
	Cards         map[Identifier]Card        `json:"cards" msgpack:"e"`
	Players       []uuid.UUID                `json:"players" msgpack:"f"`
	CurrPlayerIdx int                        `json:"curr_player_idx" msgpack:"g"`
	MovesLeft     int                        `json:"moves_left" msgpack:"h"`
	Hands         map[uuid.UUID]Cards        `json:"hands" msgpack:"i"`
	Money         map[uuid.UUID]Cards        `json:"money" msgpack:"j"`
	Properties    map[uuid.UUID]PropertySets `json:"properties" msgpack:"k"`
	Demands       map[Identifier]Demand      `json:"demands" msgpack:"l"`
	DebtChips     map[uuid.UUID]int          `json:"debt_chips" msgpack:"p"`
	Debts         []DebtObligation           `json:"debts" msgpack:"q"`
	GoAgainQueued bool                       `json:"go_again_queued" msgpack:"r"`
	LastAction    Card                       `json:"last_action" msgpack:"n"`
	SequenceNum   int                        `json:"sequence_num" msgpack:"o"`
}

func NewGame(cfg Settings, playerIDs []uuid.UUID) *Game {
	ig := NewIdentifierGenerator()

	d, cardMap := NewDeck(cfg, ig)

	hands := make(map[uuid.UUID]Cards)
	for _, playerID := range playerIDs {
		hand := d.Draw(cfg.StartNumCards)
		hands[playerID] = hand
	}

	money := make(map[uuid.UUID]Cards)
	for _, playerID := range playerIDs {
		money[playerID] = Cards{}
	}

	properties := make(map[uuid.UUID]PropertySets)
	for _, playerID := range playerIDs {
		properties[playerID] = PropertySets{}
	}

	debtChips := make(map[uuid.UUID]int)
	for _, playerID := range playerIDs {
		debtChips[playerID] = cfg.DebtChipsPerPlayer
	}

	g := &Game{
		IDGenerator:   ig,
		Deck:          d,
		Cards:         cardMap,
		Players:       playerIDs,
		CurrPlayerIdx: 0,
		MovesLeft:     cfg.MovesPerTurn,
		Hands:         hands,
		Money:         money,
		Properties:    properties,
		Demands:       make(map[Identifier]Demand),
		DebtChips:     debtChips,
		Debts:         []DebtObligation{},
		Config:        cfg,
	}

	return g
}

func (g *Game) CountMoney(playerID uuid.UUID) (int, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return 0, err
	}

	money := g.Money[playerID]
	return money.Value(), nil
}

func (g *Game) CountCompletedSets(playerID uuid.UUID) (int, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return 0, err
	}

	properties := g.Properties[playerID]
	return properties.CompleteCount(), nil
}

func (g *Game) CountHands(playerID uuid.UUID) (int, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return 0, err
	}

	hand := g.Hands[playerID]
	return hand.Len(), nil
}

// CountAvailableDebtChips returns how many of the player's debt chips are
// not currently tied up in outstanding obligations.
func (g *Game) CountAvailableDebtChips(playerID uuid.UUID) (int, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return 0, err
	}

	return g.DebtChips[playerID], nil
}

// PlayerDebts returns the outstanding obligations owed BY the player.
func (g *Game) PlayerDebts(playerID uuid.UUID) []DebtObligation {
	debts := make([]DebtObligation, 0)
	for _, d := range g.Debts {
		if d.DebtorID == playerID {
			debts = append(debts, d)
		}
	}
	return debts
}

func (g *Game) checkPlayer(playerID uuid.UUID) error {
	for _, player := range g.Players {
		if playerID == player {
			return nil
		}
	}
	return errors.PlayerNotInGame
}

func (g *Game) checkTurn(playerID uuid.UUID) error {
	playerIdx := slices.Index(g.Players, playerID)
	if playerIdx == -1 {
		return errors.PlayerNotInGame
	}

	if playerIdx != g.CurrPlayerIdx {
		return errors.NotPlayersTurn
	}

	return nil
}

func (g *Game) checkCard(cardID Identifier, categories ...Category) (Card, error) {
	card, ok := g.Cards[cardID]
	if !ok {
		return card, errors.CardDoesNotExist
	}

	valid := false
	for _, category := range categories {
		valid = valid || (category == card.Category)
	}
	if !valid {
		return card, errors.InvalidCardForAction
	}

	return card, nil
}

func (g *Game) checkMoves() error {
	if g.MovesLeft <= 0 {
		return errors.NoMovesLeft
	}
	return nil
}

func (g *Game) checkDemands() error {
	if len(g.Demands) != 0 {
		return errors.ActiveDemandExists
	}
	return nil
}

// checkDebts blocks all regular plays while the player still has
// outstanding debt chips to settle — settlements come first.
func (g *Game) checkDebts(playerID uuid.UUID) error {
	for _, d := range g.Debts {
		if d.DebtorID == playerID {
			return OutstandingDebtExists
		}
	}
	return nil
}

// checkPlay is the common preamble shared by every play of the current
// player's turn.
func (g *Game) checkPlay(playerID uuid.UUID) error {
	err := g.checkPlayer(playerID)
	if err != nil {
		return err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return err
	}

	err = g.checkMoves()
	if err != nil {
		return err
	}

	err = g.checkDemands()
	if err != nil {
		return err
	}

	return g.checkDebts(playerID)
}

// checkActionPlay verifies the play preamble plus that cardID is the given
// action card.
func (g *Game) checkActionPlay(playerID uuid.UUID, cardID Identifier, key AssetKey) (Card, error) {
	err := g.checkPlay(playerID)
	if err != nil {
		return Card{}, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return Card{}, err
	}

	if card.AssetKey != key {
		return Card{}, errors.InvalidCardForAction
	}

	return card, nil
}

// checkTargetedActionPlay is checkActionPlay plus target validation.
func (g *Game) checkTargetedActionPlay(playerID, targetID uuid.UUID, cardID Identifier, key AssetKey) (Card, error) {
	if playerID == targetID {
		return Card{}, errors.CannotStealFromSelf
	}

	err := g.checkPlayer(targetID)
	if err != nil {
		return Card{}, err
	}

	return g.checkActionPlay(playerID, cardID, key)
}

// removeFromHand takes a card out of the player's hand without discarding
// it. Use for plays that keep the card on the board (money, properties,
// shacks) — discardHand would also return the card to the deck, leaving the
// same card id in two places once the deck cycles back around.
func (g *Game) removeFromHand(playerID uuid.UUID, cardID Identifier) (Card, error) {
	hand := g.Hands[playerID]
	card, ok := hand.RemoveByID(cardID)
	if !ok {
		return Card{}, errors.PlayerDoesNotHaveCard
	}
	g.Hands[playerID] = hand

	return card, nil
}

func (g *Game) discardHand(playerID uuid.UUID, cardID Identifier) (Card, error) {
	card, err := g.removeFromHand(playerID, cardID)
	if err != nil {
		return Card{}, err
	}

	g.Deck.Add(card)

	return card, nil
}

func (g *Game) removeMoney(playerID uuid.UUID, cardID Identifier) (Card, error) {
	money := g.Money[playerID]
	card, ok := money.RemoveByID(cardID)
	if !ok {
		return card, errors.PlayerDoesNotHaveCard
	}
	g.Money[playerID] = money

	return card, nil
}

// cleanProperties drops emptied sets and discards their orphaned shacks
// back to the deck (a shack cannot exist without a set).
func (g *Game) cleanProperties(playerID uuid.UUID) {
	properties := g.Properties[playerID]
	orphanedShacks := properties.Clean()
	for _, shack := range orphanedShacks {
		g.Deck.Add(shack)
	}
	g.Properties[playerID] = properties
}

// removeProperty removes a property card from whichever of the player's
// sets holds it. No Mercy has no house/hotel locks and complete sets are
// stealable, so any card can come out; callers gate stealability.
func (g *Game) removeProperty(playerID uuid.UUID, cardID Identifier) (Card, error) {
	properties := g.Properties[playerID]
	i, j := properties.IndexByCardID(cardID)
	if i == -1 || j == -1 {
		return Card{}, errors.PlayerDoesNotHaveCard
	}

	set := properties[i]
	card, _ := set.Cards.RemoveByIdx(j)
	properties[i] = set
	g.Properties[playerID] = properties

	g.cleanProperties(playerID)

	return card, nil
}

func (g *Game) drawCards(playerID uuid.UUID, amount int) Cards {
	cards := g.Deck.Draw(amount)
	hand := g.Hands[playerID]
	hand.Add(cards...)
	g.Hands[playerID] = hand
	return cards
}

// getPayableCards returns everything on the player's table that can be used
// for a payment: bank cards and property cards. Shacks are immovable and
// never payable.
func (g *Game) getPayableCards(playerID uuid.UUID) Cards {
	payable := make(Cards, 0)
	payable.Add(g.Money[playerID]...)

	for _, set := range g.Properties[playerID] {
		payable.Add(set.Cards...)
	}

	return payable
}

func (g *Game) transferMoney(sourceID, targetID uuid.UUID, cardID Identifier) (Card, error) {
	card, err := g.removeMoney(sourceID, cardID)
	if err != nil {
		return Card{}, err
	}

	money := g.Money[targetID]
	money.Add(card)
	g.Money[targetID] = money

	return card, nil
}

// placeProperty puts a received property card into a brand-new set on the
// target's table (receivers merge sets themselves via RearrangeProperty).
func (g *Game) placeProperty(targetID uuid.UUID, card Card) PropertySet {
	color := card.ActiveColor
	if color == ColorUnspecified && len(card.Colors) > 0 {
		color = card.Colors[0]
	}

	setID := g.IDGenerator.New()
	set := NewPropertySet(setID, color)
	card.ActiveColor = color
	set.Cards.Add(card)

	properties := g.Properties[targetID]
	properties.Add(set)
	g.Properties[targetID] = properties

	return set
}

func (g *Game) transferProperty(sourceID, targetID uuid.UUID, cardID Identifier) (Card, PropertySet, error) {
	card, err := g.removeProperty(sourceID, cardID)
	if err != nil {
		return Card{}, PropertySet{}, err
	}

	set := g.placeProperty(targetID, card)
	return card, set, nil
}

// transferSet moves a whole property set — shack included — from source to
// target.
func (g *Game) transferSet(sourceID, targetID uuid.UUID, setIdx int) (PropertySet, error) {
	sourceProperties := g.Properties[sourceID]
	set, ok := sourceProperties.RemoveByIdx(setIdx)
	if !ok {
		return PropertySet{}, errors.PropertySetDoesntExist
	}
	g.Properties[sourceID] = sourceProperties

	targetProperties := g.Properties[targetID]
	targetProperties.Add(set)
	g.Properties[targetID] = targetProperties

	return set, nil
}

func removePropertyCardFromClone(properties *PropertySets, cardID Identifier) bool {
	i, j := properties.IndexByCardID(cardID)
	if i == -1 || j == -1 {
		return false
	}

	set := (*properties)[i]
	cards := set.Cards
	set.Cards = append(append(Cards(nil), cards[:j]...), cards[j+1:]...)
	(*properties)[i] = set

	return true
}

func (g *Game) canTransferCards(sourceID uuid.UUID, cardIDs ...Identifier) error {
	seen := make(map[Identifier]struct{}, len(cardIDs))
	for _, cardID := range cardIDs {
		if _, ok := g.Cards[cardID]; !ok {
			return errors.CardDoesNotExist
		}
		if _, ok := seen[cardID]; ok {
			return errors.DuplicateCardPaymentExists
		}
		seen[cardID] = struct{}{}
	}

	sourceMoney := slices.Clone(g.Money[sourceID])
	sourceProperties := slices.Clone(g.Properties[sourceID])

	for _, cardID := range cardIDs {
		if _, ok := sourceMoney.RemoveByID(cardID); ok {
			continue
		}
		if removePropertyCardFromClone(&sourceProperties, cardID) {
			continue
		}
		return errors.InvalidCardForAction
	}

	return nil
}

// transferCards moves a payment from source to target: bank/action cards
// land in the target's bank, property cards each form a new set.
func (g *Game) transferCards(sourceID, targetID uuid.UUID, cardIDs ...Identifier) (Cards, PropertySets, error) {
	err := g.canTransferCards(sourceID, cardIDs...)
	if err != nil {
		return nil, nil, err
	}

	var transferredCards Cards
	var transferredSets PropertySets

	for _, cardID := range cardIDs {
		card, err := g.transferMoney(sourceID, targetID, cardID)
		if err == nil {
			transferredCards.Add(card)
			continue
		}

		_, set, err := g.transferProperty(sourceID, targetID, cardID)
		if err != nil {
			return nil, nil, err
		}
		transferredSets.Add(set)
	}

	return transferredCards, transferredSets, nil
}

func (g *Game) CompleteMove() {
	g.MovesLeft--
}

func (g *Game) PlayMoney(playerID uuid.UUID, cardID Identifier) (*ActionPlayMoney, error) {
	err := g.checkPlay(playerID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryMoney, CategoryAction)
	if err != nil {
		return nil, err
	}

	_, err = g.removeFromHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	money := g.Money[playerID]
	money.Add(card)
	g.Money[playerID] = money

	action := NewActionPlayMoney(g.SequenceNum, playerID, card)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayProperty(playerID uuid.UUID, cardID Identifier, propSetIDPtr *Identifier, activeColorPtr *Color) (*ActionPlayProperty, error) {
	err := g.checkPlay(playerID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryPureProperty, CategoryWildProperty)
	if err != nil {
		return nil, err
	}

	var setIdx int
	var propSetID Identifier
	var propSet PropertySet
	resolvedColor := card.ActiveColor
	if propSetIDPtr != nil {
		propSetID = *propSetIDPtr

		properties := g.Properties[playerID]

		setIdx = properties.IndexBySetID(propSetID)
		if setIdx == -1 {
			return nil, errors.PropertySetDoesntExist
		}

		propSet = properties[setIdx]

		if propSet.IsComplete() {
			return nil, errors.PropertySetIsComplete
		}

		if !card.HasColor(propSet.Color) {
			return nil, errors.CardCannotBeAssignedToSet
		}

		resolvedColor = propSet.Color
	} else {
		if activeColorPtr != nil {
			if !card.HasColor(*activeColorPtr) {
				return nil, errors.CardCannotBeAssignedToSet
			}
			resolvedColor = *activeColorPtr
		}

		propSetID = g.IDGenerator.New()
		propSet = NewPropertySet(propSetID, resolvedColor)
	}

	_, err = g.removeFromHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	card.ActiveColor = resolvedColor
	propSet.Cards.Add(card)

	properties := g.Properties[playerID]
	if propSetIDPtr != nil {
		properties[setIdx] = propSet
	} else {
		properties.Add(propSet)
	}
	g.Properties[playerID] = properties

	action := NewActionPlayProperty(g.SequenceNum, playerID, propSet)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayShack attaches a SHACK to any of the player's sets — complete or not,
// any color including railroads and utilities — max one per set. The card
// stays on the table (it is NOT returned to the deck).
func (g *Game) PlayShack(playerID uuid.UUID, cardID, propSetID Identifier) (*ActionPlayShack, error) {
	card, err := g.checkActionPlay(playerID, cardID, AssetKeyShack)
	if err != nil {
		return nil, err
	}

	properties := g.Properties[playerID]
	setIdx := properties.IndexBySetID(propSetID)
	if setIdx == -1 {
		return nil, errors.PropertySetDoesntExist
	}

	propSet := properties[setIdx]
	if propSet.HasShack() {
		return nil, SetAlreadyHasShack
	}

	_, err = g.removeFromHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	propSet.Shack = &card
	properties[setIdx] = propSet
	g.Properties[playerID] = properties

	action := NewActionPlayShack(g.SequenceNum, playerID, card, propSet)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayBigPayday draws until the player's hand holds
// Config.BigPaydayHandTarget cards (nothing if already at or above it).
func (g *Game) PlayBigPayday(playerID uuid.UUID, cardID Identifier) (*ActionPlayBigPayday, error) {
	card, err := g.checkActionPlay(playerID, cardID, AssetKeyBigPayday)
	if err != nil {
		return nil, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	toDraw := g.Config.BigPaydayHandTarget - len(g.Hands[playerID])
	var cards Cards
	if toDraw > 0 {
		cards = g.drawCards(playerID, toDraw)
	}

	action := NewActionPlayBigPayday(g.SequenceNum, playerID, card, cards)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

// PlayGoAgain must be the player's last play of the turn: it consumes all
// remaining plays, and after the turn ends the same player takes another
// full turn (draw included).
func (g *Game) PlayGoAgain(playerID uuid.UUID, cardID Identifier) (*ActionPlayGoAgain, error) {
	card, err := g.checkActionPlay(playerID, cardID, AssetKeyGoAgain)
	if err != nil {
		return nil, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	g.GoAgainQueued = true
	g.MovesLeft = 0

	action := NewActionPlayGoAgain(g.SequenceNum, playerID, card)

	g.LastAction = card
	g.SequenceNum++

	return action, nil
}

func (g *Game) DiscardCards(playerID uuid.UUID, cardIDs ...Identifier) (*ActionDiscardCards, error) {
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

	err = g.checkDebts(playerID)
	if err != nil {
		return nil, err
	}

	if len(cardIDs) == 0 {
		return nil, errors.InvalidAmountOfCards
	}

	hand := g.Hands[playerID]
	inHand := make(map[Identifier]struct{}, len(hand))
	for _, card := range hand {
		inHand[card.ID] = struct{}{}
	}

	for _, cardID := range cardIDs {
		if _, ok := inHand[cardID]; !ok {
			return nil, errors.PlayerDoesNotHaveCard
		}
	}

	cards := make(Cards, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		card, err := g.discardHand(playerID, cardID)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	action := NewActionDiscardCards(g.SequenceNum, playerID, cards...)

	g.SequenceNum++

	return action, nil
}

func (g *Game) CompleteTurn(playerID uuid.UUID) (*ActionStartTurn, error) {
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

	// Outstanding debts must have been settled during this turn. The only
	// escape is a genuinely empty hand (deck exhausted): then the remaining
	// debt is forgiven and the chips return to the debtor.
	if len(g.PlayerDebts(playerID)) > 0 {
		if len(g.Hands[playerID]) > 0 {
			return nil, OutstandingDebtExists
		}
		g.forgiveDebts(playerID)
	}

	hand := g.Hands[playerID]
	if hand.Len() > g.Config.MaxHandSize {
		return nil, errors.PlayerHandHasTooManyCards
	}

	properties := g.Properties[playerID]
	if !properties.Valid() {
		return nil, errors.InvalidPropertySets
	}

	goAgain := g.GoAgainQueued
	if goAgain {
		g.GoAgainQueued = false
	} else {
		n := len(g.Players)
		g.CurrPlayerIdx = (g.CurrPlayerIdx + 1) % n
	}
	g.MovesLeft = g.Config.MovesPerTurn

	nextPlayerID := g.Players[g.CurrPlayerIdx]
	action := g.startTurn(nextPlayerID, goAgain)

	g.SequenceNum++

	return action, nil
}

func (g *Game) startTurn(playerID uuid.UUID, goAgain bool) *ActionStartTurn {
	hand := g.Hands[playerID]
	drawCount := g.Config.TurnDraw
	if hand.Len() == 0 {
		drawCount = g.Config.StartNumCards
	}

	drawn := g.drawCards(playerID, drawCount)

	action := NewActionStartTurn(g.SequenceNum, playerID, drawn, g.Config.MovesPerTurn, g.PlayerDebts(playerID), goAgain)

	return action
}

func (g *Game) StartTurn(playerID uuid.UUID) (*ActionStartTurn, error) {
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

	action := g.startTurn(playerID, false)

	g.SequenceNum++

	return action, nil
}

func (g *Game) forgiveDebts(playerID uuid.UUID) {
	remaining := g.Debts[:0]
	for _, d := range g.Debts {
		if d.DebtorID == playerID {
			g.DebtChips[playerID]++
			continue
		}
		remaining = append(remaining, d)
	}
	g.Debts = remaining
}

func (g *Game) RearrangeProperty(playerID uuid.UUID, cardID Identifier, targetSetIDPtr *Identifier, activeColorPtr *Color) (*ActionRearrangeCard, error) {
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

	err = g.checkDebts(playerID)
	if err != nil {
		return nil, err
	}

	properties := g.Properties[playerID]
	sourceSetIdx, sourceCardIdx := properties.IndexByCardID(cardID)
	if sourceSetIdx == -1 || sourceCardIdx == -1 {
		return nil, errors.PlayerDoesNotHaveCard
	}

	sourceSet := properties[sourceSetIdx]

	card := sourceSet.Cards[sourceCardIdx]
	if card.Category != CategoryPureProperty && card.Category != CategoryWildProperty {
		return nil, errors.InvalidCardForAction
	}

	if activeColorPtr != nil && !card.HasColor(*activeColorPtr) {
		return nil, errors.CardCannotBeAssignedToSet
	}

	var targetSetIdx int
	var targetSet PropertySet
	createNewSet := targetSetIDPtr == nil
	if targetSetIDPtr != nil {
		targetSetID := *targetSetIDPtr
		targetSetIdx = properties.IndexBySetID(targetSetID)
		if targetSetIdx == -1 {
			return nil, errors.PropertySetDoesntExist
		}

		if sourceSetIdx == targetSetIdx {
			if activeColorPtr == nil {
				action := NewActionRearrangeCard(g.SequenceNum, playerID, nil, properties[targetSetIdx], properties.CompleteCount())
				g.SequenceNum++
				return action, nil
			}

			if sourceSet.Cards.Len() != 1 {
				return nil, errors.InvalidCardForAction
			}

			sourceSet.Color = *activeColorPtr
			sourceSet.Cards[sourceCardIdx].ActiveColor = *activeColorPtr
			properties[sourceSetIdx] = sourceSet
			g.Properties[playerID] = properties

			action := NewActionRearrangeCard(g.SequenceNum, playerID, nil, sourceSet, properties.CompleteCount())
			g.SequenceNum++
			return action, nil
		}

		targetSet = properties[targetSetIdx]
		if targetSet.IsComplete() {
			return nil, errors.PropertySetIsComplete
		}

		if !card.HasColor(targetSet.Color) {
			return nil, errors.CardCannotBeAssignedToSet
		}
	}

	card, ok := sourceSet.Cards.RemoveByIdx(sourceCardIdx)
	if !ok {
		return nil, errors.PlayerDoesNotHaveCard
	}

	properties[sourceSetIdx] = sourceSet

	if createNewSet {
		setColor := card.ActiveColor
		if activeColorPtr != nil {
			setColor = *activeColorPtr
		}

		card.ActiveColor = setColor
		newSet := NewPropertySet(g.IDGenerator.New(), setColor)
		newSet.Cards.Add(card)
		properties.Add(newSet)
		targetSet = newSet
	} else {
		card.ActiveColor = targetSet.Color
		targetSet.Cards.Add(card)
		properties[targetSetIdx] = targetSet
	}

	g.Properties[playerID] = properties
	g.cleanProperties(playerID)
	properties = g.Properties[playerID]

	action := NewActionRearrangeCard(g.SequenceNum, playerID, &card, targetSet, properties.CompleteCount())

	g.SequenceNum++

	return action, nil
}

func (g *Game) CheckWinConditions(playerID uuid.UUID) (int, int, bool, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return 0, 0, false, err
	}

	properties := g.Properties[playerID]
	completeCount := properties.CompleteCount()
	if completeCount < g.Config.WinSetAmount {
		return 0, 0, false, nil
	}

	money := g.Money[playerID]
	value := money.Value()
	if value < g.Config.WinMoneyAmount {
		return 0, 0, false, nil
	}

	return completeCount, value, true, nil
}
