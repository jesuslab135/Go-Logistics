package dto

// DashboardStatsResponse aggregates the counters the dashboard used to assemble
// from six separate list requests.
// DashboardStatsResponse is a set of current-state counters.
//
// Two names are narrower than they sound, and the calculation is what counts:
// pending_inspections counts inspection submissions with failed items and no
// issue raised from them, and low_stock_parts counts inventory rows (a part at
// one location) at or below their reorder point, not distinct parts. They keep
// their names because the generated client already uses them; the definitions
// are documented here and in the query instead.
//
// There are no trends: real ones need periodic snapshots, and computing a
// previous period from current-state counters would only invent them.
type DashboardStatsResponse struct {
	TotalAssets        int64 `json:"total_assets"`
	ActiveWorkOrders   int64 `json:"active_work_orders"`
	OverdueIssues      int64 `json:"overdue_issues"`
	UpcomingReminders  int64 `json:"upcoming_reminders"`
	LowStockParts      int64 `json:"low_stock_parts"`
	PendingInspections int64 `json:"pending_inspections"`

	// The window upcoming_reminders was counted over, so UI copy can name the
	// same number the backend used instead of hardcoding one.
	UpcomingDays int32 `json:"upcoming_days"`
}
