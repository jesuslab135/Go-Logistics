package fuel

import "time"

// FuelComment — port of api/models/fuel_interaction_model.py:7
// UserID references Django auth_user, not Employee.

type FuelComment struct {
	ID        int64
	EntryID   int64
	UserID    int64
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
