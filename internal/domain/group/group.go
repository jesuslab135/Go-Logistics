package group

import "time"

// Group — port of api/models/group_model.py:4
// Ancestry caches the materialized path as "id1/id2/id3".

type Group struct {
	ID        int64
	CompanyID int64
	Name      string
	ParentID  *int64
	Ancestry  string
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
