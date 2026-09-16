package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"slices"
	"testing"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
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

// The cross-tenant employee row must say which client a person belongs to and
// whether they are platform staff, flattened next to the normal employee fields.
func TestToAdminEmployeeResponsesCarriesAccountAndPlatformFlag(t *testing.T) {
	account := int64(9)
	rows := []gen.Employee{
		{ID: 1, Email: "ops@platform.test", IsPlatformAdmin: true},
		{ID: 2, Email: "owner@client.test", AccountID: &account},
	}
	base := []dto.EmployeeResponse{
		{ID: 1, Email: "ops@platform.test"},
		{ID: 2, Email: "owner@client.test", IsAccountOwner: true},
	}

	got := toAdminEmployeeResponses(rows, base)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].AccountID != nil || !got[0].IsPlatformAdmin {
		t.Errorf("platform staff row = %+v, want account_id nil and is_platform_admin true", got[0])
	}
	if got[1].AccountID == nil || *got[1].AccountID != 9 || got[1].IsPlatformAdmin || !got[1].IsAccountOwner {
		t.Errorf("client row = %+v, want account_id 9, not platform staff, owner flag kept", got[1])
	}

	body, err := json.Marshal(got[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var flat map[string]any
	if err := json.Unmarshal(body, &flat); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v, present := flat["account_id"]; !present || v != nil {
		t.Errorf("account_id must be present and null for platform staff, body %s", body)
	}
	if flat["email"] != "ops@platform.test" || flat["is_platform_admin"] != true {
		t.Errorf("fields not flattened, body %s", body)
	}
}

func TestValidateMembershipAccount(t *testing.T) {
	account := int64(4)
	tests := []struct {
		name      string
		accountID *int64
		inAccount int64
		requested int
		wantErr   bool
	}{
		{"every company in the employee's account", &account, 2, 2, false},
		{"one company from another account", &account, 1, 2, true},
		{"platform staff have no account", nil, 0, 1, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMembershipAccount(tc.accountID, tc.inAccount, tc.requested)
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
				return
			}
			var apiErr *apierr.Error
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *apierr.Error, got %T (%v)", err, err)
			}
			if apiErr.Status != http.StatusUnprocessableEntity || apiErr.Code != "validation_failed" {
				t.Fatalf("status/code = %d/%q, want 422/validation_failed", apiErr.Status, apiErr.Code)
			}
			details, ok := apiErr.Details.(map[string]string)
			if !ok || details["company_ids"] == "" {
				t.Fatalf("details = %#v, want company_ids named", apiErr.Details)
			}
		})
	}
}
