// Package purchaseorder holds the purchase-order state machine.
//
// It lives in domain rather than in the handler because the legal transitions
// are a business rule, not an HTTP concern: the same table has to be readable
// by anything that later moves an order (a receiving integration, a scheduled
// close), and a second copy of it would be a second answer.
package purchaseorder

import "slices"

// The states a purchase order moves through. These are the values already
// stored in purchase_order.state and already validated as a list filter, so
// this is a description of the existing data, not a new vocabulary.
const (
	StateDraft           = "DRAFT"
	StatePendingApproval = "PENDING_APPROVAL"
	StateRejected        = "REJECTED"
	StateApproved        = "APPROVED"
	StatePurchased       = "PURCHASED"
	StateReceivedPartial = "RECEIVED_PARTIAL"
	StateReceivedFull    = "RECEIVED_FULL"
	StateClosed          = "CLOSED"
)

// The named transitions. Each is one route, and each stamps its own timestamp.
const (
	ActionSubmit         = "submit"
	ActionApprove        = "approve"
	ActionReject         = "reject"
	ActionRevise         = "revise"
	ActionPurchase       = "purchase"
	ActionReceivePartial = "receive-partial"
	ActionReceiveFull    = "receive-full"
	ActionClose          = "close"
)

// Transition describes one legal move.
type Transition struct {
	Action string
	From   []string
	To     string
	// RequiresApproval marks the transitions gated on the purchase_orders.approve
	// permission rather than ordinary update. Deciding whether money may be
	// committed is the privileged act; submitting an order for that decision, and
	// receiving the goods afterwards, are ordinary work.
	RequiresApproval bool
	// RequiresReason marks a transition that must say why. Rejection without a
	// reason sends the buyer back to guess what to change.
	RequiresReason bool
}

// transitions is the whole machine. Rejection returns to DRAFT via an explicit
// revise rather than being terminal, because procurement rejection is
// revise-and-resubmit; a terminal REJECTED would force a new order and lose the
// line items.
var transitions = []Transition{
	{Action: ActionSubmit, From: []string{StateDraft, StateRejected}, To: StatePendingApproval},
	{Action: ActionApprove, From: []string{StatePendingApproval}, To: StateApproved, RequiresApproval: true},
	{Action: ActionReject, From: []string{StatePendingApproval}, To: StateRejected, RequiresApproval: true, RequiresReason: true},
	{Action: ActionRevise, From: []string{StateRejected}, To: StateDraft},
	{Action: ActionPurchase, From: []string{StateApproved}, To: StatePurchased},
	{Action: ActionReceivePartial, From: []string{StatePurchased}, To: StateReceivedPartial},
	{Action: ActionReceiveFull, From: []string{StatePurchased, StateReceivedPartial}, To: StateReceivedFull},
	{Action: ActionClose, From: []string{StateReceivedFull}, To: StateClosed},
}

// Lookup returns the transition for an action.
func Lookup(action string) (Transition, bool) {
	for _, t := range transitions {
		if t.Action == action {
			return t, true
		}
	}
	return Transition{}, false
}

// Allows reports whether the transition may run from the given state.
func (t Transition) Allows(state string) bool {
	return slices.Contains(t.From, state)
}

// Actions lists every action name, for route registration and error messages.
func Actions() []string {
	out := make([]string, 0, len(transitions))
	for _, t := range transitions {
		out = append(out, t.Action)
	}
	return out
}
