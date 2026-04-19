package monopoly_deal

import (
	"fun-kames/internal/schema/monopoly_deal_schema"

	"github.com/google/uuid"
)

type PendingRent struct {
	SourceID   Identifier   `json:"source_id" msgpack:"a"`
	TargetIDs  []Identifier `json:"target_ids" msgpack:"b"`
	BaseAmount int          `json:"base_amount" msgpack:"c"`
	Multiplier int          `json:"multiplier" msgpack:"d"`
}

func NewPendingRent(sourceID Identifier, targetIDs []Identifier, baseAmount int) *PendingRent {
	return &PendingRent{
		SourceID:   sourceID,
		TargetIDs:  targetIDs,
		BaseAmount: baseAmount,
		Multiplier: 1,
	}
}

func (pr *PendingRent) Proto(playerUUID uuid.UUID, targetUUIDs []uuid.UUID) *monopoly_deal_schema.PendingRent {
	targetIDs := make([]string, len(targetUUIDs))
	for i, u := range targetUUIDs {
		targetIDs[i] = u.String()
	}

	return &monopoly_deal_schema.PendingRent{
		PlayerId:   playerUUID.String(),
		TargetIds:  targetIDs,
		BaseAmount: int32(pr.BaseAmount),
		Multiplier: int32(pr.Multiplier),
	}
}

func (pr *PendingRent) DoubleRent() {
	pr.Multiplier *= 2
}
