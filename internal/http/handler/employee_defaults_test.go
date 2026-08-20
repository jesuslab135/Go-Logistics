package handler

import (
	"errors"
	"testing"

	"fleet/internal/platform/apierr"
)

// default_company_id is what login scopes a session to, so accepting one the
// employee has no employee_companies row for would hand them a token for a
// tenant they are not a member of.
func TestValidateDefaultCompany(t *testing.T) {
	tests := []struct {
		name        string
		defaultID   *int64
		memberships []int64
		wantErr     bool
	}{
		{name: "null is always allowed", defaultID: nil, memberships: []int64{4}, wantErr: false},
		{name: "a real membership is allowed", defaultID: ptr(4), memberships: []int64{4, 9}, wantErr: false},
		{name: "a company the employee does not belong to is rejected", defaultID: ptr(8), memberships: []int64{4, 9}, wantErr: true},
		{name: "any value is rejected when there are no memberships", defaultID: ptr(4), memberships: nil, wantErr: true},
		{name: "zero is rejected", defaultID: ptr(0), memberships: []int64{4}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDefaultCompany(tt.defaultID, tt.memberships)
			if tt.wantErr == (err == nil) {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				return
			}
			var ae *apierr.Error
			if !errors.As(err, &ae) {
				t.Fatalf("error is %T, want *apierr.Error", err)
			}
			if ae.Status != 422 {
				t.Errorf("status = %d, want 422", ae.Status)
			}
			details, ok := ae.Details.(map[string]string)
			if !ok || details["default_company_id"] == "" {
				t.Errorf("details = %#v, want a default_company_id entry", ae.Details)
			}
		})
	}
}
