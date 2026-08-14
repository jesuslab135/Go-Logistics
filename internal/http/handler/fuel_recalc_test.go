package handler

import (
	"testing"

	"github.com/shopspring/decimal"
)

func dec(t *testing.T, s string) decimal.Decimal {
	t.Helper()
	v, err := decimal.NewFromString(s)
	if err != nil {
		t.Fatalf("bad decimal %q: %v", s, err)
	}
	return v
}

func TestDeriveFuel(t *testing.T) {
	tests := []struct {
		name      string
		prev      *fuelCalcRow
		entry     fuelCalcRow
		wantMiles string
		wantEff   string // "" means null
	}{
		{
			name:      "no previous entry is a baseline",
			prev:      nil,
			entry:     fuelCalcRow{Odometer: dec(t, "1100"), Quantity: dec(t, "10"), FullTank: true},
			wantMiles: "0",
			wantEff:   "",
		},
		{
			name:      "reset restarts the series",
			prev:      &fuelCalcRow{Odometer: dec(t, "1000"), FullTank: true},
			entry:     fuelCalcRow{Odometer: dec(t, "1100"), Quantity: dec(t, "10"), FullTank: true, Reset: true},
			wantMiles: "0",
			wantEff:   "",
		},
		{
			name:      "two full tanks give distance and efficiency",
			prev:      &fuelCalcRow{Odometer: dec(t, "1000"), FullTank: true},
			entry:     fuelCalcRow{Odometer: dec(t, "1100"), Quantity: dec(t, "10"), FullTank: true},
			wantMiles: "100",
			wantEff:   "10",
		},
		{
			name:      "efficiency needs both tanks full",
			prev:      &fuelCalcRow{Odometer: dec(t, "1000"), FullTank: false},
			entry:     fuelCalcRow{Odometer: dec(t, "1100"), Quantity: dec(t, "10"), FullTank: true},
			wantMiles: "100",
			wantEff:   "",
		},
		{
			name:      "partial fill still records distance",
			prev:      &fuelCalcRow{Odometer: dec(t, "1000"), FullTank: true},
			entry:     fuelCalcRow{Odometer: dec(t, "1100"), Quantity: dec(t, "10"), FullTank: false},
			wantMiles: "100",
			wantEff:   "",
		},
		{
			name:      "zero quantity cannot yield efficiency",
			prev:      &fuelCalcRow{Odometer: dec(t, "1000"), FullTank: true},
			entry:     fuelCalcRow{Odometer: dec(t, "1100"), Quantity: decimal.Zero, FullTank: true},
			wantMiles: "100",
			wantEff:   "",
		},
		{
			name:      "odometer rollback yields nothing",
			prev:      &fuelCalcRow{Odometer: dec(t, "1000"), FullTank: true},
			entry:     fuelCalcRow{Odometer: dec(t, "900"), Quantity: dec(t, "10"), FullTank: true},
			wantMiles: "0",
			wantEff:   "",
		},
		{
			name:      "an unchanged odometer yields nothing",
			prev:      &fuelCalcRow{Odometer: dec(t, "1000"), FullTank: true},
			entry:     fuelCalcRow{Odometer: dec(t, "1000"), Quantity: dec(t, "10"), FullTank: true},
			wantMiles: "0",
			wantEff:   "",
		},
		{
			name:      "efficiency is rounded to two places",
			prev:      &fuelCalcRow{Odometer: dec(t, "1000"), FullTank: true},
			entry:     fuelCalcRow{Odometer: dec(t, "1100"), Quantity: dec(t, "3"), FullTank: true},
			wantMiles: "100",
			wantEff:   "33.33",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			miles, eff := deriveFuel(tt.prev, tt.entry)

			if !miles.Equal(dec(t, tt.wantMiles)) {
				t.Errorf("miles = %s, want %s", miles, tt.wantMiles)
			}
			if tt.wantEff == "" {
				if eff != nil {
					t.Errorf("efficiency = %s, want null", eff)
				}
				return
			}
			if eff == nil {
				t.Fatalf("efficiency = null, want %s", tt.wantEff)
			}
			if !eff.Equal(dec(t, tt.wantEff)) {
				t.Errorf("efficiency = %s, want %s", eff, tt.wantEff)
			}
		})
	}
}
