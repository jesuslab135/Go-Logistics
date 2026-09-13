package handler

import (
	"testing"

	"github.com/shopspring/decimal"
)

// Available stock is a physical count. A negative adjustment may take it to
// zero but never below: inventory that has gone negative is not a smaller
// number, it is a valuation, a reorder point and a low-stock dashboard all
// reporting on a quantity that cannot exist.
//
// The check runs while applyAdjustment already holds the row lock, so it is
// decided against the same row the write will touch.
func TestCheckStockFloor(t *testing.T) {
	for _, tt := range []struct {
		name       string
		stock      string
		adjustment string
		wantErr    bool
	}{
		{"decrease within stock", "5", "-3", false},
		{"decrease to exactly zero", "5", "-5", false},
		{"decrease one past zero", "5", "-6", true},
		{"decrease far past zero", "5", "-1000", true},
		{"any increase", "0", "1000", false},
		{"zero adjustment on zero stock", "0", "0", false},
		{"fractional within stock", "2.50", "-2.50", false},
		{"fractional past zero", "2.50", "-2.51", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := checkStockFloor(
				decimal.RequireFromString(tt.stock),
				decimal.RequireFromString(tt.adjustment),
			)
			if tt.wantErr && err == nil {
				t.Errorf("stock %s adjusted by %s: want refusal, got nil", tt.stock, tt.adjustment)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("stock %s adjusted by %s: want nil, got %v", tt.stock, tt.adjustment, err)
			}
		})
	}
}
