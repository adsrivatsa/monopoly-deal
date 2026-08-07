package deal_no_mercy

import (
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

type GameSnapshot struct {
	Config        Settings                   `json:"config" msgpack:"a"`
	IDGenerator   IdentifierGenerator        `json:"id_generator" msgpack:"b"`
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

func (g *Game) EncodeMsgpack() ([]byte, error) {
	snapshot, err := g.Snapshot()
	if err != nil {
		return nil, err
	}

	return msgpack.Marshal(snapshot)
}

func DecodeMsgpack(data []byte) (*Game, error) {
	var snapshot GameSnapshot
	if err := msgpack.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}

	return NewGameFromSnapshot(snapshot)
}

func cloneCard(c Card) Card {
	out := c
	out.Colors = append([]Color(nil), c.Colors...)
	return out
}

func cloneCards(cards Cards) Cards {
	if cards == nil {
		return nil
	}

	out := make(Cards, len(cards))
	for i, c := range cards {
		out[i] = cloneCard(c)
	}

	return out
}

func clonePropertySets(sets PropertySets) PropertySets {
	if sets == nil {
		return nil
	}

	out := make(PropertySets, len(sets))
	for i, s := range sets {
		out[i] = s
		out[i].Cards = cloneCards(s.Cards)
		if s.Shack != nil {
			shack := cloneCard(*s.Shack)
			out[i].Shack = &shack
		}
	}

	return out
}

func cloneDebts(debts []DebtObligation) []DebtObligation {
	out := make([]DebtObligation, len(debts))
	copy(out, debts)
	return out
}

func cloneDebtChips(chips map[uuid.UUID]int) map[uuid.UUID]int {
	out := make(map[uuid.UUID]int, len(chips))
	for id, n := range chips {
		out[id] = n
	}
	return out
}

func (g *Game) Snapshot() (GameSnapshot, error) {
	cards := make(map[Identifier]Card, len(g.Cards))
	for id, card := range g.Cards {
		cards[id] = cloneCard(card)
	}

	hands := make(map[uuid.UUID]Cards, len(g.Hands))
	for id, hand := range g.Hands {
		hands[id] = cloneCards(hand)
	}

	money := make(map[uuid.UUID]Cards, len(g.Money))
	for id, pile := range g.Money {
		money[id] = cloneCards(pile)
	}

	properties := make(map[uuid.UUID]PropertySets, len(g.Properties))
	for id, sets := range g.Properties {
		properties[id] = clonePropertySets(sets)
	}

	return GameSnapshot{
		Config:        g.Config,
		IDGenerator:   *g.IDGenerator,
		Deck:          Deck{Cards: cloneCards(g.Deck.Cards)},
		Cards:         cards,
		Players:       g.Players,
		CurrPlayerIdx: g.CurrPlayerIdx,
		MovesLeft:     g.MovesLeft,
		Hands:         hands,
		Money:         money,
		Properties:    properties,
		Demands:       g.Demands,
		DebtChips:     cloneDebtChips(g.DebtChips),
		Debts:         cloneDebts(g.Debts),
		GoAgainQueued: g.GoAgainQueued,
		LastAction:    g.LastAction,
		SequenceNum:   g.SequenceNum,
	}, nil
}

func NewGameFromSnapshot(s GameSnapshot) (*Game, error) {
	cards := make(map[Identifier]Card, len(s.Cards))
	for id, card := range s.Cards {
		cards[id] = cloneCard(card)
	}

	hands := make(map[uuid.UUID]Cards, len(s.Hands))
	for id, hand := range s.Hands {
		hands[id] = cloneCards(hand)
	}

	money := make(map[uuid.UUID]Cards, len(s.Money))
	for id, pile := range s.Money {
		money[id] = cloneCards(pile)
	}

	properties := make(map[uuid.UUID]PropertySets, len(s.Properties))
	for id, sets := range s.Properties {
		properties[id] = clonePropertySets(sets)
	}

	demands := s.Demands
	if demands == nil {
		demands = make(map[Identifier]Demand)
	}

	debts := s.Debts
	if debts == nil {
		debts = []DebtObligation{}
	}

	idGenerator := s.IDGenerator
	return &Game{
		Config:        s.Config,
		IDGenerator:   &idGenerator,
		Deck:          Deck{Cards: cloneCards(s.Deck.Cards)},
		Cards:         cards,
		Players:       s.Players,
		CurrPlayerIdx: s.CurrPlayerIdx,
		MovesLeft:     s.MovesLeft,
		Hands:         hands,
		Money:         money,
		Properties:    properties,
		Demands:       demands,
		DebtChips:     cloneDebtChips(s.DebtChips),
		Debts:         cloneDebts(debts),
		GoAgainQueued: s.GoAgainQueued,
		LastAction:    s.LastAction,
		SequenceNum:   s.SequenceNum,
	}, nil
}
