package deal_no_mercy

import (
	"net/http"

	"the-deal/internal/errors"
)

// Errors specific to Deal No Mercy. Shared game errors (turn order, card
// ownership, demand plumbing, ...) are reused from the-deal/internal/errors.
var (
	OutstandingDebtExists = errors.NewError("outstanding debt must be settled first", http.StatusBadRequest, "DNM001")
	DebtDoesNotExist      = errors.NewError("debt does not exist", http.StatusBadRequest, "DNM002")
	NoDebtChipsAvailable  = errors.NewError("no debt chips available", http.StatusBadRequest, "DNM003")
	BankIsEmpty           = errors.NewError("bank is empty", http.StatusBadRequest, "DNM004")
	NoValidTargets        = errors.NewError("no valid targets", http.StatusBadRequest, "DNM005")
	InvalidDistribution   = errors.NewError("invalid distribution", http.StatusBadRequest, "DNM006")
	SetAlreadyHasShack    = errors.NewError("property set already has a shack", http.StatusBadRequest, "DNM007")
	InvalidColorForCard   = errors.NewError("invalid color for card", http.StatusBadRequest, "DNM008")
	InvalidTargetPicks    = errors.NewError("invalid target picks", http.StatusBadRequest, "DNM009")
	InvalidStealCategory  = errors.NewError("invalid steal category", http.StatusBadRequest, "DNM010")
)
