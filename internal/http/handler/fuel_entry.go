package handler

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/dbctx"
	"fleet/internal/platform/paginate"
)

type FuelEntryStore struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewFuelEntryStore(q *gen.Queries, pool *pgxpool.Pool) *FuelEntryStore {
	return &FuelEntryStore{q: q, pool: pool}
}

// inSeriesTx runs a write and the recalculation it invalidates in one
// transaction, so no reader sees an entry whose successors are stale.
func (s *FuelEntryStore) inSeriesTx(ctx context.Context, fn func(qtx *gen.Queries) error) error {
	tx, err := dbctx.Begin(ctx, s.pool)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *FuelEntryStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.FuelEntryResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListFuelEntries(ctx, gen.ListFuelEntriesParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountFuelEntries(ctx, gen.CountFuelEntriesParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.FuelEntryResponse, len(rows))
	for i, r := range rows {
		out[i] = toFuelEntryResponse(r)
	}
	return out, total, nil
}

func (s *FuelEntryStore) Get(ctx context.Context, parentID, id int64) (dto.FuelEntryResponse, error) {
	r, err := s.q.GetFuelEntry(ctx, gen.GetFuelEntryParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.FuelEntryResponse{}, err
	}
	return toFuelEntryResponse(r), nil
}

func (s *FuelEntryStore) Create(ctx context.Context, parentID int64, in dto.CreateFuelEntryRequest) (dto.FuelEntryResponse, error) {
	now := time.Now().UTC()
	var created gen.FuelEntry

	err := s.inSeriesTx(ctx, func(qtx *gen.Queries) error {
		r, err := qtx.CreateFuelEntry(ctx, gen.CreateFuelEntryParams{
			ParentID:     parentID,
			CompanyID:    middleware.CompanyFromContext(ctx),
			EmployeeID:   in.EmployeeID,
			Date:         in.Date,
			FuelType:     in.FuelType,
			Quantity:     in.Quantity,
			UnitCost:     in.UnitCost,
			TotalCost:    in.TotalCost,
			Odometer:     in.Odometer,
			VendorID:     in.VendorID,
			FullTank:     boolOrDefault(in.FullTank, true),
			State:        in.State,
			Reference:    in.Reference,
			Personal:     in.Personal,
			Reset:        in.Reset,
			Latitude:     in.Latitude,
			Longitude:    in.Longitude,
			ExternalID:   in.ExternalID,
			NoSemana:     in.NoSemana,
			EstadoProv:   in.EstadoProv,
			OperatorName: in.OperatorName,
			UpdatedAt:    now,
		})
		if err != nil {
			return err
		}
		// A backdated entry changes the entries after it too, so the series is
		// recomputed from the new row rather than only for the new row.
		if err := recalcFuelSeries(ctx, qtx, r.AssetID, r.FuelType, r.Date, r.ID); err != nil {
			return err
		}
		created, err = qtx.GetFuelEntry(ctx, gen.GetFuelEntryParams{
			ID: r.ID, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx),
		})
		return err
	})
	if err != nil {
		return dto.FuelEntryResponse{}, err
	}
	return toFuelEntryResponse(created), nil
}

func (s *FuelEntryStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateFuelEntryRequest) (dto.FuelEntryResponse, error) {
	now := time.Now().UTC()
	company := middleware.CompanyFromContext(ctx)
	var updated gen.FuelEntry

	err := s.inSeriesTx(ctx, func(qtx *gen.Queries) error {
		// The edit may move the entry to another date, asset or fuel type, so
		// the series it is leaving has to be recomputed as well as the one it
		// joins. Both are recomputed from the earlier of the two positions.
		before, err := qtx.GetFuelEntrySeries(ctx, gen.GetFuelEntrySeriesParams{ID: id, CompanyID: company})
		if err != nil {
			return err
		}

		r, err := qtx.UpdateFuelEntry(ctx, gen.UpdateFuelEntryParams{
			ID:           id,
			ParentID:     parentID,
			CompanyID:    company,
			EmployeeID:   in.EmployeeID,
			Date:         in.Date,
			FuelType:     in.FuelType,
			Quantity:     in.Quantity,
			UnitCost:     in.UnitCost,
			TotalCost:    in.TotalCost,
			Odometer:     in.Odometer,
			VendorID:     in.VendorID,
			FullTank:     in.FullTank,
			State:        in.State,
			Reference:    in.Reference,
			Personal:     in.Personal,
			Reset:        in.Reset,
			Latitude:     in.Latitude,
			Longitude:    in.Longitude,
			ExternalID:   in.ExternalID,
			UpdatedAt:    now,
			NoSemana:     in.NoSemana,
			EstadoProv:   in.EstadoProv,
			OperatorName: in.OperatorName,
		})
		if err != nil {
			return err
		}

		from, fromID := before.Date, before.ID
		if r.AssetID == before.AssetID && r.FuelType == before.FuelType {
			// Same series: recompute from whichever position comes first.
			if r.Date.Before(from) {
				from, fromID = r.Date, r.ID
			}
		} else {
			// It moved series: repair the one it left, then recompute the new
			// one from the entry's new position.
			if err := recalcFuelSeries(ctx, qtx, before.AssetID, before.FuelType, before.Date, before.ID); err != nil {
				return err
			}
			from, fromID = r.Date, r.ID
		}
		if err := recalcFuelSeries(ctx, qtx, r.AssetID, r.FuelType, from, fromID); err != nil {
			return err
		}

		updated, err = qtx.GetFuelEntry(ctx, gen.GetFuelEntryParams{ID: id, ParentID: parentID, CompanyID: company})
		return err
	})
	if err != nil {
		return dto.FuelEntryResponse{}, err
	}
	return toFuelEntryResponse(updated), nil
}

func (s *FuelEntryStore) Delete(ctx context.Context, parentID, id int64) error {
	company := middleware.CompanyFromContext(ctx)

	// Removing an entry changes what the entries after it are measured
	// against, so they are recomputed too. Django skipped this, which left
	// stale distances behind every deletion.
	return s.inSeriesTx(ctx, func(qtx *gen.Queries) error {
		removed, err := qtx.GetFuelEntrySeries(ctx, gen.GetFuelEntrySeriesParams{ID: id, CompanyID: company})
		if err != nil {
			return err
		}
		if err := qtx.DeleteFuelEntry(ctx, gen.DeleteFuelEntryParams{ID: id, ParentID: parentID, CompanyID: company}); err != nil {
			return err
		}
		return recalcFuelSeries(ctx, qtx, removed.AssetID, removed.FuelType, removed.Date, removed.ID)
	})
}

func toFuelEntryResponse(r gen.FuelEntry) dto.FuelEntryResponse {
	return dto.FuelEntryResponse{
		ID:             r.ID,
		AssetID:        r.AssetID,
		EmployeeID:     r.EmployeeID,
		Date:           r.Date,
		FuelType:       r.FuelType,
		Quantity:       r.Quantity,
		UnitCost:       r.UnitCost,
		TotalCost:      r.TotalCost,
		Odometer:       r.Odometer,
		VendorID:       r.VendorID,
		FullTank:       r.FullTank,
		MilesTraveled:  r.MilesTraveled,
		FuelEfficiency: r.FuelEfficiency,
		State:          r.State,
		Reference:      r.Reference,
		Personal:       r.Personal,
		Reset:          r.Reset,
		Latitude:       r.Latitude,
		Longitude:      r.Longitude,
		ExternalID:     r.ExternalID,
		UpdatedAt:      r.UpdatedAt,
		NoSemana:       r.NoSemana,
		EstadoProv:     r.EstadoProv,
		OperatorName:   r.OperatorName,
	}
}
