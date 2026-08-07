package deal_no_mercy

import (
	"github.com/google/uuid"
)

// DebtObligation is one outstanding debt chip: DebtorID handed one of their
// chips to CreditorID (failed payment or DEBT TRAP) and must settle it with
// any one hand card at the start of their next turn, consuming one play.
// While a chip is outstanding it is not in the debtor's available pool
// (Game.DebtChips); settling returns it.
type DebtObligation struct {
	ID         Identifier `json:"id" msgpack:"a"`
	DebtorID   uuid.UUID  `json:"debtor_id" msgpack:"b"`
	CreditorID uuid.UUID  `json:"creditor_id" msgpack:"c"`
}
