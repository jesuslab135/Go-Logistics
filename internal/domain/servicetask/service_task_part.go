package servicetask

import "github.com/shopspring/decimal"

// ServiceTaskPart — port of api/models/service_task_model.py:41

type ServiceTaskPart struct {
	ID            int64
	ServiceTaskID int64
	PartID        int64
	Quantity      decimal.Decimal
	Position      int32
}
