package handler

import "testing"

func ptr(v int64) *int64 { return &v }

// A login token is only safe if its company_id claim names a company the
// employee is actually a member of: every downstream query trusts that claim.
// With no membership the answer is now "no company" rather than "no token" —
// an account owner who has not created a company yet must be able to log in.
func TestResolveLoginCompany(t *testing.T) {
	tests := []struct {
		name        string
		defaultID   *int64
		memberships []int64
		want        *int64
	}{
		{"default company that is a real membership wins", ptr(7), []int64{3, 7, 9}, ptr(7)},
		{"nil default falls back to the lowest membership", nil, []int64{9, 3, 7}, ptr(3)},
		{"default naming a company the employee does not belong to is ignored", ptr(42), []int64{9, 3}, ptr(3)},
		{"no membership yields no company rather than a failure", ptr(7), nil, nil},
		{"no membership and no default yields no company", nil, nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveLoginCompany(tc.defaultID, tc.memberships)
			switch {
			case tc.want == nil && got != nil:
				t.Fatalf("got %d, want no company", *got)
			case tc.want != nil && got == nil:
				t.Fatalf("got no company, want %d", *tc.want)
			case tc.want != nil && *got != *tc.want:
				t.Fatalf("got %d, want %d", *got, *tc.want)
			}
		})
	}
}

// mayLogIn is the line that decides who may log in at all. resolveLoginCompany
// cannot see account_id, so a company-less employee with an account and one
// with neither are indistinguishable to it — both resolve to a nil company.
// This predicate is where that split actually happens, so it is exhaustively
// tested: two bools, four combinations.
func TestMayLogIn(t *testing.T) {
	tests := []struct {
		name      string
		companyID *int64
		accountID *int64
		want      bool
	}{
		{"scoped session with a known account may log in", ptr(7), ptr(1), true},
		{"scoped session with no account claim may log in", ptr(7), nil, true},
		{"company-less owner with an account may log in — the whole point of this change", nil, ptr(1), true},
		{"an employee belonging to nothing must never receive a token", nil, nil, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := mayLogIn(tc.companyID, tc.accountID); got != tc.want {
				t.Fatalf("mayLogIn(%v, %v) = %v, want %v", tc.companyID, tc.accountID, got, tc.want)
			}
		})
	}
}
