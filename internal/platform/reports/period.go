package reports

import (
	"fmt"
	"strings"
	"time"
)

// DefaultTimezone matches the column default of company.timezone.
const DefaultTimezone = "America/Mexico_City"

// Period is the half-open interval [Start, End) a report covers, in the
// company's own timezone.
type Period struct {
	Start time.Time
	End   time.Time
}

// LoadLocation resolves a company's timezone. An empty or unknown name falls
// back to DefaultTimezone and reports the problem, so one mistyped company
// still gets its report instead of being skipped.
func LoadLocation(name string) (*time.Location, error) {
	name = strings.TrimSpace(name)
	if name != "" {
		if loc, err := time.LoadLocation(name); err == nil {
			return loc, nil
		}
	}
	loc, err := time.LoadLocation(DefaultTimezone)
	if err != nil {
		return time.UTC, fmt.Errorf("reports: timezone %q and the default are both unavailable: %w", name, err)
	}
	return loc, fmt.Errorf("reports: unknown timezone %q, using %s", name, DefaultTimezone)
}

// WeekOf returns the Monday-to-Monday week that contains t.
func WeekOf(t time.Time, loc *time.Location) Period {
	t = t.In(loc)
	// time.Weekday counts from Sunday; shift so Monday is 0.
	back := (int(t.Weekday()) + 6) % 7
	start := time.Date(t.Year(), t.Month(), t.Day()-back, 0, 0, 0, 0, loc)
	return Period{Start: start, End: start.AddDate(0, 0, 7)}
}

// MonthOf returns the calendar month that contains t.
func MonthOf(t time.Time, loc *time.Location) Period {
	t = t.In(loc)
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
	return Period{Start: start, End: start.AddDate(0, 1, 0)}
}

// PeriodOf returns the period of the given kind that contains t.
func PeriodOf(kind Kind, t time.Time, loc *time.Location) Period {
	if kind == MaintenanceMonthly {
		return MonthOf(t, loc)
	}
	return WeekOf(t, loc)
}

// Previous returns the period immediately before p.
//
// It steps by calendar date, not by duration: a week that contains a daylight
// saving change is 167 or 169 hours long, and subtracting 168 would land an
// hour off midnight.
func (p Period) Previous(kind Kind) Period {
	if kind == MaintenanceMonthly {
		return Period{Start: p.Start.AddDate(0, -1, 0), End: p.Start}
	}
	return Period{Start: p.Start.AddDate(0, 0, -7), End: p.Start}
}

// StartDate is the period's first local day as a UTC midnight, the form the
// report_run.period_start date column is written and compared in.
func (p Period) StartDate() time.Time {
	return time.Date(p.Start.Year(), p.Start.Month(), p.Start.Day(), 0, 0, 0, 0, time.UTC)
}

// LastDay is the last local day inside the period, for display.
func (p Period) LastDay() time.Time {
	return p.End.AddDate(0, 0, -1)
}

// LatestDue returns the most recent period whose report is due at now: the
// period before the current one once sendHour has passed on the current
// period's first day, and the one before that until then.
func LatestDue(kind Kind, now time.Time, loc *time.Location, sendHour int) Period {
	current := PeriodOf(kind, now, loc)
	s := current.Start
	due := time.Date(s.Year(), s.Month(), s.Day(), sendHour, 0, 0, 0, loc)
	last := current.Previous(kind)
	if now.Before(due) {
		return last.Previous(kind)
	}
	return last
}
