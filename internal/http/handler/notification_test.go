package handler

import "testing"

// url is an unconstrained text column, so a producer is the only thing standing
// between the bell and an off-site link that looks like part of the app. That
// makes this a phishing check, not a formatting one.
func TestValidNotificationURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"an internal path", "/app/maintenance/work-orders/42", true},
		{"the root", "/", true},
		{"empty, for a notification that points nowhere", "", true},
		{"an absolute http url", "http://example.com/phish", false},
		{"an absolute https url", "https://example.com/phish", false},
		{"a protocol-relative url, which a browser treats as absolute", "//example.com/phish", false},
		{"a path that does not start at the root", "app/work-orders/42", false},
		{"a javascript scheme", "javascript:alert(1)", false},
		{"a path with a query string, which the bell does not render", "/app/x?next=//evil.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidNotificationURL(tt.url); got != tt.want {
				t.Errorf("ValidNotificationURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
