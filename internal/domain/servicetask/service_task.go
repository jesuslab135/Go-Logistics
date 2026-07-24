package servicetask

import "time"

// ServiceTask — port of api/models/service_task_model.py:4
// ParentTaskID is a self-reference for subtasks.

type ServiceTask struct {
	ID                      int64
	CompanyID               int64
	Name                    string
	Description             string
	ExpectedDurationSeconds *int32
	ParentTaskID            *int64
	ArchivedAt              *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func (t ServiceTask) IsActive() bool {
	return t.ArchivedAt == nil
}
