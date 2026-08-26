package dto

import (
	"time"
)

// The status log is append-only and has no request DTOs: rows are written by
// the operation that changes a work order's status, in the same transaction,
// not through a route. POST, PUT and DELETE on the nested path answer 405
// `route_retired`.

type WorkOrderStatusLogResponse struct {
	ID          int64     `json:"id"`
	WorkOrderID int64     `json:"work_order_id"`
	StatusID    int64     `json:"status_id"`
	ChangedAt   time.Time `json:"changed_at"`
	// ActorEmployeeID is null for a system transition, and for rows written
	// before actors were recorded. Read ActorType to tell those apart.
	ActorEmployeeID *int64 `json:"actor_employee_id"`
	// ActorType is "employee" or "system". Rows that predate this field are
	// "system": they were written through the old CRUD route with no actor
	// captured, so naming a person would be a fabrication.
	ActorType string `json:"actor_type"`
}

type WorkOrderStatusLogPage struct {
	Data    []WorkOrderStatusLogResponse `json:"data"`
	Total   int64                        `json:"total"`
	Limit   int                          `json:"limit"`
	Offset  int                          `json:"offset"`
	HasNext bool                         `json:"has_next"`
}
