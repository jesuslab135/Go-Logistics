package reports

import (
	"context"
	"time"

	"fleet/internal/db/gen"
)

// Querier is the slice of the generated queries the reports read through.
// *gen.Queries satisfies it.
type Querier interface {
	ReportFuelByAsset(ctx context.Context, arg gen.ReportFuelByAssetParams) ([]gen.ReportFuelByAssetRow, error)
	ReportMaintenanceByAsset(ctx context.Context, arg gen.ReportMaintenanceByAssetParams) ([]gen.ReportMaintenanceByAssetRow, error)
}

// FuelWeeklyFor reads one company's week, and the week before it, and builds
// the report.
func FuelWeeklyFor(ctx context.Context, q Querier, c Company, p Period, now time.Time) (FuelWeeklyReport, error) {
	current, err := fuelRows(ctx, q, c.ID, p)
	if err != nil {
		return FuelWeeklyReport{}, err
	}
	previous, err := fuelRows(ctx, q, c.ID, p.Previous(FuelWeekly))
	if err != nil {
		return FuelWeeklyReport{}, err
	}
	h := Header{Kind: FuelWeekly, Company: c, Period: p, GeneratedAt: now}
	return BuildFuelWeekly(h, current, previous), nil
}

func fuelRows(ctx context.Context, q Querier, companyID int64, p Period) ([]FuelRow, error) {
	rows, err := q.ReportFuelByAsset(ctx, gen.ReportFuelByAssetParams{
		CompanyID:   companyID,
		PeriodStart: p.Start,
		PeriodEnd:   p.End,
		Timezone:    p.Timezone(),
	})
	if err != nil {
		return nil, err
	}
	out := make([]FuelRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, FuelRow{
			AssetID:      r.AssetID,
			AssetName:    r.AssetName,
			LicensePlate: r.LicensePlate,
			MeterUnit:    r.MeterUnit,
			VolumeUnit:   r.VolumeUnit,
			FuelType:     r.FuelType,
			Fills:        r.Fills,
			Volume:       r.Volume,
			Cost:         r.Cost,
			Distance:     r.Distance,
			EffDistance:  r.EffDistance,
			EffVolume:    r.EffVolume,
		})
	}
	return out, nil
}

// MaintenanceMonthlyFor reads one company's month, and the month before it,
// and builds the report.
func MaintenanceMonthlyFor(ctx context.Context, q Querier, c Company, p Period, now time.Time) (MaintenanceMonthlyReport, error) {
	current, err := maintenanceRows(ctx, q, c.ID, p)
	if err != nil {
		return MaintenanceMonthlyReport{}, err
	}
	previous, err := maintenanceRows(ctx, q, c.ID, p.Previous(MaintenanceMonthly))
	if err != nil {
		return MaintenanceMonthlyReport{}, err
	}
	h := Header{Kind: MaintenanceMonthly, Company: c, Period: p, GeneratedAt: now}
	return BuildMaintenanceMonthly(h, current, previous), nil
}

func maintenanceRows(ctx context.Context, q Querier, companyID int64, p Period) ([]MaintenanceRow, error) {
	rows, err := q.ReportMaintenanceByAsset(ctx, gen.ReportMaintenanceByAssetParams{
		CompanyID:   companyID,
		PeriodStart: p.Start,
		PeriodEnd:   p.End,
		Timezone:    p.Timezone(),
	})
	if err != nil {
		return nil, err
	}
	out := make([]MaintenanceRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, MaintenanceRow{
			AssetID:      r.AssetID,
			AssetName:    r.AssetName,
			LicensePlate: r.LicensePlate,
			Jobs:         r.Jobs,
			Parts:        r.Parts,
			Labor:        r.Labor,
			Total:        r.Total,
			HasOverride:  r.HasOverride,
		})
	}
	return out, nil
}
