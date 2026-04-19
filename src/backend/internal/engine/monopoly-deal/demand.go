package monopoly_deal

import (
	"fun-kames/internal/schema/monopoly_deal_schema"

	"github.com/google/uuid"
)

type PaymentDemand struct {
	Amount int `json:"amount" msgpack:"a"`
}

type PropertyDemand struct {
	SourceCardID *Identifier `json:"source_card_id" msgpack:"a"`
	TargetCardID Identifier  `json:"target_card_id" msgpack:"b"`
}

type PropertySetDemand struct {
	PropertySetID Identifier `json:"property_set_id" msgpack:"a"`
}

type DemandKind int

const (
	DemandKindUnspecified DemandKind = iota
	DemandKindPayment
	DemandKindProperty
	DemandKindPropertySet
)

var DemandKindProtoMap = map[DemandKind]monopoly_deal_schema.DemandKind{
	DemandKindUnspecified: monopoly_deal_schema.DemandKind_DEMAND_KIND_UNSPECIFIED,
	DemandKindPayment:     monopoly_deal_schema.DemandKind_DEMAND_KIND_PAYMENT,
	DemandKindProperty:    monopoly_deal_schema.DemandKind_DEMAND_KIND_PROPERTY,
	DemandKindPropertySet: monopoly_deal_schema.DemandKind_DEMAND_KIND_PROPERTY_SET,
}

func (dk DemandKind) Proto() monopoly_deal_schema.DemandKind {
	return DemandKindProtoMap[dk]
}

type Demand struct {
	Kind        DemandKind        `json:"kind" msgpack:"a"`
	SourceID    Identifier        `json:"source_id" msgpack:"b"`
	TargetID    Identifier        `json:"target_id" msgpack:"c"`
	Payment     PaymentDemand     `json:"payment" msgpack:"d"`
	Property    PropertyDemand    `json:"property" msgpack:"e"`
	PropertySet PropertySetDemand `json:"property_set" msgpack:"f"`
	IsActive    bool              `json:"is_active" msgpack:"g"`
}

func NewPaymentDemand(sourceID, targetID Identifier, amount int) Demand {
	return Demand{
		Kind:     DemandKindPayment,
		SourceID: sourceID,
		TargetID: targetID,
		Payment: PaymentDemand{
			Amount: amount,
		},
		IsActive: true,
	}
}

func NewPaymentDemands(sourceID Identifier, targetIDs []Identifier, amount int) map[Identifier]Demand {
	ds := make(map[Identifier]Demand)
	for _, targetID := range targetIDs {
		ds[targetID] = NewPaymentDemand(sourceID, targetID, amount)
	}
	return ds
}

func NewPropertyDemand(sourceID, targetID Identifier, sourceCardID *Identifier, targetCardID Identifier) Demand {
	return Demand{
		Kind:     DemandKindProperty,
		SourceID: sourceID,
		TargetID: targetID,
		Property: PropertyDemand{
			SourceCardID: sourceCardID,
			TargetCardID: targetCardID,
		},
		IsActive: true,
	}
}

func NewPropertySetDemand(sourceID, targetID, propertySetID Identifier) Demand {
	return Demand{
		Kind:     DemandKindPropertySet,
		SourceID: sourceID,
		TargetID: targetID,
		PropertySet: PropertySetDemand{
			PropertySetID: propertySetID,
		},
		IsActive: true,
	}
}

func (d *Demand) Proto(sourceUUID uuid.UUID) *monopoly_deal_schema.Demand {
	switch d.Kind {
	case DemandKindPayment:
		return &monopoly_deal_schema.Demand{
			SourceId:   sourceUUID.String(),
			DemandKind: d.Kind.Proto(),
			Demand: &monopoly_deal_schema.Demand_PaymentDemand{
				PaymentDemand: &monopoly_deal_schema.PaymentDemand{
					Amount: int32(d.Payment.Amount),
				},
			},
			IsActive: d.IsActive,
		}
	case DemandKindProperty:
		var sourceCardID *string
		if d.Property.SourceCardID != nil {
			id := string(*d.Property.SourceCardID)
			sourceCardID = &id
		}

		return &monopoly_deal_schema.Demand{
			SourceId:   sourceUUID.String(),
			DemandKind: d.Kind.Proto(),
			Demand: &monopoly_deal_schema.Demand_PropertyDemand{
				PropertyDemand: &monopoly_deal_schema.PropertyDemand{
					SourceCardId: sourceCardID,
					TargetCardId: string(d.Property.TargetCardID),
				},
			},
			IsActive: d.IsActive,
		}
	case DemandKindPropertySet:
		return &monopoly_deal_schema.Demand{
			SourceId:   sourceUUID.String(),
			DemandKind: d.Kind.Proto(),
			Demand: &monopoly_deal_schema.Demand_PropertySetDemand{
				PropertySetDemand: &monopoly_deal_schema.PropertySetDemand{
					PropertySetId: string(d.PropertySet.PropertySetID),
				},
			},
			IsActive: d.IsActive,
		}

	default:
		return nil
	}
}

func (d *Demand) Deny() {
	tmp := d.SourceID
	d.SourceID = d.TargetID
	d.TargetID = tmp

	d.IsActive = !d.IsActive
}
