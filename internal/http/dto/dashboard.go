package dto

// DashboardStatsResponse aggregates the counters the dashboard used to assemble
// from six separate list requests.
type DashboardStatsResponse struct {
	TotalAssets        int64 `json:"total_assets"`
	ActiveWorkOrders   int64 `json:"active_work_orders"`
	OverdueIssues      int64 `json:"overdue_issues"`
	UpcomingReminders  int64 `json:"upcoming_reminders"`
	LowStockParts      int64 `json:"low_stock_parts"`
	PendingInspections int64 `json:"pending_inspections"`
}
