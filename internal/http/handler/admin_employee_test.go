package handler

import (
	"errors"
	"reflect"
	"slices"
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

// The audit records what changed, not what was requested. A full replace sends
// the employee's whole membership set every time, so recording the request
// verbatim would file a "granted" row on every save for companies the employee
// already had — and bury the one grant that actually mattered.
func TestMembershipDelta(t *testing.T) {
	tests := []struct {
		name   string
		before []int64
		after  []int64
		want   []membershipChange
	}{
		{
			name:   "an unchanged set records nothing",
			before: []int64{3, 8},
			after:  []int64{3, 8},
			want:   nil,
		},
		{
			name:   "an addition is recorded once",
			before: []int64{3},
			after:  []int64{3, 8},
			want:   []membershipChange{{company: 8, action: membershipGranted}},
		},
		{
			name:   "a removal is recorded once",
			before: []int64{3, 8},
			after:  []int64{3},
			want:   []membershipChange{{company: 8, action: membershipRevoked}},
		},
		{
			name:   "a swap records both halves",
			before: []int64{3},
			after:  []int64{8},
			want: []membershipChange{
				{company: 8, action: membershipGranted},
				{company: 3, action: membershipRevoked},
			},
		},
		{
			name:   "a first grant to an employee with nothing",
			before: nil,
			after:  []int64{3},
			want:   []membershipChange{{company: 3, action: membershipGranted}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := membershipDelta(tt.before, tt.after)
			if !slices.Equal(got, tt.want) {
				t.Errorf("membershipDelta(%v, %v) = %v, want %v", tt.before, tt.after, got, tt.want)
			}
		})
	}
}
