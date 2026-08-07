package deal_no_mercy

import (
	"github.com/google/uuid"
)

// PaymentDemand: target must pay Amount (rent, yoink). If they cannot pay in
// full they surrender everything payable and a debt chip is issued.
type PaymentDemand struct {
	Amount int `json:"amount" msgpack:"a"`
}

// PropertyDemand: attacker-chosen property card the target must hand over
// (market crash, one demand per opponent).
type PropertyDemand struct {
	TargetCardID Identifier `json:"target_card_id" msgpack:"a"`
}

// PropertySetDemand: complete set the target must hand over (set snatcher).
type PropertySetDemand struct {
	PropertySetID Identifier `json:"property_set_id" msgpack:"a"`
}

// ColorPropertiesDemand: target must hand over all their sets of Color
// (property raid, one demand per opponent).
type ColorPropertiesDemand struct {
	Color Color `json:"color" msgpack:"a"`
}

// BankCardDemand: attacker-chosen bank card the target must hand over
// (heist, one demand per opponent).
type BankCardDemand struct {
	TargetCardID Identifier `json:"target_card_id" msgpack:"a"`
}

// PickpocketDemand: target must hand over every hand card of Category.
type PickpocketDemand struct {
	Category StealCategory `json:"category" msgpack:"a"`
}

type DemandKind int

const (
	DemandKindUnspecified DemandKind = iota
	DemandKindPayment
	DemandKindProperty
	DemandKindPropertySet
	DemandKindColorProperties
	DemandKindBankCard
	DemandKindBankSwap
	DemandKindDebtTrap
	DemandKindRepoMan
	DemandKindTaxDay
	DemandKindPickpocket
)

type DemandSource int

const (
	DemandSourceUnspecified DemandSource = iota
	DemandSourceRent
	DemandSourceYoink
	DemandSourceSetSnatcher
	DemandSourceMarketCrash
	DemandSourcePropertyRaid
	DemandSourceHeist
	DemandSourceBankSwap
	DemandSourceDebtTrap
	DemandSourceRepoMan
	DemandSourceTaxDay
	DemandSourcePickpocket
	DemandSourceNah
)

// Demand mirrors the classic engine's demand: TargetID is who must respond.
// A NAH! denial re-issues the demand under a new ID with source/target
// swapped and IsActive flipped; complying with an inactive demand simply
// dismisses it (the denial stood).
type Demand struct {
	ID              Identifier            `json:"id" msgpack:"i"`
	Kind            DemandKind            `json:"kind" msgpack:"a"`
	SourceID        uuid.UUID             `json:"source_id" msgpack:"b"`
	TargetID        uuid.UUID             `json:"target_id" msgpack:"c"`
	Payment         PaymentDemand         `json:"payment" msgpack:"d"`
	Property        PropertyDemand        `json:"property" msgpack:"e"`
	PropertySet     PropertySetDemand     `json:"property_set" msgpack:"f"`
	ColorProperties ColorPropertiesDemand `json:"color_properties" msgpack:"j"`
	BankCard        BankCardDemand        `json:"bank_card" msgpack:"k"`
	Pickpocket      PickpocketDemand      `json:"pickpocket" msgpack:"l"`
	IsActive        bool                  `json:"is_active" msgpack:"g"`
	Source          DemandSource          `json:"source" msgpack:"h"`
}

func NewPaymentDemand(id Identifier, sourceID, targetID uuid.UUID, amount int, source DemandSource) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindPayment,
		SourceID: sourceID,
		TargetID: targetID,
		Payment: PaymentDemand{
			Amount: amount,
		},
		IsActive: true,
		Source:   source,
	}
}

func NewPropertyDemand(id Identifier, sourceID, targetID uuid.UUID, targetCardID Identifier, source DemandSource) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindProperty,
		SourceID: sourceID,
		TargetID: targetID,
		Property: PropertyDemand{
			TargetCardID: targetCardID,
		},
		IsActive: true,
		Source:   source,
	}
}

func NewPropertySetDemand(id Identifier, sourceID, targetID uuid.UUID, propertySetID Identifier, source DemandSource) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindPropertySet,
		SourceID: sourceID,
		TargetID: targetID,
		PropertySet: PropertySetDemand{
			PropertySetID: propertySetID,
		},
		IsActive: true,
		Source:   source,
	}
}

func NewColorPropertiesDemand(id Identifier, sourceID, targetID uuid.UUID, color Color, source DemandSource) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindColorProperties,
		SourceID: sourceID,
		TargetID: targetID,
		ColorProperties: ColorPropertiesDemand{
			Color: color,
		},
		IsActive: true,
		Source:   source,
	}
}

func NewBankCardDemand(id Identifier, sourceID, targetID uuid.UUID, targetCardID Identifier, source DemandSource) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindBankCard,
		SourceID: sourceID,
		TargetID: targetID,
		BankCard: BankCardDemand{
			TargetCardID: targetCardID,
		},
		IsActive: true,
		Source:   source,
	}
}

func NewBankSwapDemand(id Identifier, sourceID, targetID uuid.UUID) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindBankSwap,
		SourceID: sourceID,
		TargetID: targetID,
		IsActive: true,
		Source:   DemandSourceBankSwap,
	}
}

func NewDebtTrapDemand(id Identifier, sourceID, targetID uuid.UUID) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindDebtTrap,
		SourceID: sourceID,
		TargetID: targetID,
		IsActive: true,
		Source:   DemandSourceDebtTrap,
	}
}

func NewRepoManDemand(id Identifier, sourceID, targetID uuid.UUID) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindRepoMan,
		SourceID: sourceID,
		TargetID: targetID,
		IsActive: true,
		Source:   DemandSourceRepoMan,
	}
}

func NewTaxDayDemand(id Identifier, sourceID, targetID uuid.UUID) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindTaxDay,
		SourceID: sourceID,
		TargetID: targetID,
		IsActive: true,
		Source:   DemandSourceTaxDay,
	}
}

func NewPickpocketDemand(id Identifier, sourceID, targetID uuid.UUID, category StealCategory) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindPickpocket,
		SourceID: sourceID,
		TargetID: targetID,
		Pickpocket: PickpocketDemand{
			Category: category,
		},
		IsActive: true,
		Source:   DemandSourcePickpocket,
	}
}

func (d *Demand) Deny(newID Identifier) {
	d.ID = newID

	tmp := d.SourceID
	d.SourceID = d.TargetID
	d.TargetID = tmp

	d.IsActive = !d.IsActive
}
