package reports

import (
	"testing"
	"time"
)

func mustLoc(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	return loc
}

func day(loc *time.Location, y int, m time.Month, d, h int) time.Time {
	return time.Date(y, m, d, h, 0, 0, 0, loc)
}

func TestLatestDueWeekly(t *testing.T) {
	mx := mustLoc(t, "America/Mexico_City")
	tests := []struct {
		name      string
		now       time.Time
		wantStart time.Time
	}{
		// 2026-09-28 is a Monday.
		{"monday before the send hour still owes the week before last", day(mx, 2026, 9, 28, 5), day(mx, 2026, 9, 14, 0)},
		{"monday at the send hour owes last week", day(mx, 2026, 9, 28, 6), day(mx, 2026, 9, 21, 0)},
		{"sunday night owes the week that ended a week ago", day(mx, 2026, 9, 27, 23), day(mx, 2026, 9, 14, 0)},
		{"midweek owes last week", day(mx, 2026, 9, 30, 12), day(mx, 2026, 9, 21, 0)},
		{"year boundary", day(mx, 2027, 1, 4, 7), day(mx, 2026, 12, 28, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LatestDue(FuelWeekly, tt.now, mx, 6)
			if !got.Start.Equal(tt.wantStart) {
				t.Fatalf("start = %s, want %s", got.Start, tt.wantStart)
			}
			if want := tt.wantStart.AddDate(0, 0, 7); !got.End.Equal(want) {
				t.Fatalf("end = %s, want %s", got.End, want)
			}
		})
	}
}

func TestLatestDueMonthly(t *testing.T) {
	mx := mustLoc(t, "America/Mexico_City")
	tests := []struct {
		name      string
		now       time.Time
		wantStart time.Time
		wantEnd   time.Time
	}{
		{"the 1st before the send hour", day(mx, 2026, 10, 1, 5), day(mx, 2026, 8, 1, 0), day(mx, 2026, 9, 1, 0)},
		{"the 1st at the send hour", day(mx, 2026, 10, 1, 6), day(mx, 2026, 9, 1, 0), day(mx, 2026, 10, 1, 0)},
		{"january owes december", day(mx, 2027, 1, 15, 12), day(mx, 2026, 12, 1, 0), day(mx, 2027, 1, 1, 0)},
		{"march owes a 28-day february", day(mx, 2027, 3, 31, 12), day(mx, 2027, 2, 1, 0), day(mx, 2027, 3, 1, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LatestDue(MaintenanceMonthly, tt.now, mx, 6)
			if !got.Start.Equal(tt.wantStart) || !got.End.Equal(tt.wantEnd) {
				t.Fatalf("got %s..%s, want %s..%s", got.Start, got.End, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

// The same instant is a different local day in different timezones, so two
// companies can owe different periods at the same moment.
func TestLatestDueUsesTheCompanyTimezone(t *testing.T) {
	mx := mustLoc(t, "America/Mexico_City")
	tokyo := mustLoc(t, "Asia/Tokyo")
	now := time.Date(2026, 9, 27, 22, 0, 0, 0, time.UTC) // Sunday 16:00 in Mexico, Monday 07:00 in Tokyo

	if got, want := LatestDue(FuelWeekly, now, mx, 6).Start, day(mx, 2026, 9, 14, 0); !got.Equal(want) {
		t.Fatalf("mexico start = %s, want %s", got, want)
	}
	if got, want := LatestDue(FuelWeekly, now, tokyo, 6).Start, day(tokyo, 2026, 9, 21, 0); !got.Equal(want) {
		t.Fatalf("tokyo start = %s, want %s", got, want)
	}
}

// A week containing a daylight saving change is not 168 hours long. Both ends
// must still be local midnight.
func TestWeekAcrossDaylightSavingChange(t *testing.T) {
	ny := mustLoc(t, "America/New_York")
	// US clocks go back on Sunday 2026-11-01.
	p := LatestDue(FuelWeekly, day(ny, 2026, 11, 2, 8), ny, 6)
	if want := day(ny, 2026, 10, 26, 0); !p.Start.Equal(want) {
		t.Fatalf("start = %s, want %s", p.Start, want)
	}
	if want := day(ny, 2026, 11, 2, 0); !p.End.Equal(want) {
		t.Fatalf("end = %s, want %s", p.End, want)
	}
	if got := p.End.Sub(p.Start); got != 169*time.Hour {
		t.Fatalf("week length = %s, want 169h", got)
	}
	prev := p.Previous(FuelWeekly)
	if h, m, s := prev.Start.Clock(); h != 0 || m != 0 || s != 0 {
		t.Fatalf("previous start is not local midnight: %s", prev.Start)
	}
}

func TestStartDateKeepsTheLocalDay(t *testing.T) {
	tokyo := mustLoc(t, "Asia/Tokyo")
	p := WeekOf(day(tokyo, 2026, 9, 23, 12), tokyo)
	// Monday 00:00 in Tokyo is still Sunday in UTC; the date column must say Monday.
	if got, want := p.StartDate(), time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("start date = %s, want %s", got, want)
	}
	if got, want := p.LastDay().Format("2006-01-02"), "2026-09-27"; got != want {
		t.Fatalf("last day = %s, want %s", got, want)
	}
}

func TestLoadLocationFallsBack(t *testing.T) {
	for _, name := range []string{"", "Mars/Olympus_Mons", "  "} {
		loc, err := LoadLocation(name)
		if err == nil {
			t.Fatalf("%q: expected the fallback to be reported", name)
		}
		if loc.String() != DefaultTimezone {
			t.Fatalf("%q: location = %s, want %s", name, loc, DefaultTimezone)
		}
	}
	loc, err := LoadLocation("America/Tijuana")
	if err != nil || loc.String() != "America/Tijuana" {
		t.Fatalf("valid name: loc = %v, err = %v", loc, err)
	}
}
