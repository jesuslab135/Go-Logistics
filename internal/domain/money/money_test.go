package money

import (
	"testing"

	"github.com/shopspring/decimal"
)

func dec(v string) decimal.Decimal {
	d, err := decimal.NewFromString(v)
	if err != nil {
		panic(err)
	}
	return d
}

func pct(v string) Rate   { return Rate{Type: Percentage, Percentage: dec(v)} }
func fixed(v string) Rate { return Rate{Type: Fixed, Fixed: dec(v)} }

// The order of operations is the whole decision. Every case here would produce a
// different, plausible number under a different order, which is why they are
// pinned with the arithmetic spelled out.
func TestCompute(t *testing.T) {
	tests := []struct {
		name string
		doc  Document
		want Totals
	}{
		{
			name: "an empty document totals nothing",
			doc:  Document{},
			want: Totals{},
		},
		{
			name: "line items with no terms pass straight through",
			doc:  Document{PartsSubtotal: dec("100"), LaborSubtotal: dec("50")},
			want: Totals{
				PartsSubtotal: dec("100"), LaborSubtotal: dec("50"),
				Subtotal: dec("150"), Net: dec("150"), Total: dec("150"),
			},
		},
		{
			name: "markup applies per category, not to the combined subtotal",
			// parts 100 +10% = 110; labor 50 +20% = 60; subtotal 170.
			// A single markup on the 150 combined would give a different answer.
			doc: Document{
				PartsSubtotal: dec("100"), LaborSubtotal: dec("50"),
				PartsMarkup: pct("10"), LaborMarkup: pct("20"),
			},
			want: Totals{
				PartsSubtotal: dec("100"), LaborSubtotal: dec("50"),
				Subtotal: dec("170"), Net: dec("170"), Total: dec("170"),
			},
		},
		{
			name: "tax applies after the discount, not before",
			// subtotal 200, less 10% = 180 net, then 16% IVA = 28.80 -> 208.80.
			// Taxing first would give 232 - 23.20 = 208.80 by coincidence at a
			// single rate, so the discount here is fixed to break the symmetry.
			doc: Document{
				PartsSubtotal: dec("200"),
				Discount:      fixed("20"),
				Tax1:          pct("16"),
			},
			want: Totals{
				PartsSubtotal: dec("200"), Subtotal: dec("200"),
				DiscountAmount: dec("20"), Net: dec("180"),
				Tax1Amount: dec("28.8"), Total: dec("208.8"),
			},
		},
		{
			name: "the two taxes share one base and do not compound",
			// net 100: IVA 16% = 16, retención 4% = 4 - both on 100.
			// Compounding would make the second 4% of 116 = 4.64.
			doc: Document{
				PartsSubtotal: dec("100"),
				Tax1:          pct("16"),
				Tax2:          pct("4"),
			},
			want: Totals{
				PartsSubtotal: dec("100"), Subtotal: dec("100"), Net: dec("100"),
				Tax1Amount: dec("16"), Tax2Amount: dec("4"), Total: dec("120"),
			},
		},
		{
			name: "shipping is added after tax and is not taxed",
			// net 100, tax 16 -> 116, plus 50 freight = 166.
			// Taxing the freight would give 174.
			doc: Document{
				PartsSubtotal: dec("100"),
				Tax1:          pct("16"),
				Shipping:      dec("50"),
			},
			want: Totals{
				PartsSubtotal: dec("100"), Subtotal: dec("100"), Net: dec("100"),
				Tax1Amount: dec("16"), Shipping: dec("50"), Total: dec("166"),
			},
		},
		{
			name: "a fixed discount is taken at face value",
			doc: Document{
				PartsSubtotal: dec("100"),
				Discount:      fixed("15.50"),
			},
			want: Totals{
				PartsSubtotal: dec("100"), Subtotal: dec("100"),
				DiscountAmount: dec("15.5"), Net: dec("84.5"), Total: dec("84.5"),
			},
		},
		{
			name: "a discount larger than the subtotal goes negative rather than clamping",
			// Clamping to zero would hide a data-entry error behind a plausible
			// total; a negative total is visibly wrong, which is the point.
			doc: Document{
				PartsSubtotal: dec("50"),
				Discount:      fixed("80"),
			},
			want: Totals{
				PartsSubtotal: dec("50"), Subtotal: dec("50"),
				DiscountAmount: dec("80"), Net: dec("-30"), Total: dec("-30"),
			},
		},
		{
			name: "an unrecognised rate type contributes nothing",
			// Better a term that is visibly missing than one silently guessed.
			doc: Document{
				PartsSubtotal: dec("100"),
				Tax1:          Rate{Type: "BOGUS", Percentage: dec("16"), Fixed: dec("16")},
			},
			want: Totals{
				PartsSubtotal: dec("100"), Subtotal: dec("100"),
				Net: dec("100"), Total: dec("100"),
			},
		},
		{
			name: "a full work order: markup, discount and both taxes",
			// parts 1000 +15% = 1150; labor 400 +25% = 500; subtotal 1650.
			// less 5% (82.50) = 1567.50 net.
			// IVA 16% = 250.80; retención 4% = 62.70. Total 1881.00.
			doc: Document{
				PartsSubtotal: dec("1000"), LaborSubtotal: dec("400"),
				PartsMarkup: pct("15"), LaborMarkup: pct("25"),
				Discount: pct("5"),
				Tax1:     pct("16"), Tax2: pct("4"),
			},
			want: Totals{
				PartsSubtotal: dec("1000"), LaborSubtotal: dec("400"),
				Subtotal: dec("1650"), DiscountAmount: dec("82.5"), Net: dec("1567.5"),
				Tax1Amount: dec("250.8"), Tax2Amount: dec("62.7"), Total: dec("1881"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compute(tt.doc)
			assertEqual(t, "parts_subtotal", got.PartsSubtotal, tt.want.PartsSubtotal)
			assertEqual(t, "labor_subtotal", got.LaborSubtotal, tt.want.LaborSubtotal)
			assertEqual(t, "subtotal", got.Subtotal, tt.want.Subtotal)
			assertEqual(t, "discount", got.DiscountAmount, tt.want.DiscountAmount)
			assertEqual(t, "net", got.Net, tt.want.Net)
			assertEqual(t, "tax_1", got.Tax1Amount, tt.want.Tax1Amount)
			assertEqual(t, "tax_2", got.Tax2Amount, tt.want.Tax2Amount)
			assertEqual(t, "shipping", got.Shipping, tt.want.Shipping)
			assertEqual(t, "total", got.Total, tt.want.Total)
		})
	}
}

// Every component is persisted, so each is rounded on its own. If only the total
// were rounded, the stored components would not add up to the stored total and
// an invoice would visibly fail to reconcile.
func TestComponentsAreRoundedIndividually(t *testing.T) {
	// 33.333 at 16% is 5.33328 -> 5.33; the total must use that rounded figure,
	// not the unrounded one.
	got := Compute(Document{
		PartsSubtotal: dec("33.333"),
		Tax1:          pct("16"),
	})

	assertEqual(t, "subtotal", got.Subtotal, dec("33.33"))
	assertEqual(t, "tax_1", got.Tax1Amount, dec("5.33"))
	assertEqual(t, "total", got.Total, dec("38.66"))

	if sum := got.Net.Add(got.Tax1Amount).Add(got.Tax2Amount).Add(got.Shipping); !sum.Equal(got.Total) {
		t.Errorf("components sum to %s but total is %s", sum, got.Total)
	}
}

// Half rounds away from zero, which is what an invoice reader expects.
func TestRoundingIsHalfUp(t *testing.T) {
	got := Compute(Document{PartsSubtotal: dec("0.125")})
	assertEqual(t, "subtotal", got.Subtotal, dec("0.13"))
}

func assertEqual(t *testing.T, field string, got, want decimal.Decimal) {
	t.Helper()
	if !got.Equal(want) {
		t.Errorf("%s = %s, want %s", field, got, want)
	}
}
