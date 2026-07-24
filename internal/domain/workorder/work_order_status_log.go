package workorder

import "time"

// WorkOrderStatusLog — port of api/models/work_order_model.py:458
// Django fed this via signals; per the guide those become explicit service calls.

type WorkOrderStatusLog struct {
	ID          int64
	WorkOrderID int64
	StatusID    int64
	ChangedAt   time.Time
}
