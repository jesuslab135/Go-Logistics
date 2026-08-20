package handler

import (
	"errors"
	"reflect"
	"testing"

	"fleet/internal/platform/apierr"
)

func TestNormalizeCompanyIDs(t *testing.T) {
	tests := []struct {
		name string
		in   []int64
		want []int64
	}{
		{name: "sorts and de-duplicates", in: []int64{9, 3, 9, 1}, want: []int64{1, 3, 9}},
		{name: "already normal is unchanged", in: []int64{1, 2}, want: []int64{1, 2}},
		{name: "nil becomes an empty slice, never nil", in: nil, want: []int64{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCompanyIDs(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// The replace is a full overwrite of employee_companies, so a default that is
// not in the new set would leave the employee pointing at a company the same
// request just removed them from.
func TestValidateMembershipReplace(t *testing.T) {
	tests := []struct {
		name      string
		companies []int64
		defaultID *int64
		wantErr   bool
		wantField string
	}{
		{name: "default inside the set is allowed", companies: []int64{3, 7}, defaultID: ptr(7), wantErr: false},
		{name: "null default is allowed", companies: []int64{3, 7}, defaultID: nil, wantErr: false},
		{name: "default outside the set is rejected", companies: []int64{3, 7}, defaultID: ptr(8), wantErr: true, wantField: "default_company_id"},
		{name: "an empty set is rejected", companies: []int64{}, defaultID: nil, wantErr: true, wantField: "company_ids"},
		{name: "a non-positive company id is rejected", companies: []int64{0, 3}, defaultID: nil, wantErr: true, wantField: "company_ids"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMembershipReplace(tt.companies, tt.defaultID)
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
			if !ok || details[tt.wantField] == "" {
				t.Errorf("details = %#v, want a %s entry", ae.Details, tt.wantField)
			}
		})
	}
}

// The admin namespace is gated on RequireAdminRole alone, and role_id is a
// single global FK: granting yourself a company you do not belong to therefore
// makes you an admin of that company too, with DELETE /companies/:id behind it.
// The mirror image is just as bad — a full replace that drops a company evicts
// the employee from a tenant the caller has no access to. So the constraint is
// on the delta, not the whole set: an id the request leaves untouched needs no
// entitlement, an id it adds or removes does.
func TestValidateGrantableCompanies(t *testing.T) {
	tests := []struct {
		name            string
		requested       []int64
		current         []int64
		callerCompanies []int64
		wantErr         bool
	}{
		{name: "an unchanged set is allowed even where the caller belongs to nothing in it", requested: []int64{3, 8}, current: []int64{3, 8}, callerCompanies: []int64{3}, wantErr: false},
		{name: "adding a company the caller does not belong to is refused", requested: []int64{3, 8}, current: []int64{3}, callerCompanies: []int64{3, 7}, wantErr: true},
		{name: "removing a company the caller does not belong to is refused", requested: []int64{3}, current: []int64{3, 8}, callerCompanies: []int64{3, 7}, wantErr: true},
		{name: "adding a company the caller belongs to is allowed", requested: []int64{3, 7}, current: []int64{3}, callerCompanies: []int64{3, 7, 9}, wantErr: false},
		{name: "removing a company the caller belongs to is allowed", requested: []int64{3}, current: []int64{3, 7}, callerCompanies: []int64{3, 7}, wantErr: false},
		{name: "a caller with no companies cannot change anything", requested: []int64{3}, current: nil, callerCompanies: nil, wantErr: true},
		{name: "an empty caller slice cannot change anything", requested: []int64{3, 7}, current: []int64{3}, callerCompanies: []int64{}, wantErr: true},
		{name: "a first grant into the caller's own company is allowed", requested: []int64{3}, current: nil, callerCompanies: []int64{3}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGrantableCompanies(tt.requested, tt.current, tt.callerCompanies)
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
			if ae.Status != 403 {
				t.Errorf("status = %d, want 403", ae.Status)
			}
		})
	}
}
