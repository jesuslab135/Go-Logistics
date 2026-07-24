package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type CommentStore struct{ q *gen.Queries }

func NewCommentStore(q *gen.Queries) *CommentStore { return &CommentStore{q: q} }

func (s *CommentStore) List(ctx context.Context, p paginate.Params) ([]dto.CommentResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListComments(ctx, gen.ListCommentsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountComments(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.CommentResponse, len(rows))
	for i, r := range rows {
		out[i] = toCommentResponse(r)
	}
	return out, total, nil
}

func (s *CommentStore) Get(ctx context.Context, id int64) (dto.CommentResponse, error) {
	r, err := s.q.GetComment(ctx, gen.GetCommentParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.CommentResponse{}, err
	}
	return toCommentResponse(r), nil
}

func (s *CommentStore) Create(ctx context.Context, in dto.CreateCommentRequest) (dto.CommentResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateComment(ctx, gen.CreateCommentParams{
		CompanyID:     middleware.CompanyFromContext(ctx),
		ContentTypeID: in.ContentTypeID,
		ObjectID:      in.ObjectID,
		Body:          in.Body,
		AuthorID:      in.AuthorID,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		return dto.CommentResponse{}, err
	}
	return toCommentResponse(r), nil
}

func (s *CommentStore) Update(ctx context.Context, id int64, in dto.UpdateCommentRequest) (dto.CommentResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateComment(ctx, gen.UpdateCommentParams{
		ID:            id,
		CompanyID:     middleware.CompanyFromContext(ctx),
		ContentTypeID: in.ContentTypeID,
		ObjectID:      in.ObjectID,
		Body:          in.Body,
		AuthorID:      in.AuthorID,
		UpdatedAt:     now,
	})
	if err != nil {
		return dto.CommentResponse{}, err
	}
	return toCommentResponse(r), nil
}

func (s *CommentStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteComment(ctx, gen.DeleteCommentParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toCommentResponse(r gen.Comment) dto.CommentResponse {
	return dto.CommentResponse{
		ID:            r.ID,
		CompanyID:     r.CompanyID,
		ContentTypeID: r.ContentTypeID,
		ObjectID:      r.ObjectID,
		Body:          r.Body,
		AuthorID:      r.AuthorID,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
