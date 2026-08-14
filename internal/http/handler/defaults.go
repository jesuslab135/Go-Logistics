package handler

import (
	"context"

	"github.com/shopspring/decimal"

	"fleet/internal/http/middleware"
)

// authorFromContext returns the authenticated employee to credit for a comment
// or upload. Authorship is server-stamped rather than client-supplied; an
// unauthenticated context (id 0) yields NULL — explicitly unattributed instead
// of a false employee reference.
func authorFromContext(ctx context.Context) *int64 {
	id := middleware.EmployeeFromContext(ctx)
	if id == 0 {
		return nil
	}
	return &id
}

// Many columns carry a non-zero default in the schema (issue.state 'OPEN',
// employee.is_active true, quantity 1, ...). A create request that omits such a
// field arrives as Go's zero value, and writing that zero value overwrites the
// default the database would otherwise have applied — an employee created
// without "is_active" was inserted inactive and could never log in.
//
// The helpers below restore the default when a create input was not supplied.
// Booleans need a pointer in the request DTO because false is a meaningful value
// that cannot otherwise be told from "absent"; for the rest, the empty string and
// zero are not values these columns accept, so they can stand in for absent.
//
// The default literals necessarily repeat what the schema declares. Keep them in
// step with internal/db/migrations when a default changes.

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func boolOrDefault(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func int32OrDefault(v, fallback int32) int32 {
	if v == 0 {
		return fallback
	}
	return v
}

func decimalOrDefault(v decimal.Decimal, fallback int64) decimal.Decimal {
	if v.IsZero() {
		return decimal.NewFromInt(fallback)
	}
	return v
}
