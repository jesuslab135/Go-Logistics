package handler

import (
	"context"

	"fleet/internal/http/middleware"
)

// Who an audited change is attributed to. Two values cover both producers a
// transition can have; a third (an integration, say) can be added without a
// migration, because the column is a varchar rather than an enum.
//
// The distinction is carried explicitly rather than left to a null actor id: a
// null could mean "the system did it" or "we did not record who", and an audit
// trail that cannot tell those apart is not one.
const (
	actorEmployee = "employee"
	actorSystem   = "system"
)

// actorOf attributes a change to the authenticated caller, or to the system
// when there is none.
func actorOf(ctx context.Context) (*int64, string) {
	id := middleware.EmployeeFromContext(ctx)
	if id == 0 {
		return nil, actorSystem
	}
	return &id, actorEmployee
}
