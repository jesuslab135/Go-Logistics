package reports

import "github.com/shopspring/decimal"

// topAssets is how many assets the "most expensive" table lists.
const topAssets = 5

// MaintenanceRow is one asset's completed jobs in a period, aggregated by the
// database. Total already honours a manual override, so Parts plus Labor may
// differ from it; HasOverride says when that can be the reason.
type MaintenanceRow struct {
	AssetID      int64
	AssetName    string
	LicensePlate string
	Jobs         int64
	Parts        decimal.Decimal
	Labor        decimal.Decimal
	Total        decimal.Decimal
	HasOverride  bool
}

type MaintenanceLine struct {
	AssetName    string
	LicensePlate string
	Jobs         int64
	Parts        decimal.Decimal
	Labor        decimal.Decimal
	Total        decimal.Decimal
	HasOverride  bool
}

type MaintenanceMonthlyReport struct {
	Header
	// Lines is ordered as the database returned it: most expensive first.
	Lines       []MaintenanceLine
	Top         []MaintenanceLine
	Jobs        int64
	Parts       decimal.Decimal
	Labor       decimal.Decimal
	Total       decimal.Decimal
	HasOverride bool
	Cost        Comparison
}

func (r MaintenanceMonthlyReport) Empty() bool { return len(r.Lines) == 0 }

// BuildMaintenanceMonthly turns the aggregated rows of a month, and of the
// month before it, into the report.
func BuildMaintenanceMonthly(h Header, current, previous []MaintenanceRow) MaintenanceMonthlyReport {
	r := MaintenanceMonthlyReport{Header: h, Lines: make([]MaintenanceLine, 0, len(current))}
	for _, row := range current {
		r.Lines = append(r.Lines, MaintenanceLine{
			AssetName:    row.AssetName,
			LicensePlate: row.LicensePlate,
			Jobs:         row.Jobs,
			Parts:        row.Parts,
			Labor:        row.Labor,
			Total:        row.Total,
			HasOverride:  row.HasOverride,
		})
		r.Jobs += row.Jobs
		r.Parts = r.Parts.Add(row.Parts)
		r.Labor = r.Labor.Add(row.Labor)
		r.Total = r.Total.Add(row.Total)
		r.HasOverride = r.HasOverride || row.HasOverride
	}
	r.Top = r.Lines[:min(topAssets, len(r.Lines))]

	prev := decimal.Zero
	for _, row := range previous {
		prev = prev.Add(row.Total)
	}
	r.Cost = Compare(r.Total, prev)
	return r
}
