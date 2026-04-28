package monopoly_deal

import (
	"fmt"
	"slices"
	"the-deal/internal/errors"
	"the-deal/internal/schema/monopoly_deal_schema"

	"github.com/google/uuid"
)

type Game struct {
	Config        Settings                   `json:"config" msgpack:"a"`
	IDGenerator   *IdentifierGenerator       `json:"id_generator" msgpack:"b"`
	Deck          Deck                       `json:"deck" msgpack:"c"`
	Cards         map[Identifier]Card        `json:"cards" msgpack:"d"`
	Players       []uuid.UUID                `json:"players" msgpack:"e"`
	CurrPlayerIdx int                        `json:"curr_player_idx" msgpack:"f"`
	MovesLeft     int                        `json:"moves_left" msgpack:"g"`
	Hands         map[uuid.UUID]Cards        `json:"hands" msgpack:"i"`
	Money         map[uuid.UUID]Cards        `json:"money" msgpack:"j"`
	Properties    map[uuid.UUID]PropertySets `json:"properties" msgpack:"k"`
	Demands       map[Identifier]Demand      `json:"demands" msgpack:"l"`
	PendingRent   *PendingRent               `json:"pending_rent" msgpack:"m"`
	LastAction    Card                       `json:"last_action" msgpack:"n"`
	SequenceNum   int                        `json:"sequence_num" msgpack:"o"`
}

func NewGame(cfg Settings, playerIDs []uuid.UUID) *Game {
	ig := NewIdentifierGenerator()

	// TODO - maybe shuffle players?

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
		Config:        cfg,
	}

	firstPlayerID := g.Players[g.CurrPlayerIdx]
	_, _ = g.StartTurn(firstPlayerID)

	return g
}

func (g *Game) Proto(playerID uuid.UUID, allPlayerIDs []uuid.UUID) *monopoly_deal_schema.GameState {
	currPlayerID := g.Players[g.CurrPlayerIdx]

	hand := g.Hands[playerID]
	handProto := &monopoly_deal_schema.Hand{
		Cards: hand.Proto(),
	}

	monies := make([]*monopoly_deal_schema.Money, 0, len(allPlayerIDs))
	var properties []*monopoly_deal_schema.PropertySet
	for _, id := range allPlayerIDs {
		money := g.Money[id]
		monies = append(monies, &monopoly_deal_schema.Money{
			PlayerId: id.String(),
			Cards:    money.Proto(),
		})

		property := g.Properties[id]
		properties = append(properties, property.Proto(id)...)
	}

	demandsProto := make([]*monopoly_deal_schema.Demand, 0, len(g.Demands))
	for _, demand := range g.Demands {
		demandsProto = append(demandsProto, demand.Proto())
	}

	var pendingRentProto *monopoly_deal_schema.PendingRent
	if g.PendingRent != nil && g.PendingRent.SourceID == playerID {
		pendingRentProto = g.PendingRent.Proto()
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
		CurrentPlayerId: currPlayerID.String(),
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

	propertySets := g.Properties[playerID]
	complete := 0
	for _, propertySet := range propertySets {
		if propertySet.IsComplete() {
			complete++
		}
	}
	return complete, nil
}

func (g *Game) CountHands(playerID uuid.UUID) (int, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return 0, err
	}

	hand := g.Hands[playerID]
	return hand.Len(), nil
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

func (g *Game) checkPendingRent() error {
	if g.PendingRent != nil {
		return errors.PendingRentExists
	}
	return nil
}

func (g *Game) discardHand(playerID uuid.UUID, cardID Identifier) (Card, error) {
	hand := g.Hands[playerID]
	card, ok := hand.RemoveByID(cardID)
	if !ok {
		return Card{}, errors.PlayerDoesNotHaveCard
	}
	g.Hands[playerID] = hand

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

func (g *Game) removeProperty(playerID uuid.UUID, cardID Identifier) (Card, error) {
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
	properties.Clean()

	g.Properties[playerID] = properties

	return card, nil
}

func (g *Game) checkPlayerHasPropertyCard(playerID uuid.UUID, cardID Identifier) (Card, error) {
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

func (g *Game) checkPlayerHasPropertySet(playerID uuid.UUID, setID Identifier) (PropertySet, error) {
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

func (g *Game) drawCards(playerID uuid.UUID, amount int) Cards {
	cards := g.Deck.Draw(amount)
	hand := g.Hands[playerID]
	hand.Add(cards...)
	g.Hands[playerID] = hand
	return cards
}

func (g *Game) getPayableCards(playerID uuid.UUID) Cards {
	payable := make(Cards, 0)
	payable.Add(g.Money[playerID]...)

	for _, set := range g.Properties[playerID] {
		payable.Add(set.Cards...)
	}

	return payable
}

func (g *Game) PlayMoney(playerID uuid.UUID, cardID Identifier) (*ActionPlayMoney, error) {
	err := g.checkPlayer(playerID)
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

	action := NewActionPlayMoney(g.SequenceNum, playerID, card)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayProperty(playerID uuid.UUID, cardID Identifier, propSetIDPtr *Identifier, activeColorPtr *Color) (*ActionPlayProperty, error) {
	err := g.checkPlayer(playerID)
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

	action := NewActionPlayProperty(g.SequenceNum, playerID, propSet)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayHouse(playerID uuid.UUID, cardID, propSetID Identifier) (*ActionPlayHouse, error) {
	err := g.checkPlayer(playerID)
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

	action := NewActionPlayHouse(g.SequenceNum, playerID, card, propSet)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayHotel(playerID uuid.UUID, cardID, propSetID Identifier) (*ActionPlayHotel, error) {
	err := g.checkPlayer(playerID)
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

	if propSet.HasHotelLast() {
		return nil, errors.PropertySetHasHotel
	}

	if !propSet.HasHouseLast() {
		return nil, errors.PropertySetHasNoHouse
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	propSet.Cards.Add(card)
	properties[setIdx] = propSet
	g.Properties[playerID] = properties

	action := NewActionPlayHotel(g.SequenceNum, playerID, card, propSet)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayPassGo(playerID uuid.UUID, cardID Identifier) (*ActionPlayPassGo, error) {
	err := g.checkPlayer(playerID)
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

	action := NewActionPlayPassGo(g.SequenceNum, playerID, card, cards)

	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayItsMyBirthday(playerID uuid.UUID, cardID Identifier) (*ActionDemandsCreated, error) {
	err := g.checkPlayer(playerID)
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

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandsMap := make(map[Identifier]Demand)
	demandsSlice := make([]Demand, 0)
	for _, targetID := range g.Players {
		if playerID == targetID {
			continue
		}
		id := g.IDGenerator.New()
		demand := NewPaymentDemand(id, playerID, targetID, g.Config.ItsMyBirthdayAmount, DemandSourceItsMyBirthday)
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

func (g *Game) PlayDebtCollector(playerID uuid.UUID, targetID uuid.UUID, cardID Identifier) (*ActionDemandsCreated, error) {
	if playerID == targetID {
		return nil, errors.CannotStealFromSelf
	}

	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkPlayer(targetID)
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

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewPaymentDemand(demandID, playerID, targetID, g.Config.DebtCollectorAmount, DemansSourceDebtCollector)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayRent(playerID uuid.UUID, cardID Identifier) (*ActionPendingRentCreated, error) {
	err := g.checkPlayer(playerID)
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

	propertySets := g.Properties[playerID]
	rent := propertySets.ColorRent(colors...)

	targetIDs := make([]uuid.UUID, 0, len(g.Players)-1)
	for _, player := range g.Players {
		if player == playerID {
			continue
		}
		targetIDs = append(targetIDs, player)
	}

	pendingRent := NewPendingRent(playerID, targetIDs, rent)
	g.PendingRent = &pendingRent

	action := NewActionPendingRentCreated(g.SequenceNum, playerID, card, pendingRent)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayWildRent(playerID uuid.UUID, targetID uuid.UUID, cardID Identifier) (*ActionPendingRentCreated, error) {
	if playerID == targetID {
		return nil, errors.CannotStealFromSelf
	}

	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkPlayer(targetID)
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

	if card.AssetKey != AssetKeyRentWild {
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

	propertySets := g.Properties[playerID]
	rent := propertySets.Rent()

	pendingRent := NewPendingRent(playerID, []uuid.UUID{targetID}, rent)
	g.PendingRent = &pendingRent

	action := NewActionPendingRentCreated(g.SequenceNum, playerID, card, pendingRent)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayDoubleTheRent(playerID uuid.UUID, cardID Identifier) (*ActionPendingRentCreated, error) {
	err := g.checkPlayer(playerID)
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

	if card.AssetKey != AssetKeyDoubleTheRent {
		return nil, errors.InvalidCardForAction
	}

	err = g.checkDemands()
	if err != nil {
		return nil, err
	}

	err = g.checkPendingRent()
	if err == nil {
		return nil, errors.PendingRentDoesntExist
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	pendingRent := g.PendingRent
	pendingRent.DoubleRent()
	g.PendingRent = pendingRent

	action := NewActionPendingRentCreated(g.SequenceNum, playerID, card, *pendingRent)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlaySlyDeal(playerID uuid.UUID, targetID uuid.UUID, cardID Identifier, targetCardID Identifier) (*ActionDemandsCreated, error) {
	if playerID == targetID {
		return nil, errors.CannotStealFromSelf
	}

	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkPlayer(targetID)
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

	if card.AssetKey != AssetKeySlyDeal {
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

	// TODO - check that target card can be removed from their sets

	_, err = g.checkPlayerHasPropertyCard(targetID, targetCardID)
	if err != nil {
		return nil, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewPropertyDemand(demandID, playerID, targetID, nil, targetCardID, DemandSourceSlyDeal)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayForcedDeal(playerID uuid.UUID, targetID uuid.UUID, cardID, sourceCardID, targetCardID Identifier) (*ActionDemandsCreated, error) {
	if playerID == targetID {
		return nil, errors.CannotStealFromSelf
	}

	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkPlayer(targetID)
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

	if card.AssetKey != AssetKeyForcedDeal {
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

	// TODO - check that source and target cards can be removed from their sets

	_, err = g.checkPlayerHasPropertyCard(playerID, sourceCardID)
	if err != nil {
		return nil, err
	}

	_, err = g.checkPlayerHasPropertyCard(targetID, targetCardID)
	if err != nil {
		return nil, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewPropertyDemand(demandID, playerID, targetID, &sourceCardID, targetCardID, DemandSourceForcedDeal)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) PlayDealBreaker(playerID uuid.UUID, targetID uuid.UUID, cardID Identifier, setID Identifier) (*ActionDemandsCreated, error) {
	if playerID == targetID {
		return nil, errors.CannotStealFromSelf
	}

	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	err = g.checkPlayer(targetID)
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

	if card.AssetKey != AssetKeyDealBreaker {
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

	_, err = g.checkPlayerHasPropertySet(targetID, setID)
	if err != nil {
		return nil, err
	}

	_, err = g.discardHand(playerID, cardID)
	if err != nil {
		return nil, err
	}

	demandID := g.IDGenerator.New()
	demand := NewPropertySetDemand(demandID, playerID, targetID, setID, DemandSourceDealBreaker)
	g.Demands = map[Identifier]Demand{demandID: demand}

	action := NewActionDemandsCreated(g.SequenceNum, playerID, &card, nil, demand)

	g.LastAction = card
	g.CompleteMove()
	g.SequenceNum++

	return action, nil
}

func (g *Game) ResolvePendingRent(playerID uuid.UUID) (*ActionDemandsCreated, error) {
	err := g.checkPlayer(playerID)
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

	demandsMap := make(map[Identifier]Demand)
	demandsSlice := make([]Demand, 0, len(pendingRent.TargetIDs))
	for _, targetID := range pendingRent.TargetIDs {
		id := g.IDGenerator.New()
		demand := NewPaymentDemand(id, pendingRent.SourceID, targetID, pendingRent.BaseAmount*pendingRent.Multiplier, DemandSourceRent)
		demandsMap[id] = demand
		demandsSlice = append(demandsSlice, demand)
	}
	g.Demands = demandsMap

	g.PendingRent = nil

	action := NewActionDemandsCreated(g.SequenceNum, playerID, nil, nil, demandsSlice...)

	g.SequenceNum++

	return action, nil
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

func (g *Game) placeTransferredProperty(targetID uuid.UUID, card Card, targetSetID *Identifier) (PropertySet, error) {
	properties := g.Properties[targetID]

	if targetSetID != nil {
		setIdx := properties.IndexBySetID(*targetSetID)
		if setIdx == -1 {
			return PropertySet{}, errors.PropertySetDoesntExist
		}

		set := properties[setIdx]
		if set.IsComplete() {
			return PropertySet{}, errors.PropertySetIsComplete
		}

		if !card.HasColor(set.Color) {
			return PropertySet{}, errors.CardCannotBeAssignedToSet
		}

		card.ActiveColor = set.Color
		set.Cards.Add(card)
		properties[setIdx] = set
		g.Properties[targetID] = properties
		return set, nil
	}

	setID := g.IDGenerator.New()
	set := NewPropertySet(setID, card.ActiveColor)
	set.Cards.Add(card)
	properties.Add(set)
	g.Properties[targetID] = properties

	return set, nil
}

func (g *Game) transferProperty(sourceID, targetID uuid.UUID, cardID Identifier, targetSetID *Identifier) (*Card, *PropertySet, error) {
	card, err := g.removeProperty(sourceID, cardID)
	if err != nil {
		return nil, nil, err
	}

	switch card.Category {
	case CategoryPureProperty, CategoryWildProperty:
		set, err := g.placeTransferredProperty(targetID, card, targetSetID)
		if err != nil {
			return nil, nil, err
		}
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

func (g *Game) transferCards(sourceID, targetID uuid.UUID, cardIDs ...Identifier) (Cards, PropertySets, error) {
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

			cardPtr, setPtr, err := g.transferProperty(sourceID, targetID, cardID, nil)
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

func (g *Game) ComplyPaymentDemand(playerID uuid.UUID, demandID Identifier, cardIDs ...Identifier) (*ActionDemandComplied, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	demand, ok := g.Demands[demandID]
	if !ok {
		return nil, errors.DemandDoesNotExist
	}

	if demand.Kind != DemandKindPayment {
		return nil, errors.DemandDoesNotExist
	}

	if !demand.IsActive {
		delete(g.Demands, demandID)
		action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID, nil, nil, nil)
		g.SequenceNum++
		return action, nil
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

	sourceProperty := g.Properties[demand.SourceID]
	sourceSets := sourceProperty.CompleteCount()

	sourceMonies := g.Money[demand.SourceID]
	sourceMoney := sourceMonies.Value()

	targetProperty := g.Properties[demand.TargetID]
	targetSets := targetProperty.CompleteCount()

	targetMonies := g.Money[demand.TargetID]
	targetMoney := targetMonies.Value()

	delete(g.Demands, demandID)

	transferCards := TransferCards{
		SourceID:     playerID,
		TargetID:     demand.SourceID,
		Cards:        transferredCards,
		PropertySets: transferredSets,
		SourceSets:   sourceSets,
		SourceMoney:  sourceMoney,
		TargetSets:   targetSets,
		TargetMoney:  targetMoney,
	}
	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID, &transferCards, nil, nil)

	g.SequenceNum++

	return action, nil
}

func (g *Game) ComplyPropertyDemand(playerID uuid.UUID, demandID Identifier, targetSetID *Identifier) (*ActionDemandComplied, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	demand, ok := g.Demands[demandID]
	if !ok {
		return nil, errors.DemandDoesNotExist
	}

	if demand.Kind != DemandKindProperty {
		return nil, errors.DemandDoesNotExist
	}

	if !demand.IsActive {
		delete(g.Demands, demandID)
		action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID, nil, nil, nil)
		g.SequenceNum++
		return action, nil
	}

	var sourcePropertySets PropertySets
	var targetPropertySets PropertySets

	// forced-deal path: place source card first so selected target set
	// can still be referenced even if target card removal would empty it.
	if demand.Property.SourceCardID != nil {
		_, targetPropertySet, err := g.transferProperty(demand.SourceID, demand.TargetID, *demand.Property.SourceCardID, targetSetID)
		if err != nil {
			return nil, err
		}

		if targetPropertySet != nil {
			targetPropertySets.Add(*targetPropertySet)
		}
	}

	// target pays source with target card
	_, sourcePropertySets, err = g.transferCards(demand.TargetID, demand.SourceID, demand.Property.TargetCardID)
	if err != nil {
		return nil, err
	}

	delete(g.Demands, demandID)

	transferProperty := TransferProperty{
		SourceID:           playerID,
		TargetID:           demand.SourceID,
		SourcePropertySets: sourcePropertySets,
		TargetPropertySets: targetPropertySets,
	}
	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID, nil, &transferProperty, nil)

	g.SequenceNum++

	return action, nil
}

func (g *Game) transferPropertySet(sourceID, targetID uuid.UUID, propertySetID Identifier) (PropertySet, error) {
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

func (g *Game) ComplyPropertySetDemand(playerID uuid.UUID, demandID Identifier) (*ActionDemandComplied, error) {
	err := g.checkPlayer(playerID)
	if err != nil {
		return nil, err
	}

	demand, ok := g.Demands[demandID]
	if !ok {
		return nil, errors.DemandDoesNotExist
	}

	if demand.Kind != DemandKindPropertySet {
		return nil, errors.DemandDoesNotExist
	}

	if !demand.IsActive {
		delete(g.Demands, demandID)
		action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID, nil, nil, nil)
		g.SequenceNum++
		return action, nil
	}

	transferredSet, err := g.transferPropertySet(demand.TargetID, demand.SourceID, demand.PropertySet.PropertySetID)
	if err != nil {
		return nil, err
	}

	delete(g.Demands, demandID)

	transferPropertySet := TransferPropertySet{
		SourceID:    playerID,
		TargetID:    demand.SourceID,
		PropertySet: transferredSet,
	}
	action := NewActionDemandComplied(g.SequenceNum, playerID, demand.ID, nil, nil, &transferPropertySet)

	g.SequenceNum++

	return action, nil
}

func (g *Game) DenyDemand(playerID uuid.UUID, demandID Identifier, cardID Identifier) (*ActionDemandsCreated, error) {
	err := g.checkPlayer(playerID)
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

	demand, ok := g.Demands[demandID]
	if !ok {
		return nil, errors.DemandDoesNotExist
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
	if isCurrentPlayer {
		g.CompleteMove()
	}
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

	err = g.checkPendingRent()
	if err != nil {
		return nil, err
	}

	hand := g.Hands[playerID]
	if hand.Len() > g.Config.MaxHandSize {
		return nil, errors.PlayerHandHasTooManyCards
	}

	properties := g.Properties[playerID]
	if !properties.Valid() {
		return nil, errors.InvalidPropertySets
	}

	n := len(g.Players)
	g.CurrPlayerIdx = (g.CurrPlayerIdx + 1) % n
	g.MovesLeft = g.Config.MovesPerTurn

	nextPlayerID := g.Players[g.CurrPlayerIdx]
	action := g.startTurn(nextPlayerID)

	g.SequenceNum++

	return action, nil
}

func (g *Game) startTurn(playerID uuid.UUID) *ActionStartTurn {
	hand := g.Hands[playerID]
	drawCount := g.Config.PassGoDraw
	if hand.Len() == 0 {
		drawCount = g.Config.StartNumCards
	}

	drawn := g.drawCards(playerID, drawCount)

	action := NewActionStartTurn(g.SequenceNum, playerID, drawn, g.Config.MovesPerTurn)

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

	err = g.checkPendingRent()
	if err != nil {
		return nil, err
	}

	action := g.startTurn(playerID)

	g.SequenceNum++

	return action, nil
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

	err = g.checkPendingRent()
	if err != nil {
		return nil, err
	}

	properties := g.Properties[playerID]
	sourceSetIdx, sourceCardIdx := properties.IndexByCardID(cardID)
	if sourceSetIdx == -1 || sourceCardIdx == -1 {
		return nil, errors.PlayerDoesNotHaveCard
	}

	sourceSet := properties[sourceSetIdx]

	if sourceSet.IsLocked() && sourceCardIdx != sourceSet.Cards.Len()-1 {
		return nil, errors.InvalidCardForAction
	}

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

	properties.Clean()
	g.Properties[playerID] = properties

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
	if money.Value() < g.Config.WinMoneyAmount {
		return 0, 0, false, nil
	}

	return completeCount, value, true, nil
}
