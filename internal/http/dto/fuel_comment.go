package dto

import (
	"time"
)

// The author is taken from the authenticated context, never from the body.
type CreateFuelCommentRequest struct {
	Text string `json:"text" binding:"required"`
}

type UpdateFuelCommentRequest struct {
	Text string `json:"text" binding:"required"`
}

type FuelCommentResponse struct {
	ID         int64     `json:"id"`
	EntryID    int64     `json:"entry_id"`
	EmployeeID *int64    `json:"employee_id"`
	Text       string    `json:"text"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FuelCommentPage struct {
	Data    []FuelCommentResponse `json:"data"`
	Total   int64                 `json:"total"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
	HasNext bool                  `json:"has_next"`
}
