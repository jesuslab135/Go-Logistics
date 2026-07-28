package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type ServiceTaskPartStore struct{ q *gen.Queries }

func NewServiceTaskPartStore(q *gen.Queries) *ServiceTaskPartStore {
	return &ServiceTaskPartStore{q: q}
}

func (s *ServiceTaskPartStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.ServiceTaskPartResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListServiceTaskParts(ctx, gen.ListServiceTaskPartsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountServiceTaskParts(ctx, gen.CountServiceTaskPartsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ServiceTaskPartResponse, len(rows))
	for i, r := range rows {
		out[i] = toServiceTaskPartResponse(r)
	}
	return out, total, nil
}

func (s *ServiceTaskPartStore) Get(ctx context.Context, parentID, id int64) (dto.ServiceTaskPartResponse, error) {
	r, err := s.q.GetServiceTaskPart(ctx, gen.GetServiceTaskPartParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.ServiceTaskPartResponse{}, err
	}
	return toServiceTaskPartResponse(r), nil
}

func (s *ServiceTaskPartStore) Create(ctx context.Context, parentID int64, in dto.CreateServiceTaskPartRequest) (dto.ServiceTaskPartResponse, error) {
	r, err := s.q.CreateServiceTaskPart(ctx, gen.CreateServiceTaskPartParams{
		ParentID:  parentID,
		CompanyID: middleware.CompanyFromContext(ctx),
		PartID:    in.PartID,
		Quantity:  decimalOrDefault(in.Quantity, 1),
		Position:  in.Position,
	})
	if err != nil {
		return dto.ServiceTaskPartResponse{}, err
	}
	return toServiceTaskPartResponse(r), nil
}

func (s *ServiceTaskPartStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateServiceTaskPartRequest) (dto.ServiceTaskPartResponse, error) {
	r, err := s.q.UpdateServiceTaskPart(ctx, gen.UpdateServiceTaskPartParams{
		ID:        id,
		ParentID:  parentID,
		CompanyID: middleware.CompanyFromContext(ctx),
		PartID:    in.PartID,
		Quantity:  in.Quantity,
		Position:  in.Position,
	})
	if err != nil {
		return dto.ServiceTaskPartResponse{}, err
	}
	return toServiceTaskPartResponse(r), nil
}

func (s *ServiceTaskPartStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteServiceTaskPart(ctx, gen.DeleteServiceTaskPartParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toServiceTaskPartResponse(r gen.ServiceTaskPart) dto.ServiceTaskPartResponse {
	return dto.ServiceTaskPartResponse{
		ID:            r.ID,
		ServiceTaskID: r.ServiceTaskID,
		PartID:        r.PartID,
		Quantity:      r.Quantity,
		Position:      r.Position,
	}
}
