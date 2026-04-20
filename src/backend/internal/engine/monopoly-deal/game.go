package monopoly_deal

import (
	"fmt"
	"fun-kames/internal/errors"
	"fun-kames/internal/schema/monopoly_deal_schema"
	"slices"

	"github.com/google/uuid"
)

type Game struct {
	Config        Settings                    `json:"config" msgpack:"a"`
	IDGenerator   *IdentifierGenerator        `json:"id_generator" msgpack:"b"`
	IDTranslator  IdentifierTranslator        `json:"id_translator" msgpack:"c"`
	Deck          Deck                        `json:"deck" msgpack:"d"`
	Cards         map[Identifier]Card         `json:"cards" msgpack:"e"`
	Players       []Identifier                `json:"players" msgpack:"f"`
	CurrPlayerIdx int                         `json:"curr_player_idx" msgpack:"g"`
	MovesLeft     int                         `json:"moves_left" msgpack:"h"`
	Hands         map[Identifier]Cards        `json:"hands" msgpack:"i"`
	Money         map[Identifier]Cards        `json:"money" msgpack:"j"`
	Properties    map[Identifier]PropertySets `json:"properties" msgpack:"k"`
	Demands       map[Identifier]Demand       `json:"demands" msgpack:"l"`
	PendingRent   *PendingRent                `json:"pending_rent" msgpack:"m"`
	LastAction    Card                        `json:"last_action" msgpack:"n"`
	SequenceNum   int                         `json:"sequence_num" msgpack:"o"`
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

	g := &Game{
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

	firstPlayerID := g.Players[g.CurrPlayerIdx]
	firstPlayerUUID, _ := g.IDTranslator.GetUUID(firstPlayerID)
	_, _ = g.StartTurn(firstPlayerUUID)

	return g
}

func (g *Game) Proto(playerUUID uuid.UUID, allPlayerUUIDs []uuid.UUID) *monopoly_deal_schema.GameState {
	playerID, _ := g.IDTranslator.GetIdentifier(playerUUID)

	currPlayerID := g.Players[g.CurrPlayerIdx]
	currPlayerUUID, _ := g.IDTranslator.GetUUID(currPlayerID)

	hand := g.Hands[playerID]
	handProto := &monopoly_deal_schema.Hand{
		Cards: hand.Proto(),
	}

	monies := make([]*monopoly_deal_schema.Money, 0, len(allPlayerUUIDs))
	var properties []*monopoly_deal_schema.PropertySet
	for _, u := range allPlayerUUIDs {
		id, _ := g.IDTranslator.GetIdentifier(u)

		money := g.Money[id]
		monies = append(monies, &monopoly_deal_schema.Money{
			PlayerId: u.String(),
			Cards:    money.Proto(),
		})

		property := g.Properties[id]
		properties = append(properties, property.Proto(u)...)
	}

	demand, ok := g.Demands[playerID]
	var demandProto *monopoly_deal_schema.Demand
	if ok {
		sourceUUID, _ := g.IDTranslator.GetUUID(demand.SourceID)
		demandProto = demand.Proto(sourceUUID)
	}

	var pendingRentProto *monopoly_deal_schema.PendingRent
	if g.PendingRent != nil {
		targetIDs := g.PendingRent.TargetIDs
		targetUUIDs := make([]uuid.UUID, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			targetUUID, _ := g.IDTranslator.GetUUID(targetID)
			targetUUIDs = append(targetUUIDs, targetUUID)
		}

		pendingRentProto = g.PendingRent.Proto(playerUUID, targetUUIDs)
	}

	assetKeys := AllAssetKeys()
	assetImages := make([]*monopoly_deal_schema.AssetImage, 0, len(assetKeys))
	for _, assetKey := range assetKeys {
		assetImages = append(assetImages, &monopoly_deal_schema.AssetImage{
			AssetKey: assetKey.Proto(),
			ImageUrl: fmt.Sprintf("http://127.0.0.1:4000/static/card/%s.svg", string(assetKey)), // TODO - this is just for now
		})
	}

	return &monopoly_deal_schema.GameState{
		SeqNum:          int32(g.SequenceNum),
		Players:         nil, // populated by caller
		CurrentPlayerId: currPlayerUUID.String(),
		MovesLeft:       int32(g.MovesLeft),
		YourHand:        handProto,
		Money:           monies,
		Properties:      properties,
		Demand:          demandProto,
		PendingRent:     pendingRentProto,
		LastAction:      g.LastAction.Proto(),
		AssetImages:     assetImages,
	}
}

func (g *Game) CountMoney(playerUUID uuid.UUID) (int, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return 0, err
	}

	money := g.Money[playerID]
	return money.Value(), nil
}

func (g *Game) CountCompletedSets(playerUUID uuid.UUID) (int, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return 0, err
	}

	propertySets := g.Properties[playerID]
	complete := 0
	for _, propertySet := range propertySets {
		if propertySet.IsComplete() {
			complete++
		}
	}
	return complete, nil
}

func (g *Game) CountHands(playerUUID uuid.UUID) (int, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return 0, err
	}

	hand := g.Hands[playerID]
	return hand.Len(), nil
}

func (g *Game) checkPlayer(playerUUID uuid.UUID) (Identifier, error) {
	id, ok := g.IDTranslator.GetIdentifier(playerUUID)
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

func (g *Game) CompleteTurn(playerUUID uuid.UUID) (Cards, uuid.UUID, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}

	err = g.checkDemands()
	if err != nil {
		return nil, uuid.UUID{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return nil, uuid.UUID{}, err
	}

	hand := g.Hands[playerID]
	if hand.Len() > g.Config.MaxHandSize {
		return nil, uuid.UUID{}, errors.PlayerHandHasTooManyCards
	}

	properties := g.Properties[playerID]
	if !properties.Valid() {
		return nil, uuid.UUID{}, errors.InvalidPropertySets
	}

	n := len(g.Players)
	g.CurrPlayerIdx = (g.CurrPlayerIdx + 1) % n
	g.MovesLeft = g.Config.MovesPerTurn

	nextPlayerID := g.Players[g.CurrPlayerIdx]
	nextPlayerUUID, _ := g.IDTranslator.GetUUID(nextPlayerID)
	drawn := g.startTurn(nextPlayerID)

	g.SequenceNum++

	return drawn, nextPlayerUUID, nil
}

func (g *Game) startTurn(playerID Identifier) Cards {
	hand := g.Hands[playerID]
	drawCount := g.Config.PassGoDraw
	if hand.Len() == 0 {
		drawCount = g.Config.StartNumCards
	}

	drawn := g.drawCards(playerID, drawCount)
	return drawn
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

	drawn := g.startTurn(playerID)

	g.SequenceNum++

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

	g.SequenceNum++

	return nil
}

func (g *Game) PlayMoney(playerUUID uuid.UUID, cardID Identifier) (Card, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return Card{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return Card{}, err
	}

	card, err := g.checkCard(cardID, CategoryMoney, CategoryAction)
	if err != nil {
		return Card{}, err
	}

	err = g.checkDemands()
	if err != nil {
		return Card{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return Card{}, err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return Card{}, err
	}

	money := g.Money[playerID]
	money.Add(card)
	g.Money[playerID] = money

	g.CompleteMove()
	g.SequenceNum++

	return card, nil
}

func (g *Game) PlayProperty(playerUUID uuid.UUID, cardID Identifier, propSetIDPtr *Identifier, activeColorPtr *Color) (PropertySet, error) {
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
	resolvedColor := card.ActiveColor
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

		resolvedColor = propSet.Color
	} else {
		if activeColorPtr != nil {
			if !card.HasColor(*activeColorPtr) {
				return PropertySet{}, errors.CardCannotBeAssignedToSet
			}
			resolvedColor = *activeColorPtr
		}

		propSetID = g.IDGenerator.New()
		propSet = NewPropertySet(propSetID, resolvedColor)
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return PropertySet{}, err
	}

	card.ActiveColor = resolvedColor
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
	g.SequenceNum++

	return propSet, nil
}

func (g *Game) RearrangeProperty(playerUUID uuid.UUID, cardID Identifier, targetSetIDPtr *Identifier, activeColorPtr *Color) (PropertySet, error) {
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

	if activeColorPtr != nil && !card.HasColor(*activeColorPtr) {
		return PropertySet{}, errors.CardCannotBeAssignedToSet
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
			if activeColorPtr == nil {
				return properties[targetSetIdx], nil
			}

			if sourceSet.Cards.Len() != 1 {
				return PropertySet{}, errors.InvalidCardForAction
			}

			sourceSet.Color = *activeColorPtr
			sourceSet.Cards[sourceCardIdx].ActiveColor = *activeColorPtr
			properties[sourceSetIdx] = sourceSet
			g.Properties[playerID] = properties

			return sourceSet, nil
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

	properties.Clean()
	g.Properties[playerID] = properties

	g.SequenceNum++

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

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

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

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

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

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

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

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

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

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

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

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

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

	// TODO - check that target card can be removed from their sets

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

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

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

	// TODO - check that source and target cards can be removed from their sets

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

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

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

	_, err = g.checkPlayerHasPropertySet(targetID, setID)
	if err != nil {
		return nil, err
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	g.Demands = map[Identifier]Demand{
		targetID: NewPropertySetDemand(playerID, targetID, setID),
	}

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

	return g.Demands, nil
}

func (g *Game) DenyDemand(playerUUID uuid.UUID, cardID Identifier) (Demand, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return Demand{}, err
	}

	isCurrentPlayer := g.Players[g.CurrPlayerIdx] == playerID
	if isCurrentPlayer {
		err = g.checkMoves()
		if err != nil {
			return Demand{}, err
		}
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return Demand{}, err
	}

	if card.AssetKey != AssetKeyJustSayNo {
		return Demand{}, errors.InvalidCardForAction
	}

	if g.Demands == nil {
		return Demand{}, errors.DemandDoesNotExist
	}

	demand, ok := g.Demands[playerID]
	if !ok {
		return Demand{}, errors.DemandDoesNotExist
	}

	err = g.discardHand(playerID, cardID)
	if err != nil {
		return Demand{}, err
	}

	demand.Deny()

	// flip targets
	delete(g.Demands, playerID)
	targetID := demand.TargetID
	g.Demands[targetID] = demand

	g.LastAction = card

	if isCurrentPlayer {
		g.CompleteMove()
	}
	g.SequenceNum++

	return demand, nil
}

func (g *Game) transferMoney(sourceID Identifier, targetID Identifier, cardID Identifier) (Card, error) {
	card, err := g.removeMoney(sourceID, cardID)
	if err != nil {
		return Card{}, err
	}

	money := g.Money[targetID]
	money.Add(card)
	g.Money[targetID] = money

	return card, nil
}

func (g *Game) transferProperty(sourceID Identifier, targetID Identifier, cardID Identifier) (*Card, *PropertySet, error) {
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

func (g *Game) canTransferCards(sourceID Identifier, cardIDs ...Identifier) error {
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

func (g *Game) transferCards(sourceID Identifier, targetID Identifier, cardIDs ...Identifier) (Cards, PropertySets, error) {
	var transferredCards Cards
	var transferredSets PropertySets

	err := g.canTransferCards(sourceID, cardIDs...)
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
			card, err := g.transferMoney(sourceID, targetID, cardID)
			if err == nil {
				transferredCards.Add(card)
				progress = true
				continue
			}

			cardPtr, setPtr, err := g.transferProperty(sourceID, targetID, cardID)
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

func (g *Game) ComplyPaymentDemand(playerUUID uuid.UUID, cardIDs ...Identifier) (Cards, PropertySets, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, nil, err
	}

	if g.Demands == nil {
		return nil, nil, errors.DemandDoesNotExist
	}

	demand, ok := g.Demands[playerID]
	if !ok {
		return nil, nil, errors.DemandDoesNotExist
	}

	if demand.Kind != DemandKindPayment {
		return nil, nil, errors.DemandDoesNotExist
	}

	if !demand.IsActive {
		delete(g.Demands, playerID)
		if len(g.Demands) == 0 {
			g.Demands = nil
		}
		return nil, nil, nil
	}

	cards := make(Cards, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		card, ok := g.Cards[cardID]
		if !ok {
			return nil, nil, errors.CardDoesNotExist
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
			return nil, nil, errors.PaymentDoesNotCoverAmount
		}
	}

	transferredCards, transferredSets, err := g.transferCards(playerID, demand.SourceID, cardIDs...)
	if err != nil {
		return nil, nil, err
	}

	delete(g.Demands, playerID)
	if len(g.Demands) == 0 {
		g.Demands = nil
	}

	g.SequenceNum++

	return transferredCards, transferredSets, nil
}

func (g *Game) ComplyPropertyDemand(playerUUID uuid.UUID) (PropertySets, PropertySets, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, nil, err
	}

	if g.Demands == nil {
		return nil, nil, errors.DemandDoesNotExist
	}

	demand, ok := g.Demands[playerID]
	if !ok {
		return nil, nil, errors.DemandDoesNotExist
	}

	if demand.Kind != DemandKindProperty {
		return nil, nil, errors.DemandDoesNotExist
	}

	if !demand.IsActive {
		delete(g.Demands, playerID)
		if len(g.Demands) == 0 {
			g.Demands = nil
		}
		return nil, nil, nil
	}

	// target pays source with target card
	_, sourcePropertySets, err := g.transferCards(demand.TargetID, demand.SourceID, demand.Property.TargetCardID)
	if err != nil {
		return nil, nil, err
	}

	// forced-deal path: source pays target with source card
	var targetPropertySets PropertySets
	if demand.Property.SourceCardID != nil {
		_, targetPropertySets, err = g.transferCards(demand.SourceID, demand.TargetID, *demand.Property.SourceCardID)
		if err != nil {
			return nil, nil, err
		}
	}

	delete(g.Demands, playerID)
	if len(g.Demands) == 0 {
		g.Demands = nil
	}

	g.SequenceNum++

	return sourcePropertySets, targetPropertySets, nil

}

func (g *Game) transferPropertySet(sourceID Identifier, targetID Identifier, propertySetID Identifier) (PropertySet, error) {
	sourceProperties := g.Properties[sourceID]
	setIdx := sourceProperties.IndexBySetID(propertySetID)
	if setIdx == -1 {
		return PropertySet{}, errors.PropertySetDoesntExist
	}

	set := sourceProperties[setIdx]
	if !set.IsComplete() {
		return PropertySet{}, errors.PropertySetIsNotComplete
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

func (g *Game) ComplyPropertySetDemand(playerUUID uuid.UUID) (PropertySet, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return PropertySet{}, err
	}

	if g.Demands == nil {
		return PropertySet{}, errors.DemandDoesNotExist
	}

	demand, ok := g.Demands[playerID]
	if !ok {
		return PropertySet{}, errors.DemandDoesNotExist
	}

	if demand.Kind != DemandKindPropertySet {
		return PropertySet{}, errors.DemandDoesNotExist
	}

	if !demand.IsActive {
		delete(g.Demands, playerID)
		if len(g.Demands) == 0 {
			g.Demands = nil
		}
		return PropertySet{}, nil
	}

	transferredSet, err := g.transferPropertySet(demand.TargetID, demand.SourceID, demand.PropertySet.PropertySetID)
	if err != nil {
		return PropertySet{}, err
	}

	delete(g.Demands, playerID)
	if len(g.Demands) == 0 {
		g.Demands = nil
	}

	g.SequenceNum++

	return transferredSet, nil
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

	g.SequenceNum++

	return g.Demands, nil
}
