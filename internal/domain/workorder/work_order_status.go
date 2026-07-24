package workorder

// WorkOrderStatus — port of api/models/work_order_model.py:5

type WorkOrderStatus struct {
	ID               int64
	CompanyID        int64
	Name             string
	Description      string
	Color            string
	IsDefault        bool
	MarksAsCompleted bool
	Position         int32
}
