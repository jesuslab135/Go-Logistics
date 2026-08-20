package handler

import "testing"

func ptr(v int64) *int64 { return &v }

// A login token is only safe if its company_id claim names a company the
// employee is actually a member of: every downstream query trusts that claim.
func TestResolveLoginCompany(t *testing.T) {
	tests := []struct {
		name        string
		defaultID   *int64
		memberships []int64
		want        int64
		wantOK      bool
	}{
		{
			name:        "default company that is a real membership wins",
			defaultID:   ptr(7),
			memberships: []int64{3, 7, 9},
			want:        7,
			wantOK:      true,
		},
		{
			name:        "nil default falls back to the lowest membership",
			defaultID:   nil,
			memberships: []int64{9, 3, 7},
			want:        3,
			wantOK:      true,
		},
		{
			name:        "default naming a company the employee does not belong to is ignored",
			defaultID:   ptr(42),
			memberships: []int64{9, 3},
			want:        3,
			wantOK:      true,
		},
		{
			name:        "no membership at all refuses the login",
			defaultID:   ptr(7),
			memberships: nil,
			want:        0,
			wantOK:      false,
		},
		{
			name:        "no membership and no default refuses the login",
			defaultID:   nil,
			memberships: []int64{},
			want:        0,
			wantOK:      false,
		},
		{
			name:        "a zero default is never trusted",
			defaultID:   ptr(0),
			memberships: []int64{5},
			want:        5,
			wantOK:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resolveLoginCompany(tt.defaultID, tt.memberships)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("company = %d, want %d", got, tt.want)
			}
		})
	}
}
