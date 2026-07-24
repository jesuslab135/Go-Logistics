package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type PartManufacturerStore struct{ q *gen.Queries }

func NewPartManufacturerStore(q *gen.Queries) *PartManufacturerStore {
	return &PartManufacturerStore{q: q}
}

func (s *PartManufacturerStore) List(ctx context.Context, p paginate.Params) ([]dto.PartManufacturerResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListPartManufacturers(ctx, gen.ListPartManufacturersParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountPartManufacturers(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PartManufacturerResponse, len(rows))
	for i, r := range rows {
		out[i] = toPartManufacturerResponse(r)
	}
	return out, total, nil
}

func (s *PartManufacturerStore) Get(ctx context.Context, id int64) (dto.PartManufacturerResponse, error) {
	r, err := s.q.GetPartManufacturer(ctx, gen.GetPartManufacturerParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.PartManufacturerResponse{}, err
	}
	return toPartManufacturerResponse(r), nil
}

func (s *PartManufacturerStore) Create(ctx context.Context, in dto.CreatePartManufacturerRequest) (dto.PartManufacturerResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreatePartManufacturer(ctx, gen.CreatePartManufacturerParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		Website:   in.Website,
		CreatedAt: now,
	})
	if err != nil {
		return dto.PartManufacturerResponse{}, err
	}
	return toPartManufacturerResponse(r), nil
}

func (s *PartManufacturerStore) Update(ctx context.Context, id int64, in dto.UpdatePartManufacturerRequest) (dto.PartManufacturerResponse, error) {
	r, err := s.q.UpdatePartManufacturer(ctx, gen.UpdatePartManufacturerParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		Website:   in.Website,
	})
	if err != nil {
		return dto.PartManufacturerResponse{}, err
	}
	return toPartManufacturerResponse(r), nil
}

func (s *PartManufacturerStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeletePartManufacturer(ctx, gen.DeletePartManufacturerParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toPartManufacturerResponse(r gen.PartManufacturer) dto.PartManufacturerResponse {
	return dto.PartManufacturerResponse{
		ID:        r.ID,
		CompanyID: r.CompanyID,
		Name:      r.Name,
		Website:   r.Website,
		CreatedAt: r.CreatedAt,
	}
}
