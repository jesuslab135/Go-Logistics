package dto

import (
	"time"
)

type CreateCommentRequest struct {
	ContentTypeID int64  `json:"content_type_id"`
	ObjectID      int32  `json:"object_id"`
	Body          string `json:"body"`
	AuthorID      *int64 `json:"author_id"`
}

type UpdateCommentRequest struct {
	ContentTypeID int64  `json:"content_type_id"`
	ObjectID      int32  `json:"object_id"`
	Body          string `json:"body"`
	AuthorID      *int64 `json:"author_id"`
}

type CommentResponse struct {
	ID            int64     `json:"id"`
	CompanyID     int64     `json:"company_id"`
	ContentTypeID int64     `json:"content_type_id"`
	ObjectID      int32     `json:"object_id"`
	Body          string    `json:"body"`
	AuthorID      *int64    `json:"author_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CommentPage struct {
	Data    []CommentResponse `json:"data"`
	Total   int64             `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasNext bool              `json:"has_next"`
}
