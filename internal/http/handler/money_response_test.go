package handler

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"fleet/internal/db/gen"
	"fleet/internal/domain/money"
)

// moneyDec parses a decimal literal for these tests. Named to avoid colliding
// with the existing dec(t, s) helper in fuel_recalc_test.go.
func moneyDec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestPurchaseOrderResponseExposesComputedBreakdown(t *testing.T) {
	// subtotal 1000, 10% discount => discount 100, net 900, tax1 16% => 144,
	// tax2 5% => 45, shipping 50 => total 1139.
	row := gen.PurchaseOrder{
		Subtotal:           moneyDec("1000"),
		DiscountType:       money.Percentage,
		Discount:           moneyDec("0"),
		DiscountPercentage: moneyDec("10"),
		Tax1Type:           money.Percentage,
		Tax1Percentage:     moneyDec("16"),
		Tax2Type:           money.Percentage,
		Tax2Percentage:     moneyDec("5"),
		Shipping:           moneyDec("50"),
		TotalAmount:        moneyDec("1139"),
	}
	got := toPurchaseOrderResponse(row)
	if !got.DiscountAmount.Equal(moneyDec("100")) {
		t.Errorf("discount_amount = %s, want 100", got.DiscountAmount)
	}
	if !got.Net.Equal(moneyDec("900")) {
		t.Errorf("net = %s, want 900", got.Net)
	}
	if !got.Tax1Amount.Equal(moneyDec("144")) {
		t.Errorf("tax_1_amount = %s, want 144", got.Tax1Amount)
	}
	if !got.Tax2Amount.Equal(moneyDec("45")) {
		t.Errorf("tax_2_amount = %s, want 45", got.Tax2Amount)
	}
}

func TestPurchaseOrderResponsePassesThroughOverride(t *testing.T) {
	amt, reason, by, at := moneyDec("2000"), "negotiated", int64(42), time.Now().UTC()
	row := gen.PurchaseOrder{
		Subtotal: moneyDec("1000"), TotalAmount: moneyDec("2000"),
		DiscountType: money.Percentage, Tax1Type: money.Percentage, Tax2Type: money.Percentage,
		TotalOverride: &amt, TotalOverrideReason: &reason, TotalOverrideByID: &by, TotalOverrideAt: &at,
	}
	got := toPurchaseOrderResponse(row)
	if got.TotalOverride == nil || !got.TotalOverride.Equal(amt) {
		t.Errorf("total_override = %v, want %s", got.TotalOverride, amt)
	}
	if got.TotalOverrideReason == nil || *got.TotalOverrideReason != reason {
		t.Errorf("total_override_reason = %v, want %q", got.TotalOverrideReason, reason)
	}
	if got.TotalOverrideByID == nil || *got.TotalOverrideByID != by {
		t.Errorf("total_override_by_id = %v, want %d", got.TotalOverrideByID, by)
	}
	if got.TotalOverrideAt == nil {
		t.Error("total_override_at is nil, want a timestamp")
	}
}

func TestWorkOrderResponseExposesComputedBreakdown(t *testing.T) {
	// parts 500 + 10% markup => 550, labor 300 + 20 fixed markup => 320,
	// subtotal 870, 10% discount => 87, net 783, tax1 16% => 125.28,
	// tax2 5% => 39.15.
	row := gen.WorkOrder{
		PartsSubtotal:         moneyDec("500"),
		LaborSubtotal:         moneyDec("300"),
		PartsMarkupType:       money.Percentage,
		PartsMarkupPercentage: moneyDec("10"),
		LaborMarkupType:       money.Fixed,
		LaborMarkup:           moneyDec("20"),
		DiscountType:          money.Percentage,
		DiscountPercentage:    moneyDec("10"),
		Tax1Type:              money.Percentage,
		Tax1Percentage:        moneyDec("16"),
		Tax2Type:              money.Percentage,
		Tax2Percentage:        moneyDec("5"),
	}
	got := toWorkOrderResponse(row)
	if !got.DiscountPercentage.Equal(moneyDec("10")) {
		t.Errorf("discount_percentage = %s, want 10", got.DiscountPercentage)
	}
	if !got.Net.Equal(moneyDec("783")) {
		t.Errorf("net = %s, want 783", got.Net)
	}
	if !got.Tax1Amount.Equal(moneyDec("125.28")) {
		t.Errorf("tax_1_amount = %s, want 125.28", got.Tax1Amount)
	}
}

func TestServiceEntryResponseExposesComputedBreakdown(t *testing.T) {
	// parts 200 + labor 100 => subtotal 300, 10% discount => 30, net 270,
	// tax1 16% => 43.2, tax2 5% => 13.5.
	row := gen.ServiceEntry{
		PartsSubtotal:      moneyDec("200"),
		LaborSubtotal:      moneyDec("100"),
		DiscountType:       money.Percentage,
		DiscountPercentage: moneyDec("10"),
		Tax1Type:           money.Percentage,
		Tax1Percentage:     moneyDec("16"),
		Tax2Type:           money.Percentage,
		Tax2Percentage:     moneyDec("5"),
	}
	got := toServiceEntryResponse(row)
	if !got.DiscountPercentage.Equal(moneyDec("10")) {
		t.Errorf("discount_percentage = %s, want 10", got.DiscountPercentage)
	}
	if !got.Net.Equal(moneyDec("270")) {
		t.Errorf("net = %s, want 270", got.Net)
	}
	if !got.Tax1Amount.Equal(moneyDec("43.2")) {
		t.Errorf("tax_1_amount = %s, want 43.2", got.Tax1Amount)
	}
}
