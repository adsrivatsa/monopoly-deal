package monopoly_deal

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

func (pr *PendingRent) DoubleRent() {
	pr.Multiplier *= 2
}
