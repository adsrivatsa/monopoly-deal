package monopoly_deal

import (
	"fmt"
	"slices"
	"the-deal/internal/errors"
	"the-deal/internal/schema/monopoly_deal_schema"

	"github.com/google/uuid"
)

type Game struct {
	Config       Settings             `json:"config" msgpack:"a"`
	IDGenerator  *IdentifierGenerator `json:"id_generator" msgpack:"b"`
	IDTranslator IdentifierTranslator `json:"id_translator" msgpack:"c"`
	Deck         Deck                 `json:"deck" msgpack:"d"`

	// Identifier is the cardID
	Cards map[Identifier]Card `json:"cards" msgpack:"e"`

	Players       []Identifier `json:"players" msgpack:"f"`
	CurrPlayerIdx int          `json:"curr_player_idx" msgpack:"g"`
	MovesLeft     int          `json:"moves_left" msgpack:"h"`

	// Identifier is the playerID
	Hands map[Identifier]Cards `json:"hands" msgpack:"i"`

	// Identifier is the playerID
	Money map[Identifier]Cards `json:"money" msgpack:"j"`

	// Identifier is the playerID
	Properties map[Identifier]PropertySets `json:"properties" msgpack:"k"`

	// Identifier is the demandID
	Demands map[Identifier]Demand `json:"demands" msgpack:"l"`

	PendingRent *PendingRent `json:"pending_rent" msgpack:"m"`
	LastAction  Card         `json:"last_action" msgpack:"n"`
	SequenceNum int          `json:"sequence_num" msgpack:"o"`
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
		Demands:       make(map[Identifier]Demand),
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

	demandsProto := make([]*monopoly_deal_schema.Demand, 0, len(g.Demands))
	for _, demand := range g.Demands {
		if demand.TargetID != playerID {
			continue
		}
		sourceUUID, _ := g.IDTranslator.GetUUID(demand.SourceID)
		demandsProto = append(demandsProto, demand.Proto(sourceUUID, playerUUID))
	}

	var pendingRentProto *monopoly_deal_schema.PendingRent
	if g.PendingRent != nil && g.PendingRent.SourceID == playerID {
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
			ImageUrl: fmt.Sprintf("https://deal-backend.adsrivatsa.com/static/card/%s.svg", string(assetKey)), // TODO - this is just for now
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
		Demands:         demandsProto,
		PendingRent:     pendingRentProto,
		LastAction:      g.LastAction.Proto(),
		AssetImages:     assetImages,
		MaxHandSize:     int32(g.Config.MaxHandSize),
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

func (g *Game) discardHand(playerID Identifier, cardID Identifier) (Card, error) {
	hand := g.Hands[playerID]
	card, ok := hand.RemoveByID(cardID)
	if !ok {
		return Card{}, errors.PlayerDoesNotHaveCard
	}
	g.Hands[playerID] = hand

	g.Deck.Add(card)

	return card, nil
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

func (g *Game) DiscardCards(playerUUID uuid.UUID, cardIDs ...Identifier) (Cards, error) {
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

	//if g.MovesLeft > 0 {
	//	return errors.CannotDiscardYet
	//}

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

	cards := make(Cards, 0, len(inHand))
	for _, cardID := range cardIDs {
		card, err := g.discardHand(playerID, cardID)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	g.SequenceNum++

	return cards, nil
}

func (g *Game) PlayMoney(playerUUID uuid.UUID, cardID Identifier) (*ActionPlayMoney, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkMoves()
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryMoney, CategoryAction)
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

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	money := g.Money[playerID]
	money.Add(card)
	g.Money[playerID] = money

	action := NewActionPlayMoney(g.SequenceNum, playerUUID, card)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayProperty(playerUUID uuid.UUID, cardID Identifier, propSetIDPtr *Identifier, activeColorPtr *Color) (*ActionPlayProperty, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkMoves()
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryPureProperty, CategoryWildProperty)
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

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
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

	action := NewActionPlayProperty(g.SequenceNum, playerUUID, propSet)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayHouse(playerUUID uuid.UUID, cardID, propSetID Identifier) (*ActionPlayHouse, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	if card.AssetKey != AssetKeyHouse {
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

	properties := g.Properties[playerID]
	setIdx := properties.IndexBySetID(propSetID)
	if setIdx == -1 {
		return nil, errors.PropertySetDoesntExist
	}

	propSet := properties[setIdx]

	if !propSet.IsComplete() {
		return nil, errors.PropertySetIsNotComplete
	}

	if propSet.IsLocked() {
		return nil, errors.PropertySetHasHouse
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	propSet.Cards.Add(card)
	properties[setIdx] = propSet
	g.Properties[playerID] = properties

	action := NewActionPlayHouse(g.SequenceNum, playerUUID, card, propSet)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayHotel(playerUUID uuid.UUID, cardID, propSetID Identifier) (*ActionPlayHotel, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, err
	}

	if card.AssetKey != AssetKeyHotel {
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

	properties := g.Properties[playerID]
	setIdx := properties.IndexBySetID(propSetID)
	if setIdx == -1 {
		return nil, errors.PropertySetDoesntExist
	}

	propSet := properties[setIdx]

	if !propSet.IsComplete() {
		return nil, errors.PropertySetIsNotComplete
	}

	if propSet.IsLocked() {
		return nil, errors.PropertySetHasHotel
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	propSet.Cards.Add(card)
	properties[setIdx] = propSet
	g.Properties[playerID] = properties

	action := NewActionPlayHotel(g.SequenceNum, playerUUID, card, propSet)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) RearrangeProperty(playerUUID uuid.UUID, cardID Identifier, targetSetIDPtr *Identifier, activeColorPtr *Color) (PropertySet, int, Card, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return PropertySet{}, 0, Card{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return PropertySet{}, 0, Card{}, err
	}

	err = g.checkDemands()
	if err != nil {
		return PropertySet{}, 0, Card{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return PropertySet{}, 0, Card{}, err
	}

	properties := g.Properties[playerID]
	sourceSetIdx, sourceCardIdx := properties.IndexByCardID(cardID)
	if sourceSetIdx == -1 || sourceCardIdx == -1 {
		return PropertySet{}, 0, Card{}, errors.PlayerDoesNotHaveCard
	}

	sourceSet := properties[sourceSetIdx]

	if sourceSet.IsLocked() && sourceCardIdx != sourceSet.Cards.Len()-1 {
		return PropertySet{}, 0, Card{}, errors.InvalidCardForAction
	}

	card := sourceSet.Cards[sourceCardIdx]
	if card.Category != CategoryPureProperty && card.Category != CategoryWildProperty {
		return PropertySet{}, 0, Card{}, errors.InvalidCardForAction
	}

	if activeColorPtr != nil && !card.HasColor(*activeColorPtr) {
		return PropertySet{}, 0, Card{}, errors.CardCannotBeAssignedToSet
	}

	var targetSetIdx int
	var targetSet PropertySet
	createNewSet := targetSetIDPtr == nil
	if targetSetIDPtr != nil {
		targetSetID := *targetSetIDPtr
		targetSetIdx = properties.IndexBySetID(targetSetID)
		if targetSetIdx == -1 {
			return PropertySet{}, 0, Card{}, errors.PropertySetDoesntExist
		}

		if sourceSetIdx == targetSetIdx {
			if activeColorPtr == nil {
				return properties[targetSetIdx], properties.CompleteCount(), Card{}, nil
			}

			if sourceSet.Cards.Len() != 1 {
				return PropertySet{}, 0, Card{}, errors.InvalidCardForAction
			}

			sourceSet.Color = *activeColorPtr
			sourceSet.Cards[sourceCardIdx].ActiveColor = *activeColorPtr
			properties[sourceSetIdx] = sourceSet
			g.Properties[playerID] = properties

			return sourceSet, properties.CompleteCount(), Card{}, nil
		}

		targetSet = properties[targetSetIdx]
		if targetSet.IsComplete() {
			return PropertySet{}, 0, Card{}, errors.PropertySetIsComplete
		}

		if !card.HasColor(targetSet.Color) {
			return PropertySet{}, 0, Card{}, errors.CardCannotBeAssignedToSet
		}
	}

	card, ok := sourceSet.Cards.RemoveByIdx(sourceCardIdx)
	if !ok {
		return PropertySet{}, 0, Card{}, errors.PlayerDoesNotHaveCard
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

	return targetSet, properties.CompleteCount(), card, nil
}

func (g *Game) PlayPassGo(playerUUID uuid.UUID, cardID Identifier) (*ActionPlayPassGo, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkMoves()
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

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	cards := g.drawCards(playerID, g.Config.PassGoDraw)

	g.LastAction = card

	action := NewActionPlayPassGo(g.SequenceNum, playerUUID, cards, card)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayDoubleTheRent(playerUUID uuid.UUID, cardID Identifier) (PendingRent, Card, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	err = g.checkMoves()
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	if card.AssetKey != AssetKeyDoubleTheRent {
		return PendingRent{}, Card{}, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	err = g.checkPendingRent()
	if err == nil {
		return PendingRent{}, Card{}, errors.PendingRentDoesntExist
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	pendingRent := g.PendingRent
	pendingRent.DoubleRent()
	g.PendingRent = pendingRent

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

	return *pendingRent, card, nil
}

func (g *Game) PlayItsMyBirthday(playerUUID uuid.UUID, cardID Identifier) (map[Identifier]Demand, Card, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkMoves()
	if err != nil {
		return nil, Card{}, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, Card{}, err
	}

	if card.AssetKey != AssetKeyItsMyBirthday {
		return nil, Card{}, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return nil, Card{}, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, Card{}, err
	}

	g.Demands = NewPaymentDemands(g.IDGenerator, playerID, g.Players, g.Config.ItsMyBirthdayAmount, DemandSourceItsMyBirthday)

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

	return g.Demands, card, nil
}

func (g *Game) PlayDebtCollector(playerUUID uuid.UUID, targetUUID uuid.UUID, cardID Identifier) (map[Identifier]Demand, Card, error) {
	if playerUUID == targetUUID {
		return nil, Card{}, errors.CannotStealFromSelf
	}

	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, Card{}, err
	}

	targetID, err := g.checkPlayer(targetUUID)
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkMoves()
	if err != nil {
		return nil, Card{}, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, Card{}, err
	}

	if card.AssetKey != AssetKeyDebtCollector {
		return nil, Card{}, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return nil, Card{}, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, Card{}, err
	}

	demandID := g.IDGenerator.New()
	g.Demands = map[Identifier]Demand{
		demandID: NewPaymentDemand(demandID, playerID, targetID, g.Config.DebtCollectorAmount, DemansSourceDebtCollector),
	}

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

	return g.Demands, card, nil
}

func (g *Game) PlayRent(playerUUID uuid.UUID, cardID Identifier) (PendingRent, Card, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	err = g.checkMoves()
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return PendingRent{}, Card{}, err
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
		return PendingRent{}, Card{}, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return PendingRent{}, Card{}, err
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

	pendingRent := NewPendingRent(playerID, payers, rent)
	g.PendingRent = &pendingRent

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

	return pendingRent, card, nil
}

func (g *Game) PlayWildRent(playerUUID uuid.UUID, targetUUID uuid.UUID, cardID Identifier) (PendingRent, Card, error) {
	if playerUUID == targetUUID {
		return PendingRent{}, Card{}, errors.CannotStealFromSelf
	}

	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	targetID, err := g.checkPlayer(targetUUID)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	err = g.checkMoves()
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	if card.AssetKey != AssetKeyRentWild {
		return PendingRent{}, Card{}, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return PendingRent{}, Card{}, err
	}

	propertySets := g.Properties[playerID]
	rent := propertySets.Rent()

	pendingRent := NewPendingRent(playerID, []Identifier{targetID}, rent)
	g.PendingRent = &pendingRent

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

	return pendingRent, card, nil
}

func (g *Game) PlaySlyDeal(playerUUID uuid.UUID, targetUUID uuid.UUID, cardID Identifier, targetCardID Identifier) (map[Identifier]Demand, Card, error) {
	if playerUUID == targetUUID {
		return nil, Card{}, errors.CannotStealFromSelf
	}

	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, Card{}, err
	}

	targetID, err := g.checkPlayer(targetUUID)
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkMoves()
	if err != nil {
		return nil, Card{}, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, Card{}, err
	}

	if card.AssetKey != AssetKeySlyDeal {
		return nil, Card{}, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return nil, Card{}, err
	}

	// TODO - check that target card can be removed from their sets

	_, err = g.checkPlayerHasPropertyCard(targetID, targetCardID)
	if err != nil {
		return nil, Card{}, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, Card{}, err
	}

	demandID := g.IDGenerator.New()
	g.Demands = map[Identifier]Demand{
		demandID: NewPropertyDemand(demandID, playerID, targetID, nil, targetCardID, DemandSourceSlyDeal),
	}

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

	return g.Demands, card, nil
}

func (g *Game) PlayForcedDeal(playerUUID uuid.UUID, targetUUID uuid.UUID, cardID, sourceCardID, targetCardID Identifier) (map[Identifier]Demand, Card, error) {
	if playerUUID == targetUUID {
		return nil, Card{}, errors.CannotStealFromSelf
	}

	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, Card{}, err
	}

	targetID, err := g.checkPlayer(targetUUID)
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkMoves()
	if err != nil {
		return nil, Card{}, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, Card{}, err
	}

	if card.AssetKey != AssetKeyForcedDeal {
		return nil, Card{}, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return nil, Card{}, err
	}

	// TODO - check that source and target cards can be removed from their sets

	_, err = g.checkPlayerHasPropertyCard(playerID, sourceCardID)
	if err != nil {
		return nil, Card{}, err
	}

	_, err = g.checkPlayerHasPropertyCard(targetID, targetCardID)
	if err != nil {
		return nil, Card{}, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, Card{}, err
	}

	demandID := g.IDGenerator.New()
	g.Demands = map[Identifier]Demand{
		demandID: NewPropertyDemand(demandID, playerID, targetID, &sourceCardID, targetCardID, DemandSourceForcedDeal),
	}

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

	return g.Demands, card, nil
}

func (g *Game) PlayDealBreaker(playerUUID uuid.UUID, targetUUID uuid.UUID, cardID Identifier, setID Identifier) (map[Identifier]Demand, Card, error) {
	if playerUUID == targetUUID {
		return nil, Card{}, errors.CannotStealFromSelf
	}

	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return nil, Card{}, err
	}

	targetID, err := g.checkPlayer(targetUUID)
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkTurn(playerID)
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkMoves()
	if err != nil {
		return nil, Card{}, err
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return nil, Card{}, err
	}

	if card.AssetKey != AssetKeyDealBreaker {
		return nil, Card{}, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, Card{}, err
	}

	err = g.checkPendingRent()
	if err != nil {
		return nil, Card{}, err
	}

	_, err = g.checkPlayerHasPropertySet(targetID, setID)
	if err != nil {
		return nil, Card{}, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, Card{}, err
	}

	demandID := g.IDGenerator.New()
	g.Demands = map[Identifier]Demand{
		demandID: NewPropertySetDemand(demandID, playerID, targetID, setID, DemandSourceDealBreaker),
	}

	g.LastAction = card

	g.CompleteMove()
	g.SequenceNum++

	return g.Demands, card, nil
}

func (g *Game) DenyDemand(playerUUID uuid.UUID, demandID Identifier, cardID Identifier) (Demand, Card, error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return Demand{}, Card{}, err
	}

	isCurrentPlayer := g.Players[g.CurrPlayerIdx] == playerID
	if isCurrentPlayer {
		err = g.checkMoves()
		if err != nil {
			return Demand{}, Card{}, err
		}
	}

	card, err := g.checkCard(cardID, CategoryAction)
	if err != nil {
		return Demand{}, Card{}, err
	}

	if card.AssetKey != AssetKeyJustSayNo {
		return Demand{}, Card{}, errors.InvalidCardForAction
	}

	demand, ok := g.Demands[demandID]
	if !ok {
		return Demand{}, Card{}, errors.DemandDoesNotExist
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return Demand{}, Card{}, err
	}

	id := g.IDGenerator.New()
	demand.Deny(id)
	delete(g.Demands, demandID)
	g.Demands[id] = demand

	g.LastAction = card

	if isCurrentPlayer {
		g.CompleteMove()
	}
	g.SequenceNum++

	return demand, card, nil
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

	cards := set.Cards
	set.Cards = append(append(Cards(nil), cards[:j]...), cards[j+1:]...)

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

func (g *Game) ComplyPaymentDemand(playerUUID uuid.UUID, demandID Identifier, cardIDs ...Identifier) (sourceUUID uuid.UUID, transferredCards Cards, transferredSets PropertySets, sourceSets int, sourceMoney int, targetSets int, targetMoney int, err error) {
	playerID, err := g.checkPlayer(playerUUID)
	if err != nil {
		return
	}

	demand, ok := g.Demands[demandID]
	if !ok {
		err = errors.DemandDoesNotExist
		return
	}

	if demand.Kind != DemandKindPayment {
		err = errors.DemandDoesNotExist
		return
	}

	sourceUUID, _ = g.IDTranslator.GetUUID(demand.SourceID)

	if !demand.IsActive {
		delete(g.Demands, demandID)
		return
	}

	cards := make(Cards, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		card, ok := g.Cards[cardID]
		if !ok {
			err = errors.CardDoesNotExist
			return
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
			err = errors.PaymentDoesNotCoverAmount
			return
		}
	}

	transferredCards, transferredSets, err = g.transferCards(playerID, demand.SourceID, cardIDs...)
	if err != nil {
		return
	}

	sourceProperty := g.Properties[demand.SourceID]
	sourceSets = sourceProperty.CompleteCount()

	sourceMonies := g.Money[demand.SourceID]
	sourceMoney = sourceMonies.Value()

	targetProperty := g.Properties[demand.TargetID]
	targetSets = targetProperty.CompleteCount()

	targetMonies := g.Money[demand.TargetID]
	targetMoney = targetMonies.Value()

	delete(g.Demands, demandID)

	g.SequenceNum++

	return
}

func (g *Game) ComplyPropertyDemand(playerUUID uuid.UUID, demandID Identifier) (uuid.UUID, PropertySets, PropertySets, error) {
	_, err := g.checkPlayer(playerUUID)
	if err != nil {
		return uuid.UUID{}, nil, nil, err
	}

	demand, ok := g.Demands[demandID]
	if !ok {
		return uuid.UUID{}, nil, nil, errors.DemandDoesNotExist
	}

	if demand.Kind != DemandKindProperty {
		return uuid.UUID{}, nil, nil, errors.DemandDoesNotExist
	}

	sourceUUID, _ := g.IDTranslator.GetUUID(demand.SourceID)

	if !demand.IsActive {
		delete(g.Demands, demandID)
		return sourceUUID, nil, nil, nil
	}

	// target pays source with target card
	_, sourcePropertySets, err := g.transferCards(demand.TargetID, demand.SourceID, demand.Property.TargetCardID)
	if err != nil {
		return uuid.UUID{}, nil, nil, err
	}

	// forced-deal path: source pays target with source card
	var targetPropertySets PropertySets
	if demand.Property.SourceCardID != nil {
		_, targetPropertySets, err = g.transferCards(demand.SourceID, demand.TargetID, *demand.Property.SourceCardID)
		if err != nil {
			return uuid.UUID{}, nil, nil, err
		}
	}

	delete(g.Demands, demandID)

	g.SequenceNum++

	return sourceUUID, sourcePropertySets, targetPropertySets, nil
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

	targetProperties := g.Properties[targetID]
	targetProperties.Add(set)
	g.Properties[targetID] = targetProperties

	return set, nil
}

func (g *Game) ComplyPropertySetDemand(playerUUID uuid.UUID, demandID Identifier) (uuid.UUID, PropertySet, error) {
	_, err := g.checkPlayer(playerUUID)
	if err != nil {
		return uuid.UUID{}, PropertySet{}, err
	}

	demand, ok := g.Demands[demandID]
	if !ok {
		return uuid.UUID{}, PropertySet{}, errors.DemandDoesNotExist
	}

	if demand.Kind != DemandKindPropertySet {
		return uuid.UUID{}, PropertySet{}, errors.DemandDoesNotExist
	}

	sourceUUID, _ := g.IDTranslator.GetUUID(demand.SourceID)

	if !demand.IsActive {
		delete(g.Demands, demandID)
		return sourceUUID, PropertySet{}, nil
	}

	transferredSet, err := g.transferPropertySet(demand.TargetID, demand.SourceID, demand.PropertySet.PropertySetID)
	if err != nil {
		return uuid.UUID{}, PropertySet{}, err
	}

	delete(g.Demands, demandID)

	g.SequenceNum++

	return sourceUUID, transferredSet, nil
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

	g.Demands = NewPaymentDemands(g.IDGenerator, pendingRent.SourceID, pendingRent.TargetIDs, pendingRent.BaseAmount*pendingRent.Multiplier, DemandSourceRent)
	g.PendingRent = nil

	g.SequenceNum++

	return g.Demands, nil
}

func (g *Game) CheckWinConditions(playerUUID uuid.UUID) (int, int, bool, error) {
	playerID, err := g.checkPlayer(playerUUID)
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
	if money.Value() < g.Config.WinMoneyAmount {
		return 0, 0, false, nil
	}

	return completeCount, value, true, nil
}
