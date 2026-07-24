package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type LocationStore struct{ q *gen.Queries }

func NewLocationStore(q *gen.Queries) *LocationStore { return &LocationStore{q: q} }

func (s *LocationStore) List(ctx context.Context, p paginate.Params) ([]dto.LocationResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListLocations(ctx, gen.ListLocationsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountLocations(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.LocationResponse, len(rows))
	for i, r := range rows {
		out[i] = toLocationResponse(r)
	}
	return out, total, nil
}

func (s *LocationStore) Get(ctx context.Context, id int64) (dto.LocationResponse, error) {
	r, err := s.q.GetLocation(ctx, gen.GetLocationParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.LocationResponse{}, err
	}
	return toLocationResponse(r), nil
}

func (s *LocationStore) Create(ctx context.Context, in dto.CreateLocationRequest) (dto.LocationResponse, error) {
	r, err := s.q.CreateLocation(ctx, gen.CreateLocationParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		IsActive:  in.IsActive,
	})
	if err != nil {
		return dto.LocationResponse{}, err
	}
	return toLocationResponse(r), nil
}

func (s *LocationStore) Update(ctx context.Context, id int64, in dto.UpdateLocationRequest) (dto.LocationResponse, error) {
	r, err := s.q.UpdateLocation(ctx, gen.UpdateLocationParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		IsActive:  in.IsActive,
	})
	if err != nil {
		return dto.LocationResponse{}, err
	}
	return toLocationResponse(r), nil
}

func (s *LocationStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteLocation(ctx, gen.DeleteLocationParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toLocationResponse(r gen.Location) dto.LocationResponse {
	return dto.LocationResponse{
		ID:        r.ID,
		CompanyID: r.CompanyID,
		Name:      r.Name,
		IsActive:  r.IsActive,
	}
}
