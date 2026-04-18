package monopoly_deal

import (
	"fun-kames/internal/errors"
	"slices"
)

type DemandKind int

const (
	DemandKindDenied DemandKind = iota
	DemandKindPayment
	DemandKindProperty
	DemandKindPropertySet
)

type Demand interface {
	GetKind() DemandKind
	GetSource() Identifier
	GetTarget() Identifier
	Deny(id Identifier) (Demand, error)
	IsCompliant(id Identifier, cards ...Card) error
}

type DeniedDemand struct {
	Kind     DemandKind `msgpack:"kind"`
	Source   Identifier `msgpack:"source"`
	Target   Identifier `msgpack:"target"`
	Original Demand     `msgpack:"original"`
}

func NewDeniedDemand(source Identifier, target Identifier, original Demand) Demand {
	return &DeniedDemand{
		Kind:     DemandKindDenied,
		Source:   source,
		Target:   target,
		Original: original,
	}
}

func (d *DeniedDemand) GetKind() DemandKind {
	return d.Kind
}

func (d *DeniedDemand) GetSource() Identifier {
	return d.Source
}

func (d *DeniedDemand) GetTarget() Identifier {
	return d.Target
}

func (d *DeniedDemand) Deny(id Identifier) (Demand, error) {
	if id != d.Target {
		return nil, errors.CannotRespondToDemand
	}

	return d.Original, nil
}

func (d *DeniedDemand) IsCompliant(id Identifier, cards ...Card) error {
	if d.Target != id {
		return errors.CannotRespondToDemand
	}

	// nothing to comply with

	return nil
}

type PaymentDemand struct {
	Kind   DemandKind `msgpack:"kind"`
	Source Identifier `msgpack:"source"`
	Target Identifier `msgpack:"target"`
	Amount int        `msgpack:"amount"`
}

func NewPaymentDemand(source Identifier, target Identifier, amount int) Demand {
	return &PaymentDemand{
		Kind:   DemandKindPayment,
		Source: source,
		Target: target,
		Amount: amount,
	}
}

func NewPaymentDemands(source Identifier, target []Identifier, amount int) map[Identifier]Demand {
	demands := make(map[Identifier]Demand)
	for _, id := range target {
		if source == id {
			continue
		}
		demands[id] = NewPaymentDemand(source, id, amount)
	}
	return demands
}

func (pd *PaymentDemand) GetKind() DemandKind {
	return pd.Kind
}

func (pd *PaymentDemand) GetSource() Identifier {
	return pd.Source
}

func (pd *PaymentDemand) GetTarget() Identifier {
	return pd.Target
}

func (pd *PaymentDemand) Deny(id Identifier) (Demand, error) {
	if id != pd.Target {
		return nil, errors.CannotRespondToDemand
	}

	return NewDeniedDemand(pd.Target, pd.Source, pd), nil
}

func (pd *PaymentDemand) IsCompliant(id Identifier, cards ...Card) error {
	if pd.Target != id {
		return errors.CannotRespondToDemand
	}

	var value int
	for _, card := range cards {
		value += card.Value
	}

	if value < pd.Amount {
		return errors.PaymentDoesNotCoverAmount
	}

	return nil
}

type PropertyDemand struct {
	Kind         DemandKind  `msgpack:"kind"`
	Source       Identifier  `msgpack:"source"`
	Target       Identifier  `msgpack:"target"`
	SourceCardID *Identifier `msgpack:"source_card"`
	TargetCardID Identifier  `msgpack:"target_card"`
}

func NewPropertyDemand(source Identifier, target Identifier, sourceCardID *Identifier, targetCardID Identifier) Demand {
	return &PropertyDemand{
		Kind:         DemandKindProperty,
		Source:       source,
		Target:       target,
		SourceCardID: sourceCardID,
		TargetCardID: targetCardID,
	}
}

func (pd *PropertyDemand) GetKind() DemandKind {
	return pd.Kind
}

func (pd *PropertyDemand) GetSource() Identifier {
	return pd.Source
}

func (pd *PropertyDemand) GetTarget() Identifier {
	return pd.Target
}

func (pd *PropertyDemand) Deny(id Identifier) (Demand, error) {
	if id != pd.Target {
		return nil, errors.CannotRespondToDemand
	}

	return NewDeniedDemand(pd.Target, pd.Source, pd), nil
}

func (pd *PropertyDemand) IsCompliant(id Identifier, cards ...Card) error {
	if pd.Target != id {
		return errors.CannotRespondToDemand
	}

	if len(cards) != 1 {
		return errors.InvalidAmountOfCards
	}

	card := cards[0]

	if card.ID != pd.TargetCardID {
		return errors.InvalidCardForAction
	}

	return nil
}

type PropertySetDemand struct {
	Kind   DemandKind `msgpack:"kind"`
	Source Identifier `msgpack:"source"`
	Target Identifier `msgpack:"target"`
	Cards  Cards      `msgpack:"cards"`
}

func NewPropertySetDemand(source Identifier, target Identifier, cards ...Card) Demand {
	return &PropertySetDemand{
		Kind:   DemandKindPropertySet,
		Source: source,
		Target: target,
		Cards:  cards,
	}
}

func (psd *PropertySetDemand) GetKind() DemandKind {
	return psd.Kind
}

func (psd *PropertySetDemand) GetSource() Identifier {
	return psd.Source
}

func (psd *PropertySetDemand) GetTarget() Identifier {
	return psd.Target
}

func (psd *PropertySetDemand) Deny(id Identifier) (Demand, error) {
	if id != psd.Target {
		return nil, errors.CannotRespondToDemand
	}

	return NewDeniedDemand(psd.Target, psd.Source, psd), nil
}

func (psd *PropertySetDemand) IsCompliant(id Identifier, cards ...Card) error {
	if psd.Target != id {
		return errors.CannotRespondToDemand
	}

	if len(cards) != len(psd.Cards) {
		return errors.InvalidAmountOfCards
	}

	expected := make([]Identifier, len(psd.Cards))
	for i, card := range psd.Cards {
		expected[i] = card.ID
	}

	provided := make([]Identifier, len(cards))
	for i, card := range cards {
		provided[i] = card.ID
	}

	slices.Sort(expected)
	slices.Sort(provided)

	if !slices.Equal(expected, provided) {
		return errors.InvalidCardForAction
	}

	return nil
}
