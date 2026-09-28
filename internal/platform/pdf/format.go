// Package pdf renders a report to PDF bytes. It knows layout and formatting;
// it does not know the database or email.
package pdf

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const dateLayout = "02/01/2006"

// dash stands in for a value that does not exist, such as the efficiency of
// an asset with no full-tank interval.
const dash = "—"

// Number formats d with thousands separators and a fixed number of decimals:
// 1234567.5 with 2 places is "1,234,567.50".
func Number(d decimal.Decimal, places int32) string {
	s := d.Abs().StringFixed(places)
	whole, frac, _ := strings.Cut(s, ".")

	var b strings.Builder
	if d.Round(places).Sign() < 0 {
		b.WriteByte('-')
	}
	lead := len(whole) % 3
	if lead > 0 {
		b.WriteString(whole[:lead])
	}
	for i := lead; i < len(whole); i += 3 {
		if b.Len() > 0 && whole[:i] != "" {
			b.WriteByte(',')
		}
		b.WriteString(whole[i : i+3])
	}
	if frac != "" {
		b.WriteByte('.')
		b.WriteString(frac)
	}
	return b.String()
}

// Money formats an amount in the company's currency: "$1,234.50 MXN".
func Money(d decimal.Decimal, currency string) string {
	s := Number(d, 2)
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign, s = "-", s[1:]
	}
	out := sign + "$" + s
	if currency != "" {
		out += " " + currency
	}
	return out
}

// Optional formats a value that may be absent.
func Optional(d *decimal.Decimal, places int32) string {
	if d == nil {
		return dash
	}
	return Number(*d, places)
}

// Percent formats a signed percentage change: "+25.0%", "-3.2%".
func Percent(d *decimal.Decimal) string {
	if d == nil {
		return dash
	}
	sign := ""
	if d.Sign() > 0 {
		sign = "+"
	}
	return sign + Number(*d, 1) + "%"
}

// Date formats t as dd/mm/yyyy in loc.
func Date(t time.Time, loc *time.Location) string {
	if loc == nil {
		loc = time.UTC
	}
	return t.In(loc).Format(dateLayout)
}

// unitLabels translates the unit codes the API stores into what a Spanish
// reader expects. An unknown code is printed as it is stored.
var unitLabels = map[string]string{
	"liters":  "L",
	"gallons": "gal",
	"km":      "km",
	"mi":      "mi",
	"hr":      "h",
}

func Unit(code string) string {
	if label, ok := unitLabels[strings.ToLower(code)]; ok {
		return label
	}
	return code
}
