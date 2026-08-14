package handler

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"fleet/internal/db/gen"
)

// Distance and efficiency are derived by the server, ported from Django's
// FuelRecalculationService. A client cannot compute them without holding the
// asset's whole history, and two clients disagreeing would corrupt the series.
//
// The series for a derivation is one asset and one fuel type, ordered by
// (date, id): diesel and DEF fills for the same truck are independent runs.

// fuelCalcRow is the slice of a fuel entry the derivation actually reads.
type fuelCalcRow struct {
	ID       int64
	Date     time.Time
	Odometer decimal.Decimal
	Quantity decimal.Decimal
	FullTank bool
	Reset    bool
}

// deriveFuel returns the distance covered since prev and the efficiency over
// that distance, or a null efficiency where it cannot be known:
//
//   - no previous entry, or reset set — the entry starts a new series;
//   - a non-positive odometer delta — a rollback or a correction, so the
//     distance is not trustworthy;
//   - either tank not filled — the fuel burned over the interval is unknown,
//     though the distance still is;
//   - a zero quantity, which has no meaningful ratio.
func deriveFuel(prev *fuelCalcRow, entry fuelCalcRow) (decimal.Decimal, *decimal.Decimal) {
	if prev == nil || entry.Reset {
		return decimal.Zero, nil
	}

	distance := entry.Odometer.Sub(prev.Odometer)
	if distance.Sign() <= 0 {
		return decimal.Zero, nil
	}
	if entry.FullTank && prev.FullTank && entry.Quantity.Sign() > 0 {
		efficiency := distance.Div(entry.Quantity).Round(2)
		return distance, &efficiency
	}
	return distance, nil
}

// recalcFuelSeries recomputes every entry at or after the given position in a
// series, since each entry's values depend on the one before it. Editing a
// backdated entry, moving one between dates, or deleting one all shift the
// entries that follow, so they are all recomputed rather than left stale.
//
// It must run inside the caller's transaction: ListFuelEntriesFrom locks the
// affected rows so a concurrent write cannot interleave with the walk.
func recalcFuelSeries(ctx context.Context, qtx *gen.Queries, assetID int64, fuelType string, from time.Time, fromID int64) error {
	entries, err := qtx.ListFuelEntriesFrom(ctx, gen.ListFuelEntriesFromParams{
		AssetID:  assetID,
		FuelType: fuelType,
		Date:     from,
		ID:       fromID,
	})
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}

	// Seed from the entry just before the first affected one, which is itself
	// unaffected and therefore not re-read.
	var prev *fuelCalcRow
	head := entries[0]
	previous, err := qtx.GetPreviousFuelEntry(ctx, gen.GetPreviousFuelEntryParams{
		AssetID:  assetID,
		FuelType: fuelType,
		Date:     head.Date,
		ID:       head.ID,
	})
	switch {
	case err == nil:
		prev = &fuelCalcRow{
			ID:       previous.ID,
			Date:     previous.Date,
			Odometer: previous.Odometer,
			Quantity: previous.Quantity,
			FullTank: previous.FullTank,
			Reset:    previous.Reset,
		}
	case errors.Is(err, pgx.ErrNoRows):
		// The first affected entry is the start of the series.
	default:
		return err
	}

	for _, e := range entries {
		row := fuelCalcRow{
			ID:       e.ID,
			Date:     e.Date,
			Odometer: e.Odometer,
			Quantity: e.Quantity,
			FullTank: e.FullTank,
			Reset:    e.Reset,
		}
		miles, efficiency := deriveFuel(prev, row)
		if err := qtx.SetFuelEntryDerived(ctx, gen.SetFuelEntryDerivedParams{
			ID:             e.ID,
			MilesTraveled:  &miles,
			FuelEfficiency: efficiency,
		}); err != nil {
			return err
		}
		prev = &row
	}
	return nil
}
