package handler

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"fleet/internal/db/gen"
	"fleet/internal/domain/money"
)

// Totals are recomputed by the server, never accepted from a client.
//
// Every recompute runs inside the transaction that caused it, so a document's
// stored totals always describe its stored lines. Doing it afterwards, or on
// read, would leave a window where they disagree — and on a document somebody
// is about to invoice, that window is the whole problem.
//
// The trigger is any write that can change an input: the document's own policy
// terms (rates, types, percentages), or any line beneath it.

// recalcWorkOrder recomputes and stores a work order's totals.
func recalcWorkOrder(ctx context.Context, qtx *gen.Queries, workOrderID int64, at time.Time) error {
	order, err := qtx.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return err
	}

	sums, err := qtx.SumWorkOrderLineCosts(ctx, workOrderID)
	if err != nil {
		return err
	}

	totals := money.Compute(money.Document{
		PartsSubtotal: sums.PartsSubtotal,
		LaborSubtotal: sums.LaborSubtotal,
		PartsMarkup:   money.Rate{Type: order.PartsMarkupType, Fixed: order.PartsMarkup, Percentage: order.PartsMarkupPercentage},
		LaborMarkup:   money.Rate{Type: order.LaborMarkupType, Fixed: order.LaborMarkup, Percentage: order.LaborMarkupPercentage},
		Discount:      money.Rate{Type: order.DiscountType, Fixed: order.Discount, Percentage: order.DiscountPercentage},
		Tax1:          money.Rate{Type: order.Tax1Type, Fixed: order.Tax1, Percentage: order.Tax1Percentage},
		Tax2:          money.Rate{Type: order.Tax2Type, Fixed: order.Tax2, Percentage: order.Tax2Percentage},
	})

	return qtx.StoreWorkOrderTotals(ctx, gen.StoreWorkOrderTotalsParams{
		ID:            workOrderID,
		PartsSubtotal: totals.PartsSubtotal,
		LaborSubtotal: totals.LaborSubtotal,
		Subtotal:      totals.Subtotal,
		TotalAmount:   effectiveTotal(totals, order.TotalOverride),
		UpdatedAt:     at,
	})
}

// recalcServiceEntry recomputes and stores a service entry's totals. Service
// entries have no markup columns, so those terms are absent rather than zero —
// the same formula with two fewer inputs.
func recalcServiceEntry(ctx context.Context, qtx *gen.Queries, entryID int64, at time.Time) error {
	entry, err := qtx.GetServiceEntryByID(ctx, entryID)
	if err != nil {
		return err
	}

	sums, err := qtx.SumServiceEntryLineCosts(ctx, entryID)
	if err != nil {
		return err
	}

	totals := money.Compute(money.Document{
		PartsSubtotal: sums.PartsSubtotal,
		LaborSubtotal: sums.LaborSubtotal,
		Discount:      money.Rate{Type: entry.DiscountType, Fixed: entry.Discount, Percentage: entry.DiscountPercentage},
		Tax1:          money.Rate{Type: entry.Tax1Type, Fixed: entry.Tax1, Percentage: entry.Tax1Percentage},
		Tax2:          money.Rate{Type: entry.Tax2Type, Fixed: entry.Tax2, Percentage: entry.Tax2Percentage},
	})

	return qtx.StoreServiceEntryTotals(ctx, gen.StoreServiceEntryTotalsParams{
		ID:            entryID,
		PartsSubtotal: totals.PartsSubtotal,
		LaborSubtotal: totals.LaborSubtotal,
		Subtotal:      totals.Subtotal,
		TotalAmount:   effectiveTotal(totals, entry.TotalOverride),
		UpdatedAt:     at,
	})
}

// recalcPurchaseOrder recomputes and stores a purchase order's totals. A
// purchase order buys parts, so its lines are one undifferentiated subtotal with
// no markup — and it is the only surface with shipping.
func recalcPurchaseOrder(ctx context.Context, qtx *gen.Queries, orderID int64, at time.Time) error {
	order, err := qtx.GetPurchaseOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	subtotal, err := qtx.SumPurchaseOrderLineCosts(ctx, orderID)
	if err != nil {
		return err
	}

	totals := money.Compute(money.Document{
		PartsSubtotal: subtotal,
		Discount:      money.Rate{Type: order.DiscountType, Fixed: order.Discount, Percentage: order.DiscountPercentage},
		Tax1:          money.Rate{Type: order.Tax1Type, Fixed: order.Tax1, Percentage: order.Tax1Percentage},
		Tax2:          money.Rate{Type: order.Tax2Type, Fixed: order.Tax2, Percentage: order.Tax2Percentage},
		Shipping:      order.Shipping,
	})

	return qtx.StorePurchaseOrderTotals(ctx, gen.StorePurchaseOrderTotalsParams{
		ID:          orderID,
		Subtotal:    totals.Subtotal,
		TotalAmount: effectiveTotal(totals, order.TotalOverride),
		UpdatedAt:   at,
	})
}

// effectiveTotal is the override where one stands, and the computed total
// otherwise.
//
// The components beside it are still stored computed: an override replaces the
// bottom line, not the workings, and a reader needs to be able to see the
// difference between what the document adds up to and what it is being charged
// at. That difference is the entire reason an override is audited.
func effectiveTotal(totals money.Totals, override *decimal.Decimal) decimal.Decimal {
	if override != nil {
		return *override
	}
	return totals.Total
}

// recalcWorkOrderLineItem recomputes one line from its sub-line items, then the
// document from all of them. Both levels are stored, so both have to be
// derived: a line showing a subtotal its sub-lines do not add up to is as wrong
// as a document showing one its lines do not.
func recalcWorkOrderLineItem(ctx context.Context, qtx *gen.Queries, lineItemID int64, at time.Time) error {
	sums, err := qtx.SumWorkOrderSubLineCosts(ctx, lineItemID)
	if err != nil {
		return err
	}
	if err := qtx.StoreWorkOrderLineItemTotals(ctx, gen.StoreWorkOrderLineItemTotalsParams{
		ID:        lineItemID,
		PartsCost: sums.PartsCost,
		LaborCost: sums.LaborCost,
		Subtotal:  sums.Subtotal,
		UpdatedAt: at,
	}); err != nil {
		return err
	}

	workOrderID, err := qtx.GetWorkOrderIDForLineItem(ctx, lineItemID)
	if err != nil {
		return err
	}
	// The lines changed, so any override on the document no longer describes it.
	if err := qtx.ClearWorkOrderTotalOverride(ctx, workOrderID); err != nil {
		return err
	}
	return recalcWorkOrder(ctx, qtx, workOrderID, at)
}

// recalcPurchaseOrderLineItem prices the line and then the order.
func recalcPurchaseOrderLineItem(ctx context.Context, qtx *gen.Queries, lineItemID int64, at time.Time) error {
	if err := qtx.StorePurchaseOrderLineItemSubtotal(ctx, gen.StorePurchaseOrderLineItemSubtotalParams{
		ID: lineItemID, UpdatedAt: at,
	}); err != nil {
		return err
	}
	orderID, err := qtx.GetPurchaseOrderIDForLineItem(ctx, lineItemID)
	if err != nil {
		return err
	}
	if err := qtx.ClearPurchaseOrderTotalOverride(ctx, orderID); err != nil {
		return err
	}
	return recalcPurchaseOrder(ctx, qtx, orderID, at)
}

// recalcServiceEntryLineItem prices the line and then the entry.
//
// The line's subtotal is derived (quantity x unit_cost), but parts_cost and
// labor_cost are not: they are the operator's split of that line between the
// two buckets, which the schema gives every line and which nothing else can
// infer. The entry's parts_subtotal and labor_subtotal sum them.
func recalcServiceEntryLineItem(ctx context.Context, qtx *gen.Queries, lineItemID int64, at time.Time) error {
	if err := qtx.StoreServiceEntryLineItemSubtotal(ctx, gen.StoreServiceEntryLineItemSubtotalParams{
		ID: lineItemID, UpdatedAt: at,
	}); err != nil {
		return err
	}
	entryID, err := qtx.GetServiceEntryIDForLineItem(ctx, lineItemID)
	if err != nil {
		return err
	}
	if err := qtx.ClearServiceEntryTotalOverride(ctx, entryID); err != nil {
		return err
	}
	return recalcServiceEntry(ctx, qtx, entryID, at)
}

// inTx runs fn in a transaction, committing only if the write and the recompute
// that follows it both succeed. Every money write goes through this: a document
// whose totals were not updated is worse than one that was not written at all,
// because it looks finished.
func inTx(ctx context.Context, pool *pgxpool.Pool, q *gen.Queries, fn func(qtx *gen.Queries) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
