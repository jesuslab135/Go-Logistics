package dto

import (
	"time"
)

// CommentContentTypes are the parent kinds a comment may be filed against.
// They replace Django's content_type_id, which held a row id assigned at
// migrate time and was never reproducible in a fresh database.
const (
	CommentContentTypeAsset        = "asset"
	CommentContentTypeIssue        = "issue"
	CommentContentTypeWorkOrder    = "work_order"
	CommentContentTypeServiceEntry = "service_entry"
)

type CreateCommentRequest struct {
	ContentType string `json:"content_type" binding:"required,oneof=asset issue work_order service_entry"`
	ObjectID    int64  `json:"object_id" binding:"required,min=1"`
	Body        string `json:"body" binding:"required"`
}

// The parent and the author are fixed at creation; an edit changes the body.
type UpdateCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

type CommentResponse struct {
	ID          int64     `json:"id"`
	CompanyID   int64     `json:"company_id"`
	ContentType string    `json:"content_type"`
	ObjectID    int64     `json:"object_id"`
	Body        string    `json:"body"`
	AuthorID    *int64    `json:"author_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CommentPage struct {
	Data    []CommentResponse `json:"data"`
	Total   int64             `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasNext bool              `json:"has_next"`
}
