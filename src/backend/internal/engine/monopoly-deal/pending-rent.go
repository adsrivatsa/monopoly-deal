package monopoly_deal

type PendingRent struct {
	SourceID   Identifier   `msgpack:"source_id"`
	TargetIDs  []Identifier `msgpack:"target_ids"`
	BaseAmount int          `msgpack:"base_amount"`
	Multiplier int          `msgpack:"multiplier"`
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
