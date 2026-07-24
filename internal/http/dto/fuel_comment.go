package dto

import (
	"time"
)

type CreateFuelCommentRequest struct {
	UserID int64  `json:"user_id"`
	Text   string `json:"text"`
}

type UpdateFuelCommentRequest struct {
	UserID int64  `json:"user_id"`
	Text   string `json:"text"`
}

type FuelCommentResponse struct {
	ID        int64     `json:"id"`
	EntryID   int64     `json:"entry_id"`
	UserID    int64     `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FuelCommentPage struct {
	Data    []FuelCommentResponse `json:"data"`
	Total   int64                 `json:"total"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
	HasNext bool                  `json:"has_next"`
}
