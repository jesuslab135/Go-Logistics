package handler

import (
	"errors"
	"net/http"
	"testing"

	"fleet/internal/http/dto"
	"fleet/internal/platform/apierr"
)

// Provisioning is one call because the two-call alternative can leave an account
// whose owner exists but cannot log in. Every field the owner needs must
// therefore be present and validated up front.
func TestValidateCreateAccountRequest(t *testing.T) {
	tests := []struct {
		name string
		in   dto.CreateAccountRequest
		want string
	}{
		{"complete request passes", dto.CreateAccountRequest{
			Name: "Acme", OwnerFirstName: "Ada", OwnerLastName: "Byron",
			OwnerEmail: "ada@acme.test", OwnerPassword: "correct-horse-battery",
		}, ""},
		{"missing account name", dto.CreateAccountRequest{
			OwnerFirstName: "Ada", OwnerLastName: "Byron",
			OwnerEmail: "ada@acme.test", OwnerPassword: "correct-horse-battery",
		}, "name"},
		{"missing owner email", dto.CreateAccountRequest{
			Name: "Acme", OwnerFirstName: "Ada", OwnerLastName: "Byron",
			OwnerPassword: "correct-horse-battery",
		}, "owner_email"},
		{"short password is refused", dto.CreateAccountRequest{
			Name: "Acme", OwnerFirstName: "Ada", OwnerLastName: "Byron",
			OwnerEmail: "ada@acme.test", OwnerPassword: "short",
		}, "owner_password"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCreateAccount(tc.in)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("expected valid, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected an error naming %q, got nil", tc.want)
			}

			// apierr.Validation always sets Message to the fixed string
			// "validation failed" -- every other validation error in this API
			// (admin_company.go, admin_employee.go) relies on that constant
			// message, with the offending fields living only in Details. This
			// asserts against the contract a client actually sees: status,
			// code, and the per-field details map -- not the Go Error()
			// string, which never carries the field name.
			var apiErr *apierr.Error
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected an *apierr.Error, got %T", err)
			}
			if apiErr.Status != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want %d", apiErr.Status, http.StatusUnprocessableEntity)
			}
			if apiErr.Code != "validation_failed" {
				t.Fatalf("code = %q, want %q", apiErr.Code, "validation_failed")
			}
			details, ok := apiErr.Details.(map[string]string)
			if !ok {
				t.Fatalf("details = %T, want map[string]string", apiErr.Details)
			}
			if _, named := details[tc.want]; !named {
				t.Fatalf("details %v does not name the field %q", details, tc.want)
			}
		})
	}
}
