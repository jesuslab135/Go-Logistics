package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
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
	company := middleware.CompanyFromContext(ctx)

	// The parent is verified to exist and to belong to the caller's company
	// before anything is written, so a comment cannot be filed against another
	// tenant's record — or against nothing at all.
	ok, err := s.q.CommentParentExists(ctx, gen.CommentParentExistsParams{
		ContentType:    in.ContentType,
		ObjectID:       in.ObjectID,
		ScopeCompanyID: company,
	})
	if err != nil {
		return dto.CommentResponse{}, err
	}
	if !ok {
		return dto.CommentResponse{}, apierr.NotFound("the comment's parent does not exist")
	}

	r, err := s.q.CreateComment(ctx, gen.CreateCommentParams{
		CompanyID:   company,
		ContentType: in.ContentType,
		ObjectID:    in.ObjectID,
		Body:        in.Body,
		AuthorID:    authorFromContext(ctx),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return dto.CommentResponse{}, err
	}
	return toCommentResponse(r), nil
}

func (s *CommentStore) Update(ctx context.Context, id int64, in dto.UpdateCommentRequest) (dto.CommentResponse, error) {
	r, err := s.q.UpdateComment(ctx, gen.UpdateCommentParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Body:      in.Body,
		UpdatedAt: time.Now().UTC(),
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
		ID:          r.ID,
		CompanyID:   r.CompanyID,
		ContentType: r.ContentType,
		ObjectID:    r.ObjectID,
		Body:        r.Body,
		AuthorID:    r.AuthorID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
