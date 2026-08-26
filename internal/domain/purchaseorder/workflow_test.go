package purchaseorder

import (
	"slices"
	"testing"
)

// The machine decides what a purchase order may do next, and getting it wrong
// means either committing money without approval or stranding an order in a
// state nothing can move it out of. Every edge is pinned.
func TestTransitions(t *testing.T) {
	tests := []struct {
		name   string
		action string
		from   string
		want   bool
	}{
		{"a draft is submitted for approval", ActionSubmit, StateDraft, true},
		{"a rejected order is resubmitted directly", ActionSubmit, StateRejected, true},
		{"an approved order is not submitted again", ActionSubmit, StateApproved, false},

		{"a pending order is approved", ActionApprove, StatePendingApproval, true},
		{"a draft cannot be approved without being submitted", ActionApprove, StateDraft, false},
		{"an approved order is not approved twice", ActionApprove, StateApproved, false},

		{"a pending order is rejected", ActionReject, StatePendingApproval, true},
		{"an approved order cannot be rejected after the fact", ActionReject, StateApproved, false},

		// Rejection returns to draft rather than ending the order: procurement
		// rejection is revise-and-resubmit, and a terminal REJECTED would force a
		// new order and lose the line items.
		{"a rejected order is revised back to draft", ActionRevise, StateRejected, true},
		{"a draft is already a draft", ActionRevise, StateDraft, false},

		{"an approved order is purchased", ActionPurchase, StateApproved, true},
		{"a pending order cannot be purchased", ActionPurchase, StatePendingApproval, false},

		{"a purchased order is partially received", ActionReceivePartial, StatePurchased, true},
		{"a purchased order is fully received", ActionReceiveFull, StatePurchased, true},
		{"a partially received order is then fully received", ActionReceiveFull, StateReceivedPartial, true},
		{"a partially received order is not partially received again", ActionReceivePartial, StateReceivedPartial, false},

		{"a fully received order is closed", ActionClose, StateReceivedFull, true},
		{"a partially received order is not closed", ActionClose, StateReceivedPartial, false},
		{"a closed order is terminal", ActionSubmit, StateClosed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transition, ok := Lookup(tt.action)
			if !ok {
				t.Fatalf("no transition named %q", tt.action)
			}
			if got := transition.Allows(tt.from); got != tt.want {
				t.Errorf("%s from %s = %v, want %v", tt.action, tt.from, got, tt.want)
			}
		})
	}
}

// Approval is the privileged act: it commits money. Submitting an order for
// that decision, and receiving the goods afterwards, are ordinary work. If this
// ever inverts, either everyone can approve or nobody can receive.
func TestOnlyApprovalDecisionsAreGated(t *testing.T) {
	gated := map[string]bool{ActionApprove: true, ActionReject: true}

	for _, action := range Actions() {
		transition, _ := Lookup(action)
		if transition.RequiresApproval != gated[action] {
			t.Errorf("%s: RequiresApproval = %v, want %v", action, transition.RequiresApproval, gated[action])
		}
	}
}

// A rejection with no reason sends the buyer back to guess what to change.
func TestRejectionRequiresAReason(t *testing.T) {
	reject, _ := Lookup(ActionReject)
	if !reject.RequiresReason {
		t.Error("reject must require a reason")
	}

	for _, action := range Actions() {
		if action == ActionReject {
			continue
		}
		transition, _ := Lookup(action)
		if transition.RequiresReason {
			t.Errorf("%s requires a reason; only reject should", action)
		}
	}
}

// Every state must be reachable and, except the terminal one, leavable —
// otherwise an order can be parked somewhere nothing moves it out of.
func TestNoStateIsStranded(t *testing.T) {
	states := []string{
		StateDraft, StatePendingApproval, StateRejected, StateApproved,
		StatePurchased, StateReceivedPartial, StateReceivedFull, StateClosed,
	}

	var reachable, leavable []string
	for _, action := range Actions() {
		transition, _ := Lookup(action)
		reachable = append(reachable, transition.To)
		leavable = append(leavable, transition.From...)
	}

	for _, state := range states {
		if state != StateDraft && !slices.Contains(reachable, state) {
			t.Errorf("%s cannot be reached by any transition", state)
		}
		if state != StateClosed && !slices.Contains(leavable, state) {
			t.Errorf("%s cannot be left by any transition", state)
		}
	}
}

func TestLookupRejectsAnUnknownAction(t *testing.T) {
	if _, ok := Lookup("cancel"); ok {
		t.Error("Lookup accepted an action the machine does not define")
	}
}
