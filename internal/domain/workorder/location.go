package workorder

// Location — port of api/models/work_order_model.py:39

type Location struct {
	ID        int64
	CompanyID int64
	Name      string
	IsActive  bool
}
