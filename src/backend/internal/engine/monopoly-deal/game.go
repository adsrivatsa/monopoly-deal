package monopoly_deal

import (
	"fun-kames/internal/errors"
	"slices"

	"github.com/google/uuid"
)

type Game struct {
	Deck          Deck                       `json:"deck"`
	Players       []uuid.UUID                `json:"players"`
	CurrPlayerIdx int                        `json:"curr_player_idx"`
	MovesLeft     int                        `json:"moves_left"`
	Hands         map[uuid.UUID]Cards        `json:"hands"`
	Money         map[uuid.UUID]Cards        `json:"money"`
	Properties    map[uuid.UUID]PropertySets `json:"properties"`
	Config        Settings                   `json:"config"`
}

func NewGame(cfg Settings, playerIDs []uuid.UUID) *Game {
	d := NewDeck(cfg, true)

	hands := make(map[uuid.UUID]Cards)
	for _, playerID := range playerIDs {
		hand := d.Draw(cfg.StartNumCards)
		hands[playerID] = hand
	}

	// TODO - maybe shuffle players?

	return &Game{
		Deck:          d,
		Players:       playerIDs,
		CurrPlayerIdx: 0,
		MovesLeft:     cfg.MovesPerTurn,
		Hands:         hands,
		Money:         make(map[uuid.UUID]Cards),
		Properties:    make(map[uuid.UUID]PropertySets),
		Config:        cfg,
	}
}

func (g *Game) CheckPlayerExists(playerID uuid.UUID) error {
	_, ok := g.Hands[playerID]
	if !ok {
		return errors.PlayerNotInGame
	}
	return nil
}

func (g *Game) CompleteTurn() error {
	// TODO - don't check here, doesn't give an opportunity to discard cards when 3 moves over
	currPlayer := g.Players[g.CurrPlayerIdx]
	hand := g.Hands[currPlayer]
	if hand.Len() > g.Config.MaxHandSize {
		return errors.PlayerHandHasTooManyCards
	}

	n := len(g.Players)
	g.CurrPlayerIdx = (g.CurrPlayerIdx + 1) % n
	g.MovesLeft = g.Config.MovesPerTurn

	return nil
}

func (g *Game) MoveSandwich(playerID uuid.UUID, ck CardKey, fn func(card Card) error, validCategories ...Category) error {
	err := g.CheckPlayerExists(playerID)
	if err != nil {
		return err
	}

	playerIdx := slices.Index(g.Players, playerID)
	if playerIdx == -1 {
		return errors.PlayerNotInGame
	}

	if playerIdx != g.CurrPlayerIdx {
		return errors.NotPlayersTurn
	}

	card, ok := CardByKey[ck]
	if !ok {
		return errors.CardDoesNotExist
	}

	valid := false
	for _, cat := range validCategories {
		valid = valid || (cat == card.Category)
	}
	if !valid {
		return errors.InvalidCardForAction
	}

	if g.MovesLeft <= 0 {
		return errors.NoMovesLeft
	}

	err = fn(card)
	if err != nil {
		return err
	}

	g.MovesLeft--

	return nil
}

func (g *Game) PlayMoney(playerID uuid.UUID, ck CardKey) error {
	return g.MoveSandwich(playerID, ck, func(card Card) error {
		hand := g.Hands[playerID]
		_, ok := hand.Remove(ck)
		if !ok {
			return errors.PlayerDoesNotHaveCard
		}
		g.Hands[playerID] = hand

		money := g.Money[playerID]
		money.Add(card)
		g.Money[playerID] = money

		return nil
	}, CategoryMoney, CategoryAction)
}

func (g *Game) PlayProperty(playerID uuid.UUID, ck CardKey, propertySetID *uuid.UUID) (PropertySet, error) {
	var propertySet PropertySet
	err := g.MoveSandwich(playerID, ck, func(card Card) error {
		if propertySetID != nil {
			properties := g.Properties[playerID]

			setIdx := properties.Index(*propertySetID)
			if setIdx == -1 {
				return errors.PropertySetDoesntExist
			}

			propertySet = properties[setIdx]

			if propertySet.IsFull() {
				return errors.PropertySetIsFull
			}

			if !card.HasColor(propertySet.Color) {
				return errors.CardCannotBeAssignedToSet
			}

			hand := g.Hands[playerID]
			_, ok := hand.Remove(ck)
			if !ok {
				return errors.PlayerDoesNotHaveCard
			}
			g.Hands[playerID] = hand

			propertySet.Cards.Add(card)
			properties[setIdx] = propertySet
			g.Properties[playerID] = properties

			return nil
		}

		// create a new set
		propertySet = NewPropertySet(card)

		properties := g.Properties[playerID]

		hand := g.Hands[playerID]
		_, ok := hand.Remove(ck)
		if !ok {
			return errors.PlayerDoesNotHaveCard
		}
		g.Hands[playerID] = hand

		properties.Add(propertySet)
		g.Properties[playerID] = properties

		return nil
	}, CategoryPureProperty, CategoryWildProperty)
	return propertySet, err
}

func (g *Game) PlayPassGo(playerID uuid.UUID) (Cards, error) {
	var cards Cards
	err := g.MoveSandwich(playerID, CardKeyPassGo, func(card Card) error {
		hand := g.Hands[playerID]
		_, ok := hand.Remove(CardKeyPassGo)
		if !ok {
			return errors.PlayerDoesNotHaveCard
		}

		cards = g.Deck.Draw(g.Config.PassGoDraw)
		hand.Add(cards...)
		g.Hands[playerID] = hand
		return nil
	}, CategoryAction)

	return cards, err
}

//func (g *Game) PlayItsMyBirthday(playerID uuid.UUID) (Demand, error) {
//	err := g.MoveSandwich(playerID, CardKeyItsMyBirthday, func(card Card) error {
//		hand := g.Hands[playerID]
//		_, ok := hand.Remove(CardKeyItsMyBirthday)
//		if !ok {
//			return errors.PlayerDoesNotHaveCard
//		}
//		g.Hands[playerID] = hand
//	}, CategoryAction)
//}
