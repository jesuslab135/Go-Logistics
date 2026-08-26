package dto

import "github.com/shopspring/decimal"

// TotalOverrideRequest replaces a document's computed total with an explicit
// figure, or clears an existing override.
//
// The override exists for the figure the formula cannot express — a negotiated
// total, a rounding agreed with a vendor, a correction to a historical document.
// It is a distinct action rather than a writable field so that setting one is a
// decision somebody made, and it is recorded with who made it and why.
type TotalOverrideRequest struct {
	// Amount is the figure to charge. Null clears the override and returns the
	// document to its computed total.
	Amount *decimal.Decimal `json:"amount"`
	// Reason is required whenever Amount is given: an unexplained override is a
	// number nobody can defend later.
	Reason string `json:"reason" binding:"omitempty,max=2000"`
}
