package reports

import (
	"time"

	"github.com/shopspring/decimal"
)

// Company is what a report needs to know about the tenant it is for.
type Company struct {
	ID       int64
	Name     string
	Logo     *string
	Timezone string
	Currency string
}

// Header is the part every report shares.
type Header struct {
	Kind        Kind
	Company     Company
	Period      Period
	GeneratedAt time.Time
}

// Comparison sets a figure against the same figure one period earlier.
// Percent is nil when the earlier figure is zero: there is no percentage
// change from nothing.
type Comparison struct {
	Current    decimal.Decimal
	Previous   decimal.Decimal
	Difference decimal.Decimal
	Percent    *decimal.Decimal
}

var hundred = decimal.NewFromInt(100)

func Compare(current, previous decimal.Decimal) Comparison {
	c := Comparison{Current: current, Previous: previous, Difference: current.Sub(previous)}
	if !previous.IsZero() {
		pct := c.Difference.Div(previous).Mul(hundred).Round(1)
		c.Percent = &pct
	}
	return c
}

// UnitTotal is a quantity summed over the rows that share a unit. Units are
// set per asset, so a company total is only meaningful unit by unit.
type UnitTotal struct {
	Label  string
	Unit   string
	Amount decimal.Decimal
}
