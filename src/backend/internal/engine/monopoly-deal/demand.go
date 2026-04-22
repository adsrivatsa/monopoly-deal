package monopoly_deal

import (
	"the-deal/internal/schema/monopoly_deal_schema"

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

type DemandSource int

const (
	DemandSourceUnspecified DemandSource = iota
	DemandSourceRent
	DemandSourceItsMyBirthday
	DemandSourceForcedDeal
	DemandSourceSlyDeal
	DemandSourceDealBreaker
	DemandSourceJustSayNo
	DemansSourceDebtCollector
)

var DemandSourceProtoMap = map[DemandSource]monopoly_deal_schema.DemandSource{
	DemandSourceUnspecified:   monopoly_deal_schema.DemandSource_DEMAND_SOURCE_UNSPECIFIED,
	DemandSourceRent:          monopoly_deal_schema.DemandSource_DEMAND_SOURCE_RENT,
	DemandSourceItsMyBirthday: monopoly_deal_schema.DemandSource_DEMAND_SOURCE_ITS_MY_BIRTHDAY,
	DemandSourceForcedDeal:    monopoly_deal_schema.DemandSource_DEMAND_SOURCE_FORCED_DEAL,
	DemandSourceSlyDeal:       monopoly_deal_schema.DemandSource_DEMAND_SOURCE_SLY_DEAL,
	DemandSourceDealBreaker:   monopoly_deal_schema.DemandSource_DEMAND_SOURCE_DEAL_BREAKER,
	DemandSourceJustSayNo:     monopoly_deal_schema.DemandSource_DEMAND_SOURCE_JUST_SAY_NO,
	DemansSourceDebtCollector: monopoly_deal_schema.DemandSource_DEMAND_SOURCE_DEBT_COLLECTOR,
}

func (ds DemandSource) Proto() monopoly_deal_schema.DemandSource {
	return DemandSourceProtoMap[ds]
}

type Demand struct {
	ID          Identifier        `json:"id" msgpack:"i"`
	Kind        DemandKind        `json:"kind" msgpack:"a"`
	SourceID    Identifier        `json:"source_id" msgpack:"b"`
	TargetID    Identifier        `json:"target_id" msgpack:"c"`
	Payment     PaymentDemand     `json:"payment" msgpack:"d"`
	Property    PropertyDemand    `json:"property" msgpack:"e"`
	PropertySet PropertySetDemand `json:"property_set" msgpack:"f"`
	IsActive    bool              `json:"is_active" msgpack:"g"`
	Source      DemandSource      `json:"source" msgpack:"h"`
}

func NewPaymentDemand(id Identifier, sourceID, targetID Identifier, amount int, source DemandSource) Demand {
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

func NewPaymentDemands(gen *IdentifierGenerator, sourceID Identifier, targetIDs []Identifier, amount int, source DemandSource) map[Identifier]Demand {
	ds := make(map[Identifier]Demand)
	for _, targetID := range targetIDs {
		if sourceID == targetID {
			continue
		}
		id := gen.New()
		ds[id] = NewPaymentDemand(id, sourceID, targetID, amount, source)
	}
	return ds
}

func NewPropertyDemand(id Identifier, sourceID, targetID Identifier, sourceCardID *Identifier, targetCardID Identifier, source DemandSource) Demand {
	return Demand{
		ID:       id,
		Kind:     DemandKindProperty,
		SourceID: sourceID,
		TargetID: targetID,
		Property: PropertyDemand{
			SourceCardID: sourceCardID,
			TargetCardID: targetCardID,
		},
		IsActive: true,
		Source:   source,
	}
}

func NewPropertySetDemand(id Identifier, sourceID, targetID, propertySetID Identifier, source DemandSource) Demand {
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

func (d *Demand) Proto(sourceUUID, targetUUID uuid.UUID) *monopoly_deal_schema.Demand {
	switch d.Kind {
	case DemandKindPayment:
		return &monopoly_deal_schema.Demand{
			Id:         string(d.ID),
			PlayerId:   targetUUID.String(),
			SourceId:   sourceUUID.String(),
			DemandKind: d.Kind.Proto(),
			Demand: &monopoly_deal_schema.Demand_PaymentDemand{
				PaymentDemand: &monopoly_deal_schema.PaymentDemand{
					Amount: int32(d.Payment.Amount),
				},
			},
			IsActive:     d.IsActive,
			DemandSource: d.Source.Proto(),
		}
	case DemandKindProperty:
		var sourceCardID *string
		if d.Property.SourceCardID != nil {
			id := string(*d.Property.SourceCardID)
			sourceCardID = &id
		}

		return &monopoly_deal_schema.Demand{
			Id:         string(d.ID),
			PlayerId:   targetUUID.String(),
			SourceId:   sourceUUID.String(),
			DemandKind: d.Kind.Proto(),
			Demand: &monopoly_deal_schema.Demand_PropertyDemand{
				PropertyDemand: &monopoly_deal_schema.PropertyDemand{
					SourceCardId: sourceCardID,
					TargetCardId: string(d.Property.TargetCardID),
				},
			},
			IsActive:     d.IsActive,
			DemandSource: d.Source.Proto(),
		}
	case DemandKindPropertySet:
		return &monopoly_deal_schema.Demand{
			Id:         string(d.ID),
			PlayerId:   targetUUID.String(),
			SourceId:   sourceUUID.String(),
			DemandKind: d.Kind.Proto(),
			Demand: &monopoly_deal_schema.Demand_PropertySetDemand{
				PropertySetDemand: &monopoly_deal_schema.PropertySetDemand{
					PropertySetId: string(d.PropertySet.PropertySetID),
				},
			},
			IsActive:     d.IsActive,
			DemandSource: d.Source.Proto(),
		}

	default:
		return nil
	}
}

func (d *Demand) Deny(newID Identifier) {
	d.ID = newID

	tmp := d.SourceID
	d.SourceID = d.TargetID
	d.TargetID = tmp

	d.IsActive = !d.IsActive
}
