package main

import (
	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"

	"github.com/google/uuid"
)

// stateBuilder edits a decoded engine Game in place, keeping the id->card
// registry (g.Cards) consistent. Every card it mints gets a fresh unique id
// from the game's own IdentifierGenerator so ids never collide with existing
// cards or with the deck.
type stateBuilder struct {
	g *deal_no_mercy.Game
}

func newBuilder(g *deal_no_mercy.Game) *stateBuilder {
	return &stateBuilder{g: g}
}

// newCard mints a card of the given asset key with a fresh id, registering it
// in g.Cards. It copies the canonical definition (category/value/colors).
func (b *stateBuilder) newCard(key deal_no_mercy.AssetKey) deal_no_mercy.Card {
	def := deal_no_mercy.CardByAssetKey[key]
	id := b.g.IDGenerator.New()
	card := deal_no_mercy.NewCard(id, def.Category, key, def.Value, def.Colors...)
	b.g.Cards[id] = card
	return card
}

// setHand replaces a player's hand with fresh cards of the given keys.
func (b *stateBuilder) setHand(pid uuid.UUID, keys ...deal_no_mercy.AssetKey) deal_no_mercy.Cards {
	// remove old hand cards from the registry to avoid orphan growth
	for _, c := range b.g.Hands[pid] {
		delete(b.g.Cards, c.ID)
	}
	hand := deal_no_mercy.Cards{}
	for _, k := range keys {
		hand.Add(b.newCard(k))
	}
	b.g.Hands[pid] = hand
	return hand
}

// setBank replaces a player's money pile with fresh cards of the given keys.
func (b *stateBuilder) setBank(pid uuid.UUID, keys ...deal_no_mercy.AssetKey) deal_no_mercy.Cards {
	for _, c := range b.g.Money[pid] {
		delete(b.g.Cards, c.ID)
	}
	bank := deal_no_mercy.Cards{}
	for _, k := range keys {
		bank.Add(b.newCard(k))
	}
	b.g.Money[pid] = bank
	return bank
}

// addSet appends a property set (its own color) of fresh cards to a player,
// optionally with a shack. Returns the set (including its assigned id).
func (b *stateBuilder) addSet(pid uuid.UUID, color deal_no_mercy.Color, withShack bool, keys ...deal_no_mercy.AssetKey) deal_no_mercy.PropertySet {
	setID := b.g.IDGenerator.New()
	set := deal_no_mercy.NewPropertySet(setID, color)
	for _, k := range keys {
		c := b.newCard(k)
		c.ActiveColor = color
		set.Cards.Add(c)
	}
	if withShack {
		shack := b.newCard(deal_no_mercy.AssetKeyShack)
		set.Shack = &shack
	}
	props := b.g.Properties[pid]
	props.Add(set)
	b.g.Properties[pid] = props
	return set
}

// clearProperties empties a player's tableau.
func (b *stateBuilder) clearProperties(pid uuid.UUID) {
	for _, set := range b.g.Properties[pid] {
		for _, c := range set.Cards {
			delete(b.g.Cards, c.ID)
		}
		if set.Shack != nil {
			delete(b.g.Cards, set.Shack.ID)
		}
	}
	b.g.Properties[pid] = deal_no_mercy.PropertySets{}
}

// setDeckTop puts fresh cards on top of the deck so the next Draw yields them.
// The deck draws from the front (index 0) per Deck.Draw.
func (b *stateBuilder) setDeckTop(keys ...deal_no_mercy.AssetKey) deal_no_mercy.Cards {
	top := deal_no_mercy.Cards{}
	for _, k := range keys {
		top.Add(b.newCard(k))
	}
	top.Add(b.g.Deck.Cards...)
	b.g.Deck.Cards = top
	return top
}

// setTurn makes the given player the current player with full moves and no
// demands/debts, clearing the go-again flag.
func (b *stateBuilder) setTurn(pid uuid.UUID) {
	b.setTurnKeepDemands(pid)
	b.g.Demands = map[deal_no_mercy.Identifier]deal_no_mercy.Demand{}
}

// setTurnKeepDemands makes the given player current with full moves but leaves
// outstanding demands/debts in place (used to reach mandatory-settlement).
func (b *stateBuilder) setTurnKeepDemands(pid uuid.UUID) {
	for i, p := range b.g.Players {
		if p == pid {
			b.g.CurrPlayerIdx = i
		}
	}
	b.g.MovesLeft = b.g.Config.MovesPerTurn
	if b.g.Demands == nil {
		b.g.Demands = map[deal_no_mercy.Identifier]deal_no_mercy.Demand{}
	}
	b.g.GoAgainQueued = false
}
