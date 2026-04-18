package monopoly_deal

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

type DemandSnapshot struct {
	Kind         DemandKind      `json:"kind" msgpack:"a"`
	Source       Identifier      `json:"source" msgpack:"b"`
	Target       Identifier      `json:"target" msgpack:"c"`
	Amount       int             `json:"amount" msgpack:"d,omitempty"`
	SourceCardID *Identifier     `json:"source_card_id" msgpack:"e,omitempty"`
	TargetCardID Identifier      `json:"target_card_id" msgpack:"f,omitempty"`
	Cards        Cards           `json:"cards" msgpack:"g,omitempty"`
	Original     *DemandSnapshot `json:"original" msgpack:"h,omitempty"`
}

type GameSnapshot struct {
	IDGenerator   IdentifierGenerator           `json:"id_generator" msgpack:"a"`
	IDTranslator  IdentifierTranslator          `json:"id_translator" msgpack:"b"`
	Deck          Deck                          `json:"deck" msgpack:"c"`
	Cards         map[Identifier]Card           `json:"cards" msgpack:"d"`
	Players       []Identifier                  `json:"players" msgpack:"e"`
	CurrPlayerIdx int                           `json:"curr_player_idx" msgpack:"f"`
	MovesLeft     int                           `json:"moves_left" msgpack:"g"`
	Hands         map[Identifier]Cards          `json:"hands" msgpack:"h"`
	Money         map[Identifier]Cards          `json:"money" msgpack:"i"`
	Properties    map[Identifier]PropertySets   `json:"properties" msgpack:"j"`
	Demands       map[Identifier]DemandSnapshot `json:"demands" msgpack:"k"`
	PendingRent   *PendingRent                  `json:"pending_rent" msgpack:"l"`
	Config        Settings                      `json:"config" msgpack:"m"`
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

func cloneCards(cards Cards) Cards {
	if cards == nil {
		return nil
	}

	out := make(Cards, len(cards))
	for i, c := range cards {
		out[i] = c
		out[i].Colors = append([]Color(nil), c.Colors...)
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
	}

	return out
}

func clonePendingRent(pr *PendingRent) *PendingRent {
	if pr == nil {
		return nil
	}

	out := *pr
	out.TargetIDs = append([]Identifier(nil), pr.TargetIDs...)
	return &out
}

func demandToSnapshot(d Demand) (DemandSnapshot, error) {
	switch t := d.(type) {
	case *DeniedDemand:
		if t.Original == nil {
			return DemandSnapshot{}, fmt.Errorf("denied demand missing original")
		}

		original, err := demandToSnapshot(t.Original)
		if err != nil {
			return DemandSnapshot{}, err
		}

		return DemandSnapshot{
			Kind:     DemandKindDenied,
			Source:   t.Source,
			Target:   t.Target,
			Original: &original,
		}, nil
	case *PaymentDemand:
		return DemandSnapshot{
			Kind:   DemandKindPayment,
			Source: t.Source,
			Target: t.Target,
			Amount: t.Amount,
		}, nil
	case *PropertyDemand:
		var sourceCardID *Identifier
		if t.SourceCardID != nil {
			id := *t.SourceCardID
			sourceCardID = &id
		}

		return DemandSnapshot{
			Kind:         DemandKindProperty,
			Source:       t.Source,
			Target:       t.Target,
			SourceCardID: sourceCardID,
			TargetCardID: t.TargetCardID,
		}, nil
	case *PropertySetDemand:
		return DemandSnapshot{
			Kind:   DemandKindPropertySet,
			Source: t.Source,
			Target: t.Target,
			Cards:  cloneCards(t.Cards),
		}, nil
	default:
		return DemandSnapshot{}, fmt.Errorf("unsupported demand type %T", d)
	}
}

func demandFromSnapshot(s DemandSnapshot) (Demand, error) {
	switch s.Kind {
	case DemandKindDenied:
		if s.Original == nil {
			return nil, fmt.Errorf("denied demand snapshot missing original")
		}

		original, err := demandFromSnapshot(*s.Original)
		if err != nil {
			return nil, err
		}

		return &DeniedDemand{
			Kind:     DemandKindDenied,
			Source:   s.Source,
			Target:   s.Target,
			Original: original,
		}, nil
	case DemandKindPayment:
		return &PaymentDemand{
			Kind:   DemandKindPayment,
			Source: s.Source,
			Target: s.Target,
			Amount: s.Amount,
		}, nil
	case DemandKindProperty:
		var sourceCardID *Identifier
		if s.SourceCardID != nil {
			id := *s.SourceCardID
			sourceCardID = &id
		}

		return &PropertyDemand{
			Kind:         DemandKindProperty,
			Source:       s.Source,
			Target:       s.Target,
			SourceCardID: sourceCardID,
			TargetCardID: s.TargetCardID,
		}, nil
	case DemandKindPropertySet:
		return &PropertySetDemand{
			Kind:   DemandKindPropertySet,
			Source: s.Source,
			Target: s.Target,
			Cards:  cloneCards(s.Cards),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported demand kind %d", s.Kind)
	}
}

func (g *Game) Snapshot() (GameSnapshot, error) {
	//uuidToIdentifier := make(map[string]Identifier, len(g.IDTranslator.UUIDToIdentifier))
	//for k, v := range g.IDTranslator.UUIDToIdentifier {
	//	uuidToIdentifier[k.String()] = v
	//}
	//
	//identifierToUUID := make(map[Identifier]string, len(g.IDTranslator.IdentifierToUUID))
	//for k, v := range g.IDTranslator.IdentifierToUUID {
	//	identifierToUUID[k] = v.String()
	//}

	demands := make(map[Identifier]DemandSnapshot, len(g.Demands))
	for id, demand := range g.Demands {
		s, err := demandToSnapshot(demand)
		if err != nil {
			return GameSnapshot{}, err
		}
		demands[id] = s
	}

	cards := make(map[Identifier]Card, len(g.Cards))
	for id, card := range g.Cards {
		copyCard := card
		copyCard.Colors = append([]Color(nil), card.Colors...)
		cards[id] = copyCard
	}

	hands := make(map[Identifier]Cards, len(g.Hands))
	for id, hand := range g.Hands {
		hands[id] = cloneCards(hand)
	}

	money := make(map[Identifier]Cards, len(g.Money))
	for id, pile := range g.Money {
		money[id] = cloneCards(pile)
	}

	properties := make(map[Identifier]PropertySets, len(g.Properties))
	for id, sets := range g.Properties {
		properties[id] = clonePropertySets(sets)
	}

	return GameSnapshot{
		IDGenerator:   *g.IDGenerator,
		IDTranslator:  g.IDTranslator,
		Deck:          Deck{Cards: cloneCards(g.Deck.Cards)},
		Cards:         cards,
		Players:       g.Players,
		CurrPlayerIdx: g.CurrPlayerIdx,
		MovesLeft:     g.MovesLeft,
		Hands:         hands,
		Money:         money,
		Properties:    properties,
		Demands:       demands,
		PendingRent:   clonePendingRent(g.PendingRent),
		Config:        g.Config,
	}, nil
}

func NewGameFromSnapshot(s GameSnapshot) (*Game, error) {
	//uuidToIdentifier := make(map[uuid.UUID]Identifier, len(s.IDTranslator.UUIDToIdentifier))
	//for k, v := range s.IDTranslator.UUIDToIdentifier {
	//	u, err := uuid.Parse(k)
	//	if err != nil {
	//		return nil, fmt.Errorf("invalid uuid_to_identifier key %q: %w", k, err)
	//	}
	//	uuidToIdentifier[u] = v
	//}
	//
	//identifierToUUID := make(map[Identifier]uuid.UUID, len(s.IDTranslator.IdentifierToUUID))
	//for k, v := range s.IDTranslator.IdentifierToUUID {
	//	u, err := uuid.Parse(v)
	//	if err != nil {
	//		return nil, fmt.Errorf("invalid identifier_to_uuid value %q: %w", v, err)
	//	}
	//	identifierToUUID[k] = u
	//}

	demands := make(map[Identifier]Demand, len(s.Demands))
	for id, ds := range s.Demands {
		d, err := demandFromSnapshot(ds)
		if err != nil {
			return nil, err
		}
		demands[id] = d
	}

	cards := make(map[Identifier]Card, len(s.Cards))
	for id, card := range s.Cards {
		copyCard := card
		copyCard.Colors = append([]Color(nil), card.Colors...)
		cards[id] = copyCard
	}

	hands := make(map[Identifier]Cards, len(s.Hands))
	for id, hand := range s.Hands {
		hands[id] = cloneCards(hand)
	}

	money := make(map[Identifier]Cards, len(s.Money))
	for id, pile := range s.Money {
		money[id] = cloneCards(pile)
	}

	properties := make(map[Identifier]PropertySets, len(s.Properties))
	for id, sets := range s.Properties {
		properties[id] = clonePropertySets(sets)
	}

	idGenerator := s.IDGenerator
	return &Game{
		IDGenerator:   &idGenerator,
		IDTranslator:  s.IDTranslator,
		Deck:          Deck{Cards: cloneCards(s.Deck.Cards)},
		Cards:         cards,
		Players:       s.Players,
		CurrPlayerIdx: s.CurrPlayerIdx,
		MovesLeft:     s.MovesLeft,
		Hands:         hands,
		Money:         money,
		Properties:    properties,
		Demands:       demands,
		PendingRent:   clonePendingRent(s.PendingRent),
		Config:        s.Config,
	}, nil
}
