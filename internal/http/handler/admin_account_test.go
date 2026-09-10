package handler

import (
	"strings"
	"testing"

	"fleet/internal/http/dto"
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
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want it to name %q", err, tc.want)
			}
		})
	}
}
