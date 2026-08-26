package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

// TrailerClassificationStore is the company's own vocabulary for classifying
// trailers, following the fuel_type / measurement_unit pattern every other
// catalog here uses.
//
// It replaces two free-text columns. Nothing is seeded: the migration filled
// each company's catalog from the values that company was already using, and
// seeding a guessed list on top of that would put words in ops' mouths.
type TrailerClassificationStore struct{ q *gen.Queries }

func NewTrailerClassificationStore(q *gen.Queries) *TrailerClassificationStore {
	return &TrailerClassificationStore{q: q}
}

func (s *TrailerClassificationStore) List(ctx context.Context, p paginate.Params) ([]dto.TrailerClassificationResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListTrailerClassifications(ctx, gen.ListTrailerClassificationsParams{
		CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountTrailerClassifications(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.TrailerClassificationResponse, len(rows))
	for i, r := range rows {
		out[i] = toTrailerClassificationResponse(r)
	}
	return out, total, nil
}

func (s *TrailerClassificationStore) Get(ctx context.Context, id int64) (dto.TrailerClassificationResponse, error) {
	r, err := s.q.GetTrailerClassification(ctx, gen.GetTrailerClassificationParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.TrailerClassificationResponse{}, err
	}
	return toTrailerClassificationResponse(r), nil
}

func (s *TrailerClassificationStore) Create(ctx context.Context, in dto.CreateTrailerClassificationRequest) (dto.TrailerClassificationResponse, error) {
	r, err := s.q.CreateTrailerClassification(ctx, gen.CreateTrailerClassificationParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		Position:  in.Position,
	})
	if err != nil {
		return dto.TrailerClassificationResponse{}, err
	}
	return toTrailerClassificationResponse(r), nil
}

func (s *TrailerClassificationStore) Update(ctx context.Context, id int64, in dto.UpdateTrailerClassificationRequest) (dto.TrailerClassificationResponse, error) {
	r, err := s.q.UpdateTrailerClassification(ctx, gen.UpdateTrailerClassificationParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		Position:  in.Position,
	})
	if err != nil {
		return dto.TrailerClassificationResponse{}, err
	}
	return toTrailerClassificationResponse(r), nil
}

// Delete removes a classification. Trailers referencing it keep their row: the
// FK is ON DELETE SET NULL, so they become unclassified rather than
// undeletable. Retiring a term somebody stopped using should not require
// re-classifying every trailer that ever carried it.
func (s *TrailerClassificationStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteTrailerClassification(ctx, gen.DeleteTrailerClassificationParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toTrailerClassificationResponse(r gen.TrailerClassification) dto.TrailerClassificationResponse {
	return dto.TrailerClassificationResponse{
		ID:        r.ID,
		CompanyID: r.CompanyID,
		Name:      r.Name,
		Position:  r.Position,
	}
}
