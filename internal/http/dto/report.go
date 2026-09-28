package dto

import "time"

type CreateReportRecipientRequest struct {
	// ReportKind is "fuel_weekly" or "maintenance_monthly".
	ReportKind string `json:"report_kind" binding:"required,max=30"`
	Email      string `json:"email" binding:"required,max=254"`
}

type ReportRecipientResponse struct {
	ID         int64     `json:"id"`
	ReportKind string    `json:"report_kind"`
	Email      string    `json:"email"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type ReportRecipientListResponse struct {
	Data []ReportRecipientResponse `json:"data"`
}

type ReportRunResponse struct {
	ID         int64  `json:"id"`
	ReportKind string `json:"report_kind"`
	// PeriodStart is the first day of the period, as YYYY-MM-DD in the
	// company's own timezone.
	PeriodStart string `json:"period_start"`
	// Status is one of: running, sent, not_sent, skipped_empty,
	// skipped_no_recipients, failed.
	Status     string     `json:"status"`
	Attempts   int32      `json:"attempts"`
	Error      *string    `json:"error"`
	Recipients int32      `json:"recipients"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
}
