package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type WheelPositionDefinitionStore struct{ q *gen.Queries }

func NewWheelPositionDefinitionStore(q *gen.Queries) *WheelPositionDefinitionStore {
	return &WheelPositionDefinitionStore{q: q}
}

func (s *WheelPositionDefinitionStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.WheelPositionDefinitionResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWheelPositionDefinitions(ctx, gen.ListWheelPositionDefinitionsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWheelPositionDefinitions(ctx, gen.CountWheelPositionDefinitionsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.WheelPositionDefinitionResponse, len(rows))
	for i, r := range rows {
		out[i] = toWheelPositionDefinitionResponse(r)
	}
	return out, total, nil
}

func (s *WheelPositionDefinitionStore) Get(ctx context.Context, parentID, id int64) (dto.WheelPositionDefinitionResponse, error) {
	r, err := s.q.GetWheelPositionDefinition(ctx, gen.GetWheelPositionDefinitionParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.WheelPositionDefinitionResponse{}, err
	}
	return toWheelPositionDefinitionResponse(r), nil
}

func (s *WheelPositionDefinitionStore) Create(ctx context.Context, parentID int64, in dto.CreateWheelPositionDefinitionRequest) (dto.WheelPositionDefinitionResponse, error) {
	r, err := s.q.CreateWheelPositionDefinition(ctx, gen.CreateWheelPositionDefinitionParams{
		ParentID:  parentID,
		CompanyID: middleware.CompanyFromContext(ctx),
		Code:      in.Code,
		Side:      in.Side,
		Slot:      int32OrDefault(in.Slot, 1),
	})
	if err != nil {
		return dto.WheelPositionDefinitionResponse{}, err
	}
	return toWheelPositionDefinitionResponse(r), nil
}

func (s *WheelPositionDefinitionStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateWheelPositionDefinitionRequest) (dto.WheelPositionDefinitionResponse, error) {
	r, err := s.q.UpdateWheelPositionDefinition(ctx, gen.UpdateWheelPositionDefinitionParams{
		ID:        id,
		ParentID:  parentID,
		CompanyID: middleware.CompanyFromContext(ctx),
		Code:      in.Code,
		Side:      in.Side,
		Slot:      in.Slot,
	})
	if err != nil {
		return dto.WheelPositionDefinitionResponse{}, err
	}
	return toWheelPositionDefinitionResponse(r), nil
}

func (s *WheelPositionDefinitionStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteWheelPositionDefinition(ctx, gen.DeleteWheelPositionDefinitionParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toWheelPositionDefinitionResponse(r gen.WheelPositionDefinition) dto.WheelPositionDefinitionResponse {
	return dto.WheelPositionDefinitionResponse{
		ID:     r.ID,
		AxleID: r.AxleID,
		Code:   r.Code,
		Side:   r.Side,
		Slot:   r.Slot,
	}
}
