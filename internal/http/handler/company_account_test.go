package handler

import (
	"context"
	"testing"

	"fleet/internal/http/middleware"
)

// The account a company lands in comes from the caller's identity, never from
// the request body. Honouring a body-supplied account would let an owner plant
// a company inside another client.
func TestAccountForNewCompanyComesFromIdentity(t *testing.T) {
	ctx := middleware.ContextWithIdentity(context.Background(), middleware.Identity{
		EmployeeID: 5, AccountID: ptr(3), IsAccountOwner: true, IsActive: true,
	})
	got := middleware.AccountFromContext(ctx)
	if got == nil || *got != 3 {
		t.Fatalf("account = %v, want 3", got)
	}
}

// Platform staff belong to no account, so they have no account to create a
// company into. The handler must refuse rather than insert a NULL.
func TestPlatformStaffHaveNoAccountForCompanyCreation(t *testing.T) {
	ctx := middleware.ContextWithIdentity(context.Background(), middleware.Identity{
		EmployeeID: 1, AccountID: nil, IsPlatformAdmin: true, IsActive: true,
	})
	if middleware.AccountFromContext(ctx) != nil {
		t.Fatal("platform staff must resolve to no account")
	}
}
