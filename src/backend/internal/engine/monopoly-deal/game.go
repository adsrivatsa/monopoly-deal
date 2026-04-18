package monopoly_deal

import (
	stderrors "errors"
	"fun-kames/internal/errors"
	"slices"

	"github.com/google/uuid"
)

type Game struct {
	IDGenerator   *IdentifierGenerator        `msgpack:"id_generator"`
	IDTranslator  IdentifierTranslator        `msgpack:"id_translator"`
	Deck          Deck                        `msgpack:"deck"`
	Cards         map[Identifier]Card         `msgpack:"cards"`
	Players       []Identifier                `msgpack:"players"`
	CurrPlayerIdx int                         `msgpack:"curr_player_idx"`
	MovesLeft     int                         `msgpack:"moves_left"`
	Hands         map[Identifier]Cards        `msgpack:"hands"`
	Money         map[Identifier]Cards        `msgpack:"money"`
	Properties    map[Identifier]PropertySets `msgpack:"properties"`
	Demands       map[Identifier]Demand       `msgpack:"demands"`
	PendingRent   *PendingRent                `msgpack:"pending_rent"`
	Config        Settings                    `msgpack:"config"`
}

func NewGame(cfg Settings, playerUUIDs []uuid.UUID) *Game {
	ig := NewIdentifierGenerator()
	it := NewIdentifierTranslator(ig, playerUUIDs)

	// TODO - maybe shuffle players?

	playerIDs := make([]Identifier, len(playerUUIDs))
	for i, playerUUID := range playerUUIDs {
		playerIDs[i], _ = it.GetIdentifier(playerUUID)
	}

	d, cardMap := NewDeck(cfg, ig)

	hands := make(map[Identifier]Cards)
	for _, playerID := range playerIDs {
		hand := d.Draw(cfg.StartNumCards)
		hands[playerID] = hand
	}

	money := make(map[Identifier]Cards)
	for _, playerID := range playerIDs {
		money[playerID] = Cards{}
	}

	properties := make(map[Identifier]PropertySets)
	for _, playerID := range playerIDs {
		properties[playerID] = PropertySets{}
	}

	return &Game{
		IDGenerator:   ig,
		IDTranslator:  it,
		Deck:          d,
		Cards:         cardMap,
		Players:       playerIDs,
		CurrPlayerIdx: 0,
		MovesLeft:     cfg.MovesPerTurn,
		Hands:         hands,
		Money:         money,
		Properties:    properties,
		Demands:       nil,
		Config:        cfg,
	}
}

func (g *Game) checkPlayer(playerID uuid.UUID) (Identifier, error) {
	id, ok := g.IDTranslator.GetIdentifier(playerID)
	if !ok {
		return id, errors.PlayerNotInGame
	}
	return id, nil
}

func (g *Game) checkTurn(playerID Identifier) error {
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

func (g *Game) checkPendingRent() error {
	if g.PendingRent != nil {
		return errors.PendingRentExists
	}
	return nil
}

func (g *Game) discardHand(playerID Identifier, cardID Identifier) error {
	hand := g.Hands[playerID]
	card, ok := hand.RemoveByID(cardID)
	if !ok {
		return errors.PlayerDoesNotHaveCard
	}
	g.Hands[playerID] = hand

	g.Deck.Add(card)

	return nil
}

func (g *Game) removeMoney(playerID Identifier, cardID Identifier) (Card, error) {
	money := g.Money[playerID]
	card, ok := money.RemoveByID(cardID)
	if !ok {
		return card, errors.PlayerDoesNotHaveCard
	}
	g.Money[playerID] = money

	return card, nil
}

func (g *Game) removeProperty(playerID Identifier, cardID Identifier) (Card, error) {
	properties := g.Properties[playerID]
	i, j := properties.IndexByCardID(cardID)
	if i == -1 || j == -1 {
		return Card{}, errors.PlayerDoesNotHaveCard
	}

	set := properties[i]
	if set.IsLocked() && j != set.Cards.Len()-1 {
		return Card{}, errors.InvalidCardForAction
	}

	card, _ := set.Cards.RemoveByIdx(j)

	properties[i] = set

	g.Properties[playerID] = properties

	return card, nil

}

func (g *Game) checkPlayerHasPropertyCard(playerID Identifier, cardID Identifier) (Card, error) {
	card, err := g.checkCard(cardID, CategoryPureProperty, CategoryWildProperty)
	if err != nil {
		return Card{}, err
	}

	properties := g.Properties[playerID]
	i, j := properties.IndexByCardID(cardID)
	if i == -1 || j == -1 {
		return Card{}, errors.PlayerDoesNotHaveCard
	}

	set := properties[i]
	if set.IsComplete() || set.IsLocked() {
		return Card{}, errors.CardCannotBeStolen
	}

	return card, nil
}

func (g *Game) checkPlayerHasPropertySet(playerID Identifier, setID Identifier) (PropertySet, error) {
	properties := g.Properties[playerID]
	i := properties.IndexBySetID(setID)
	if i == -1 {
		return PropertySet{}, errors.PlayerDoesNotHaveCard
	}

	set := properties[i]
	if !set.IsComplete() {
		return PropertySet{}, errors.PropertySetIsNotComplete
	}

	return set, nil
}

func (g *Game) CompleteMove() {
	g.MovesLeft--
}

func (g *Game) drawCards(playerID Identifier, amount int) Cards {
	cards := g.Deck.Draw(amount)
	hand := g.Hands[playerID]
	hand.Add(cards...)
	g.Hands[playerID] = hand
	return cards
}

func (g *Game) getPayableCards(playerID Identifier) Cards {
	payable := make(Cards, 0)
	payable.Add(g.Money[playerID]...)

	for _, set := range g.Properties[playerID] {
		payable.Add(set.Cards...)
	}

	return payable
}

func (g *Game) CompleteTurn(playerUUID uuid.UUID) error {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return err
	}

	err = g.checkDemands()
	if err != nil {
		return err
	}

	err = g.checkPendingRent()
	if err != nil {
		return err
	}

	hand := g.Hands[playerID]
	if hand.Len() > g.Config.MaxHandSize {
		return errors.PlayerHandHasTooManyCards
	}

	properties := g.Properties[playerID]
	if !properties.Valid() {
		return errors.InvalidPropertySets
	}

	n := len(g.Players)
	g.CurrPlayerIdx = (g.CurrPlayerIdx + 1) % n
	g.MovesLeft = g.Config.MovesPerTurn

	return nil
}

func (g *Game) StartTurn(playerUUID uuid.UUID) (Cards, error) {
	playerID, err := g.checkPlayer(playerUUID)
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

	err = g.checkPendingRent()
	if err != nil {
		return nil, err
	}

	hand := g.Hands[playerID]
	drawCount := g.Config.PassGoDraw
	if hand.Len() == 0 {
		drawCount = g.Config.StartNumCards
	}

	drawn := g.drawCards(playerID, drawCount)
	return drawn, nil
}

func (g *Game) DiscardCards(playerUUID uuid.UUID, cardIDs ...Identifier) error {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return err
	}

	err = g.checkDemands()
	if err != nil {
		return err
	}

	err = g.checkPendingRent()
	if err != nil {
		return err
	}

	//if g.MovesLeft > 0 {
	//	return errors.CannotDiscardYet
	//}

	if len(cardIDs) == 0 {
		return errors.InvalidAmountOfCards
	}

	hand := g.Hands[playerID]
	inHand := make(map[Identifier]struct{}, len(hand))
	for _, card := range hand {
		inHand[card.ID] = struct{}{}
	}

	for _, cardID := range cardIDs {
		if _, ok := inHand[cardID]; !ok {
			return errors.PlayerDoesNotHaveCard
		}
	}

	for _, cardID := range cardIDs {
		err = g.discardHand(playerID, cardID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Game) PlayMoney(playerUUID uuid.UUID, cardID Identifier) error {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return err
	}

	card, err := g.checkCard(cardID, CategoryMoney, CategoryAction)
	if err != nil {
		return err
	}

	err = g.checkDemands()
	if err != nil {
		return err
	}

	err = g.checkPendingRent()
	if err != nil {
		return err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return err
	}

	money := g.Money[playerID]
	money.Add(card)
	g.Money[playerID] = money

	g.CompleteMove()

	return nil
}

func (g *Game) PlayProperty(playerUUID uuid.UUID, cardID Identifier, propSetIDPtr *Identifier) (PropertySet, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return PropertySet{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return PropertySet{}, err
	}

	card, err := g.checkCard(cardID, CategoryPureProperty, CategoryWildProperty)
	if err != nil {
		return PropertySet{}, err
	}

	err = g.checkDemands()
	if err != nil {
		return PropertySet{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return PropertySet{}, err
	}

	var setIdx int
	var propSetID Identifier
	var propSet PropertySet
	if propSetIDPtr != nil {
		propSetID = *propSetIDPtr

		properties := g.Properties[playerID]

		setIdx = properties.IndexBySetID(propSetID)
		if setIdx == -1 {
			return PropertySet{}, errors.PropertySetDoesntExist
		}

		propSet = properties[setIdx]

		if propSet.IsComplete() {
			return PropertySet{}, errors.PropertySetIsComplete
		}

		if !card.HasColor(propSet.Color) {
			return PropertySet{}, errors.CardCannotBeAssignedToSet
		}
	} else {
		propSetID = g.IDGenerator.New()

		propSet = NewPropertySet(propSetID, card.ActiveColor)
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return PropertySet{}, err
	}

	propSet.Cards.Add(card)

	var properties PropertySets
	if propSetIDPtr != nil {
		properties = g.Properties[playerID]
		properties[setIdx] = propSet
	} else {
		properties = g.Properties[playerID]
		properties.Add(propSet)
	}
	g.Properties[playerID] = properties

	g.CompleteMove()

	return propSet, nil
}

func (g *Game) RearrangeProperty(playerUUID uuid.UUID, cardID Identifier, targetSetIDPtr *Identifier) (PropertySet, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return PropertySet{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return PropertySet{}, err
	}

	err = g.checkDemands()
	if err != nil {
		return PropertySet{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return PropertySet{}, err
	}

	properties := g.Properties[playerID]
	sourceSetIdx, sourceCardIdx := properties.IndexByCardID(cardID)
	if sourceSetIdx == -1 || sourceCardIdx == -1 {
		return PropertySet{}, errors.PlayerDoesNotHaveCard
	}

	sourceSet := properties[sourceSetIdx]

	if sourceSet.IsLocked() && sourceCardIdx != sourceSet.Cards.Len()-1 {
		return PropertySet{}, errors.InvalidCardForAction
	}

	card := sourceSet.Cards[sourceCardIdx]
	if card.Category != CategoryPureProperty && card.Category != CategoryWildProperty {
		return PropertySet{}, errors.InvalidCardForAction
	}

	var targetSetIdx int
	var targetSet PropertySet
	createNewSet := targetSetIDPtr == nil
	if !createNewSet {
		targetSetID := *targetSetIDPtr
		targetSetIdx = properties.IndexBySetID(targetSetID)
		if targetSetIdx == -1 {
			return PropertySet{}, errors.PropertySetDoesntExist
		}

		if sourceSetIdx == targetSetIdx {
			return properties[targetSetIdx], nil
		}

		targetSet = properties[targetSetIdx]
		if targetSet.IsComplete() {
			return PropertySet{}, errors.PropertySetIsComplete
		}

		if !card.HasColor(targetSet.Color) {
			return PropertySet{}, errors.CardCannotBeAssignedToSet
		}
	}

	card, ok := sourceSet.Cards.RemoveByIdx(sourceCardIdx)
	if !ok {
		return PropertySet{}, errors.PlayerDoesNotHaveCard
	}

	properties[sourceSetIdx] = sourceSet

	if createNewSet {
		newSet := NewPropertySet(g.IDGenerator.New(), card.ActiveColor)
		newSet.Cards.Add(card)
		properties.Add(newSet)
		targetSet = newSet
	} else {
		card.ActiveColor = targetSet.Color
		targetSet.Cards.Add(card)
		properties[targetSetIdx] = targetSet
	}

	properties.Clean()
	g.Properties[playerID] = properties

	return targetSet, nil
}

func (g *Game) PlayPassGo(playerUUID uuid.UUID, cardID Identifier) (Cards, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	if card.AssetKey != AssetKeyPassGo {
		return nil, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return nil, err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	cards := g.drawCards(playerID, g.Config.PassGoDraw)

	g.CompleteMove()

	return cards, nil
}

func (g *Game) PlayDoubleTheRent(playerUUID uuid.UUID, cardID Identifier) error {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return err
	}

	if card.AssetKey != AssetKeyDoubleTheRent {
		return errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return err
	}

	err = g.checkPendingRent()
	if err == nil {
		return errors.PendingRentDoesntExist
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return err
	}

	pendingRent := g.PendingRent
	pendingRent.DoubleRent()
	g.PendingRent = pendingRent

	g.CompleteMove()

	return nil
}

func (g *Game) PlayItsMyBirthday(playerUUID uuid.UUID, cardID Identifier) (map[Identifier]Demand, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	if card.AssetKey != AssetKeyItsMyBirthday {
		return nil, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return nil, err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	g.Demands = NewPaymentDemands(playerID, g.Players, g.Config.ItsMyBirthdayAmount)

	g.CompleteMove()

	return g.Demands, nil
}

func (g *Game) PlayDebtCollector(playerUUID uuid.UUID, targetUUID uuid.UUID, cardID Identifier) (map[Identifier]Demand, error) {
	if playerUUID == targetUUID {
		return nil, errors.CannotStealFromSelf
	}

	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	targetID, err := g.checkPlayer(targetUUID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	if card.AssetKey != AssetKeyDebtCollector {
		return nil, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return nil, err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	g.Demands = map[Identifier]Demand{
		targetID: NewPaymentDemand(playerID, targetID, g.Config.DebtCollectorAmount),
	}

	g.CompleteMove()

	return g.Demands, nil
}

func (g *Game) PlayRent(playerUUID uuid.UUID, cardID Identifier) error {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return err
	}

	var colors []Color
	switch card.AssetKey {
	case AssetKeyRentBrownSky:
		colors = append(colors, ColorBrown, ColorSky)
	case AssetKeyRentPinkOrange:
		colors = append(colors, ColorPink, ColorOrange)
	case AssetKeyRentRedYellow:
		colors = append(colors, ColorRed, ColorYellow)
	case AssetKeyRentGreenBlue:
		colors = append(colors, ColorGreen, ColorBlue)
	case AssetKeyRentUtilityRailroad:
		colors = append(colors, ColorUtility, ColorRailroad)
	default:
		return errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return err
	}

	err = g.checkPendingRent()
	if err != nil {
		return err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return err
	}

	propertySets := g.Properties[playerID]
	rent := propertySets.ColorRent(colors...)

	payers := make([]Identifier, 0, len(g.Players)-1)
	for _, player := range g.Players {
		if player == playerID {
			continue
		}
		payers = append(payers, player)
	}

	g.PendingRent = NewPendingRent(playerID, payers, rent)

	g.CompleteMove()

	return nil
}

func (g *Game) PlayWildRent(playerUUID uuid.UUID, targetUUID uuid.UUID, cardID Identifier) error {
	if playerUUID == targetUUID {
		return errors.CannotStealFromSelf
	}

	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return err
	}

	targetID, err := g.checkPlayer(targetUUID)
	if err != nil {
		return err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return err
	}

	if card.AssetKey != AssetKeyRentWild {
		return errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return err
	}

	err = g.checkPendingRent()
	if err != nil {
		return err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return err
	}

	propertySets := g.Properties[playerID]
	rent := propertySets.Rent()

	g.PendingRent = NewPendingRent(playerID, []Identifier{targetID}, rent)

	g.CompleteMove()

	return nil
}

func (g *Game) PlaySlyDeal(playerUUID uuid.UUID, targetUUID uuid.UUID, cardID Identifier, targetCardID Identifier) (map[Identifier]Demand, error) {
	if playerUUID == targetUUID {
		return nil, errors.CannotStealFromSelf
	}

	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	targetID, err := g.checkPlayer(targetUUID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	if card.AssetKey != AssetKeySlyDeal {
		return nil, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, err
	}

	_, err = g.checkPlayerHasPropertyCard(targetID, targetCardID)
	if err != nil {
		return nil, err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	g.Demands = map[Identifier]Demand{
		targetID: NewPropertyDemand(playerID, targetID, nil, targetCardID),
	}

	g.CompleteMove()

	return g.Demands, nil
}

func (g *Game) PlayForcedDeal(playerUUID uuid.UUID, targetUUID uuid.UUID, cardID, sourceCardID, targetCardID Identifier) (map[Identifier]Demand, error) {
	if playerUUID == targetUUID {
		return nil, errors.CannotStealFromSelf
	}

	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	targetID, err := g.checkPlayer(targetUUID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	if card.AssetKey != AssetKeyForcedDeal {
		return nil, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, err
	}

	_, err = g.checkPlayerHasPropertyCard(playerID, sourceCardID)
	if err != nil {
		return nil, err
	}

	_, err = g.checkPlayerHasPropertyCard(targetID, targetCardID)
	if err != nil {
		return nil, err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	g.Demands = map[Identifier]Demand{
		targetID: NewPropertyDemand(playerID, targetID, &sourceCardID, targetCardID),
	}

	g.CompleteMove()

	return g.Demands, nil
}

func (g *Game) PlayDealBreaker(playerUUID uuid.UUID, targetUUID uuid.UUID, cardID Identifier, setID Identifier) (map[Identifier]Demand, error) {
	if playerUUID == targetUUID {
		return nil, errors.CannotStealFromSelf
	}

	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	targetID, err := g.checkPlayer(targetUUID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	if card.AssetKey != AssetKeyDealBreaker {
		return nil, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, err
	}

	set, err := g.checkPlayerHasPropertySet(targetID, setID)
	if err != nil {
		return nil, err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	g.Demands = map[Identifier]Demand{
		targetID: NewPropertySetDemand(playerID, targetID, set.Cards...),
	}

	g.CompleteMove()

	return g.Demands, nil
}

func (g *Game) DenyDemand(playerUUID uuid.UUID, cardID Identifier) (Demand, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	isCurrentPlayer := g.Players[g.CurrPlayerIdx] == playerID
	if isCurrentPlayer {
		err = g.checkMoves()
		if err != nil {
			return nil, err
		}
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	if card.AssetKey != AssetKeyJustSayNo {
		return nil, errors.InvalidCardForAction
	}

	if g.Demands == nil {
		return nil, errors.DemandDoesNotExist
	}

	demand, ok := g.Demands[playerID]
	if !ok {
		return nil, errors.DemandDoesNotExist
	}

	denied, err := demand.Deny(playerID)
	if err != nil {
		return nil, err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	// flip targets
	oldTarget := demand.GetTarget()
	delete(g.Demands, oldTarget)
	newTarget := denied.GetTarget()
	g.Demands[newTarget] = denied

	if isCurrentPlayer {
		g.CompleteMove()
	}

	return denied, nil
}

func (g *Game) TransferMoney(sourceID Identifier, targetID Identifier, cardID Identifier) (Card, error) {
	card, err := g.removeMoney(sourceID, cardID)
	if err != nil {
		return Card{}, err
	}

	money := g.Money[targetID]
	money.Add(card)
	g.Money[targetID] = money

	return card, nil
}

func (g *Game) TransferProperty(sourceID Identifier, targetID Identifier, cardID Identifier) (*Card, *PropertySet, error) {
	card, err := g.removeProperty(sourceID, cardID)
	if err != nil {
		return nil, nil, err
	}

	switch card.Category {
	case CategoryPureProperty, CategoryWildProperty:
		setID := g.IDGenerator.New()
		set := NewPropertySet(setID, card.ActiveColor)
		set.Cards.Add(card)

		properties := g.Properties[targetID]
		properties.Add(set)
		g.Properties[targetID] = properties
		return nil, &set, nil
	case CategoryAction:
		if card.AssetKey != AssetKeyHouse && card.AssetKey != AssetKeyHotel {
			return nil, nil, errors.InvalidCardForAction
		}

		money := g.Money[targetID]
		money.Add(card)
		g.Money[targetID] = money
		return &card, nil, nil
	default:
		return nil, nil, errors.InvalidCardForAction
	}
}

func canRemovePropertyCard(properties *PropertySets, cardID Identifier) bool {
	i, j := properties.IndexByCardID(cardID)
	if i == -1 || j == -1 {
		return false
	}

	set := (*properties)[i]
	if set.IsLocked() && j != set.Cards.Len()-1 {
		return false
	}

	_, ok := set.Cards.RemoveByIdx(j)
	if !ok {
		return false
	}

	(*properties)[i] = set
	properties.Clean()

	return true
}

func (g *Game) CanTransferCards(sourceID Identifier, cardIDs ...Identifier) error {
	seen := make(map[Identifier]struct{}, len(cardIDs))
	for _, cardID := range cardIDs {
		if _, ok := g.Cards[cardID]; !ok {
			return errors.CardDoesNotExist
		}
		if _, ok := seen[cardID]; ok {
			return errors.InvalidCardForAction
		}
		seen[cardID] = struct{}{}
	}

	sourceMoney := append(Cards(nil), g.Money[sourceID]...)
	sourceProperties := append(PropertySets(nil), g.Properties[sourceID]...)

	pending := append([]Identifier(nil), cardIDs...)
	for len(pending) > 0 {
		nextPending := make([]Identifier, 0, len(pending))
		progress := false

		for _, cardID := range pending {
			if _, ok := sourceMoney.RemoveByID(cardID); ok {
				progress = true
				continue
			}

			if canRemovePropertyCard(&sourceProperties, cardID) {
				progress = true
				continue
			}

			nextPending = append(nextPending, cardID)
		}

		if !progress {
			return errors.InvalidCardForAction
		}

		pending = nextPending
	}

	return nil
}

func (g *Game) TransferCards(sourceID Identifier, targetID Identifier, cardIDs ...Identifier) (Cards, PropertySets, error) {
	var transferredCards Cards
	var transferredSets PropertySets

	err := g.CanTransferCards(sourceID, cardIDs...)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		properties := g.Properties[sourceID]
		properties.Clean()
		g.Properties[sourceID] = properties
	}()

	pending := make([]Identifier, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		pending = append(pending, cardID)
	}

	for len(pending) > 0 {
		nextPending := make([]Identifier, 0, len(pending))
		progress := false

		for _, cardID := range pending {
			card, err := g.TransferMoney(sourceID, targetID, cardID)
			if err == nil {
				transferredCards.Add(card)
				progress = true
				continue
			}

			cardPtr, setPtr, err := g.TransferProperty(sourceID, targetID, cardID)
			if err != nil {
				nextPending = append(nextPending, cardID)
				continue
			}

			if cardPtr != nil {
				transferredCards.Add(*cardPtr)
			}
			if setPtr != nil {
				transferredSets.Add(*setPtr)
			}
			progress = true
		}

		if !progress {
			return nil, nil, errors.InvalidCardForAction
		}

		pending = nextPending
	}

	return transferredCards, transferredSets, nil
}

func (g *Game) TransferPropertySet(sourceID Identifier, targetID Identifier, cardIDs ...Identifier) (PropertySet, error) {
	if len(cardIDs) == 0 {
		return PropertySet{}, errors.InvalidAmountOfCards
	}

	sourceProperties := g.Properties[sourceID]
	setIdx, _ := sourceProperties.IndexByCardID(cardIDs[0])
	if setIdx == -1 {
		return PropertySet{}, errors.PlayerDoesNotHaveCard
	}

	set := sourceProperties[setIdx]
	if len(cardIDs) != set.Cards.Len() {
		return PropertySet{}, errors.InvalidAmountOfCards
	}

	selected := make(map[Identifier]struct{}, len(cardIDs))
	for _, cardID := range cardIDs {
		selected[cardID] = struct{}{}
	}

	for _, card := range set.Cards {
		if _, ok := selected[card.ID]; !ok {
			return PropertySet{}, errors.InvalidCardForAction
		}
	}

	_, ok := sourceProperties.RemoveByIdx(setIdx)
	if !ok {
		return PropertySet{}, errors.PropertySetDoesntExist
	}
	g.Properties[sourceID] = sourceProperties

	transferredSet := set
	transferredSet.ID = g.IDGenerator.New()
	targetProperties := g.Properties[targetID]
	targetProperties.Add(transferredSet)
	g.Properties[targetID] = targetProperties

	return transferredSet, nil
}

func (g *Game) ComplyDemand(playerUUID uuid.UUID, cardIDs ...Identifier) error {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return err
	}

	if g.Demands == nil {
		return errors.DemandDoesNotExist
	}

	demand, ok := g.Demands[playerID]
	if !ok {
		return errors.DemandDoesNotExist
	}

	cards := make(Cards, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		card, ok := g.Cards[cardID]
		if !ok {
			return errors.CardDoesNotExist
		}

		cards = append(cards, card)
	}

	err = demand.IsCompliant(playerID, cards...)
	if err != nil {
		if demand.GetKind() != DemandKindPayment {
			return err
		}

		var e errors.Error
		ok = stderrors.As(err, &e)
		if !ok || e.Code != errors.PaymentDoesNotCoverAmount.Code {
			return err
		}

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
			return err
		}
	}

	sourceID := demand.GetSource()
	targetID := demand.GetTarget()

	switch demand.GetKind() {
	case DemandKindDenied:
		// canceled demand, nothing to transfer
	case DemandKindPayment:
		_, _, err = g.TransferCards(targetID, sourceID, cardIDs...)
		if err != nil {
			return err
		}
	case DemandKindProperty:
		_, _, err = g.TransferCards(targetID, sourceID, cardIDs...)
		if err != nil {
			return err
		}

		propertyDemand, ok := demand.(*PropertyDemand)
		if !ok {
			return errors.InvalidCardForAction
		}

		if propertyDemand.SourceCardID != nil {
			_, _, err = g.TransferCards(sourceID, targetID, *propertyDemand.SourceCardID)
			if err != nil {
				return err
			}
		}
	case DemandKindPropertySet:
		_, err = g.TransferPropertySet(targetID, sourceID, cardIDs...)
		if err != nil {
			return err
		}
	default:
		return errors.InvalidCardForAction
	}

	delete(g.Demands, playerID)
	if len(g.Demands) == 0 {
		g.Demands = nil
	}

	// only non-current players can comply, no need to check for moves

	return nil
}

func (g *Game) ResolvePendingRent(playerUUID uuid.UUID) (map[Identifier]Demand, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	// cant exists, but ok
	err = g.checkDemands()
	if err != nil {
		return nil, err
	}

	err = g.checkPendingRent()
	if err == nil {
		return nil, errors.PendingRentDoesntExist
	}

	pendingRent := g.PendingRent

	g.Demands = NewPaymentDemands(pendingRent.SourceID, pendingRent.TargetIDs, pendingRent.BaseAmount*pendingRent.Multiplier)
	g.PendingRent = nil

	return g.Demands, nil
}
