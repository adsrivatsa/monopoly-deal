package monopoly_deal

import (
	"the-deal/internal/schema/monopoly_deal_schema"

	"github.com/google/uuid"
)

type PendingRent struct {
	SourceID   uuid.UUID   `json:"source_id" msgpack:"a"`
	TargetIDs  []uuid.UUID `json:"target_ids" msgpack:"b"`
	BaseAmount int         `json:"base_amount" msgpack:"c"`
	Multiplier int         `json:"multiplier" msgpack:"d"`
}

func NewPendingRent(sourceID uuid.UUID, targetIDs []uuid.UUID, baseAmount int) PendingRent {
	return PendingRent{
		SourceID:   sourceID,
		TargetIDs:  targetIDs,
		BaseAmount: baseAmount,
		Multiplier: 1,
	}
}

func (pr *PendingRent) Proto() *monopoly_deal_schema.PendingRent {
	targetIDs := make([]string, len(pr.TargetIDs))
	for i, id := range pr.TargetIDs {
		targetIDs[i] = id.String()
	}

	return &monopoly_deal_schema.PendingRent{
		PlayerId:   pr.SourceID.String(),
		TargetIds:  targetIDs,
		BaseAmount: int32(pr.BaseAmount),
		Multiplier: int32(pr.Multiplier),
	}
}

func (pr *PendingRent) DoubleRent() {
	pr.Multiplier *= 2
}
