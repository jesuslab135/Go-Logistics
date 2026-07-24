package dto

import (
	"time"
)

type CreateWorkOrderStatusLogRequest struct {
	StatusID  int64     `json:"status_id"`
	ChangedAt time.Time `json:"changed_at"`
}

type UpdateWorkOrderStatusLogRequest struct {
	StatusID  int64     `json:"status_id"`
	ChangedAt time.Time `json:"changed_at"`
}

type WorkOrderStatusLogResponse struct {
	ID          int64     `json:"id"`
	WorkOrderID int64     `json:"work_order_id"`
	StatusID    int64     `json:"status_id"`
	ChangedAt   time.Time `json:"changed_at"`
}

type WorkOrderStatusLogPage struct {
	Data    []WorkOrderStatusLogResponse `json:"data"`
	Total   int64                        `json:"total"`
	Limit   int                          `json:"limit"`
	Offset  int                          `json:"offset"`
	HasNext bool                         `json:"has_next"`
}
