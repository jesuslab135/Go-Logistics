package dto

// Link request bodies. A join table carries no attributes of its own, so the
// body is just the far side of the relationship.

type LinkEmployeeRequest struct {
	EmployeeID int64 `json:"employee_id" binding:"required,min=1"`
}

type LinkIssueRequest struct {
	IssueID int64 `json:"issue_id" binding:"required,min=1"`
}

type LinkFaultRequest struct {
	FaultID int64 `json:"fault_id" binding:"required,min=1"`
}
