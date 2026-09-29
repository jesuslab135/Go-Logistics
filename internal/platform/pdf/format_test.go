package pdf

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestNumber(t *testing.T) {
	tests := []struct {
		in     string
		places int32
		want   string
	}{
		{"0", 2, "0.00"},
		{"5", 0, "5"},
		{"999.999", 2, "1,000.00"},
		{"1234.5", 2, "1,234.50"},
		{"1234567.891", 2, "1,234,567.89"},
		{"12345678", 0, "12,345,678"},
		{"-1234.5", 2, "-1,234.50"},
		{"-0.001", 2, "0.00"},
		{"100", 1, "100.0"},
	}
	for _, tt := range tests {
		if got := Number(dec(tt.in), tt.places); got != tt.want {
			t.Errorf("Number(%s, %d) = %q, want %q", tt.in, tt.places, got, tt.want)
		}
	}
}

func TestMoney(t *testing.T) {
	if got := Money(dec("25000"), "MXN"); got != "$25,000.00 MXN" {
		t.Errorf("got %q", got)
	}
	if got := Money(dec("-12.5"), "USD"); got != "-$12.50 USD" {
		t.Errorf("got %q", got)
	}
	if got := Money(dec("1"), ""); got != "$1.00" {
		t.Errorf("got %q", got)
	}
}

func TestPercentAndOptional(t *testing.T) {
	up, down, flat := dec("25"), dec("-3.24"), dec("0")
	if got := Percent(&up); got != "+25.0%" {
		t.Errorf("got %q", got)
	}
	if got := Percent(&down); got != "-3.2%" {
		t.Errorf("got %q", got)
	}
	if got := Percent(&flat); got != "0.0%" {
		t.Errorf("got %q", got)
	}
	if got := Percent(nil); got != dash {
		t.Errorf("got %q", got)
	}
	if got := Optional(nil, 2); got != dash {
		t.Errorf("got %q", got)
	}
}

func TestDateUsesTheLocation(t *testing.T) {
	mx, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		t.Fatal(err)
	}
	// 03:00 UTC on the 28th is still the 27th in Mexico City.
	at := time.Date(2026, 9, 28, 3, 0, 0, 0, time.UTC)
	if got := Date(at, mx); got != "27/09/2026" {
		t.Errorf("got %q", got)
	}
	if got := Date(at, nil); got != "28/09/2026" {
		t.Errorf("got %q", got)
	}
}

func TestUnit(t *testing.T) {
	if Unit("liters") != "L" || Unit("GALLONS") != "gal" || Unit("furlongs") != "furlongs" {
		t.Error("unit labels")
	}
}
