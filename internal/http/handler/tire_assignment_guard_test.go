package handler

import (
	"reflect"
	"testing"

	"fleet/internal/http/dto"
)

// Approving a tire assignment is gated on tire_approvals/approve, a permission
// deliberately separate from the four method-derived actions. If the ordinary
// update DTO carries the resolution fields, any caller who can reach the CRUD
// route can forge an approval with a PUT and attribute it to another employee,
// which is exactly the gate the separate permission exists to enforce.
//
// state, approved_by_id and resolved_at are owned by ResolveTireAssignmentRequest,
// which guards on state = 'PENDING' and stamps the resolver.
func TestTireAssignmentUpdateCannotCarryResolutionFields(t *testing.T) {
	forbidden := []string{"State", "ApprovedByID", "ResolvedAt"}
	typ := reflect.TypeOf(dto.UpdateTireAssignmentRequestRequest{})
	for _, name := range forbidden {
		if _, ok := typ.FieldByName(name); ok {
			t.Errorf("UpdateTireAssignmentRequestRequest still carries %s; "+
				"resolution belongs to POST /tire-assignment-requests/{id}/approve", name)
		}
	}
}

// The fields a caller may legitimately edit must survive, or this fix would
// have quietly turned the update route into a no-op.
func TestTireAssignmentUpdateKeepsItsEditableFields(t *testing.T) {
	editable := []string{"TireID", "VehicleID", "PositionCode", "RequestedByID", "RequestedAt", "Notes"}
	typ := reflect.TypeOf(dto.UpdateTireAssignmentRequestRequest{})
	for _, name := range editable {
		if _, ok := typ.FieldByName(name); !ok {
			t.Errorf("UpdateTireAssignmentRequestRequest lost %s, which is an ordinary editable field", name)
		}
	}
}
