// Package money computes document totals.
//
// This is the one place the formula lives. Three surfaces — work orders,
// purchase orders and service entries — carry roughly eighteen money columns
// each, and before this package nothing computed any of them: the client sent
// every value and the server stored it. Three UIs each doing their own
// arithmetic is three answers to the same question, and the first time they
// disagree it is on an invoice somebody has already paid.
//
// The Django original did not compute them either. Its money fields are plain
// DecimalFields with no save() override, and the only derived values anywhere
// are three read-only serializer properties on work orders that sum
// quantity × unit_cost and ignore markup, discount and tax entirely. So this is
// new design rather than a port, and the choices it encodes are recorded here
// rather than left to be inferred from the code.
package money

import "github.com/shopspring/decimal"

// How a discount, tax or markup is expressed. Both forms exist in the schema:
// a fixed amount, or a percentage of whatever base the term applies to.
const (
	Fixed      = "FIXED"
	Percentage = "PERCENTAGE"
)

// scale is the precision every persisted money column has. Each named component
// is rounded to it, not just the final total: the components are stored, and
// rounding only at the end would persist a set of values that do not add up to
// the total sitting beside them.
const scale = 2

// Rate is a discount, tax or markup term: a type, a fixed amount, and a
// percentage. Which of the latter two is used depends on the type.
type Rate struct {
	Type       string
	Fixed      decimal.Decimal
	Percentage decimal.Decimal
}

// Amount resolves the rate against a base. An unrecognised type contributes
// nothing rather than guessing: silently treating an unknown type as a
// percentage would produce a plausible number that is wrong.
func (r Rate) Amount(base decimal.Decimal) decimal.Decimal {
	switch r.Type {
	case Fixed:
		return round(r.Fixed)
	case Percentage:
		return round(base.Mul(r.Percentage).Div(decimal.NewFromInt(100)))
	default:
		return decimal.Zero
	}
}

// Totals is a computed document. Every field is derived; none is accepted from
// a client.
type Totals struct {
	PartsSubtotal  decimal.Decimal
	LaborSubtotal  decimal.Decimal
	Subtotal       decimal.Decimal
	DiscountAmount decimal.Decimal
	// Net is the subtotal after discount, and the base both taxes apply to.
	Net        decimal.Decimal
	Tax1Amount decimal.Decimal
	Tax2Amount decimal.Decimal
	Shipping   decimal.Decimal
	Total      decimal.Decimal
}

// Document is everything the formula needs: the line-item sums, and the policy
// terms stored on the parent record.
type Document struct {
	// PartsSubtotal and LaborSubtotal come from the line items. A caller that
	// has only one undifferentiated subtotal (a purchase order) puts it in
	// PartsSubtotal and leaves the markups zero.
	PartsSubtotal decimal.Decimal
	LaborSubtotal decimal.Decimal

	// PartsMarkup and LaborMarkup apply per category, and only work orders have
	// them: the columns are on work_order alone. They are aggregate, not
	// per-line, because that is where the schema puts them — a per-line markup
	// would need columns that do not exist.
	PartsMarkup Rate
	LaborMarkup Rate

	Discount Rate
	Tax1     Rate
	Tax2     Rate

	// Shipping is a purchase-order term. It is added after tax and is not
	// taxed: it sits after the tax columns in the schema, and taxing freight
	// would need a rate of its own.
	Shipping decimal.Decimal
}

// Compute applies the formula:
//
//	parts_marked  = parts_subtotal + markup(parts)
//	labor_marked  = labor_subtotal + markup(labor)
//	subtotal      = parts_marked + labor_marked
//	net           = subtotal - discount(subtotal)
//	total         = net + tax1(net) + tax2(net) + shipping
//
// Two choices in there are worth stating, because both are defensible the other
// way and neither is recoverable from the column names:
//
// Tax applies to the net, after the discount. company.currency defaults to MXN
// and Mexican IVA is charged on the value after commercial discounts; the same
// order is what Fleetio, which the markup columns were copied from, uses.
//
// tax_1 and tax_2 both apply to that same net and do not compound. They are two
// jurisdictions charged on one taxable value — in Mexico, IVA alongside a
// retención — not a tax levied on a tax.
func Compute(d Document) Totals {
	partsSubtotal := round(d.PartsSubtotal)
	laborSubtotal := round(d.LaborSubtotal)

	partsMarked := partsSubtotal.Add(d.PartsMarkup.Amount(partsSubtotal))
	laborMarked := laborSubtotal.Add(d.LaborMarkup.Amount(laborSubtotal))

	subtotal := round(partsMarked.Add(laborMarked))
	discount := d.Discount.Amount(subtotal)
	net := round(subtotal.Sub(discount))

	tax1 := d.Tax1.Amount(net)
	tax2 := d.Tax2.Amount(net)
	shipping := round(d.Shipping)

	return Totals{
		PartsSubtotal:  partsSubtotal,
		LaborSubtotal:  laborSubtotal,
		Subtotal:       subtotal,
		DiscountAmount: discount,
		Net:            net,
		Tax1Amount:     tax1,
		Tax2Amount:     tax2,
		Shipping:       shipping,
		Total:          round(net.Add(tax1).Add(tax2).Add(shipping)),
	}
}

// round is half-away-from-zero at the persisted scale, which is what an invoice
// reader expects: 0.125 becomes 0.13, not 0.12.
func round(d decimal.Decimal) decimal.Decimal {
	return d.Round(scale)
}
