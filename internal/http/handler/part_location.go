package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type PartLocationStore struct{ q *gen.Queries }

func NewPartLocationStore(q *gen.Queries) *PartLocationStore { return &PartLocationStore{q: q} }

func (s *PartLocationStore) List(ctx context.Context, p paginate.Params) ([]dto.PartLocationResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListPartLocations(ctx, gen.ListPartLocationsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountPartLocations(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PartLocationResponse, len(rows))
	for i, r := range rows {
		out[i] = toPartLocationResponse(r)
	}
	return out, total, nil
}

func (s *PartLocationStore) Get(ctx context.Context, id int64) (dto.PartLocationResponse, error) {
	r, err := s.q.GetPartLocation(ctx, gen.GetPartLocationParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.PartLocationResponse{}, err
	}
	return toPartLocationResponse(r), nil
}

func (s *PartLocationStore) Create(ctx context.Context, in dto.CreatePartLocationRequest) (dto.PartLocationResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreatePartLocation(ctx, gen.CreatePartLocationParams{
		CompanyID:  middleware.CompanyFromContext(ctx),
		Name:       in.Name,
		Address:    in.Address,
		City:       in.City,
		Region:     in.Region,
		LocationID: in.LocationID,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		return dto.PartLocationResponse{}, err
	}
	return toPartLocationResponse(r), nil
}

func (s *PartLocationStore) Update(ctx context.Context, id int64, in dto.UpdatePartLocationRequest) (dto.PartLocationResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdatePartLocation(ctx, gen.UpdatePartLocationParams{
		ID:         id,
		CompanyID:  middleware.CompanyFromContext(ctx),
		Name:       in.Name,
		Address:    in.Address,
		City:       in.City,
		Region:     in.Region,
		LocationID: in.LocationID,
		UpdatedAt:  now,
	})
	if err != nil {
		return dto.PartLocationResponse{}, err
	}
	return toPartLocationResponse(r), nil
}

func (s *PartLocationStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeletePartLocation(ctx, gen.DeletePartLocationParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toPartLocationResponse(r gen.PartLocation) dto.PartLocationResponse {
	return dto.PartLocationResponse{
		ID:         r.ID,
		CompanyID:  r.CompanyID,
		Name:       r.Name,
		Address:    r.Address,
		City:       r.City,
		Region:     r.Region,
		LocationID: r.LocationID,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}
