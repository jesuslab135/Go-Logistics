package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type AxleDefinitionStore struct{ q *gen.Queries }

func NewAxleDefinitionStore(q *gen.Queries) *AxleDefinitionStore { return &AxleDefinitionStore{q: q} }

func (s *AxleDefinitionStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.AxleDefinitionResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListAxleDefinitions(ctx, gen.ListAxleDefinitionsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountAxleDefinitions(ctx, gen.CountAxleDefinitionsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.AxleDefinitionResponse, len(rows))
	for i, r := range rows {
		out[i] = toAxleDefinitionResponse(r)
	}
	return out, total, nil
}

func (s *AxleDefinitionStore) Get(ctx context.Context, parentID, id int64) (dto.AxleDefinitionResponse, error) {
	r, err := s.q.GetAxleDefinition(ctx, gen.GetAxleDefinitionParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.AxleDefinitionResponse{}, err
	}
	return toAxleDefinitionResponse(r), nil
}

func (s *AxleDefinitionStore) Create(ctx context.Context, parentID int64, in dto.CreateAxleDefinitionRequest) (dto.AxleDefinitionResponse, error) {
	r, err := s.q.CreateAxleDefinition(ctx, gen.CreateAxleDefinitionParams{
		ParentID:         parentID,
		CompanyID:        middleware.CompanyFromContext(ctx),
		PositionIndex:    in.PositionIndex,
		Label:            in.Label,
		AxleRole:         in.AxleRole,
		PositionsPerSide: in.PositionsPerSide,
	})
	if err != nil {
		return dto.AxleDefinitionResponse{}, err
	}
	return toAxleDefinitionResponse(r), nil
}

func (s *AxleDefinitionStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateAxleDefinitionRequest) (dto.AxleDefinitionResponse, error) {
	r, err := s.q.UpdateAxleDefinition(ctx, gen.UpdateAxleDefinitionParams{
		ID:               id,
		ParentID:         parentID,
		CompanyID:        middleware.CompanyFromContext(ctx),
		PositionIndex:    in.PositionIndex,
		Label:            in.Label,
		AxleRole:         in.AxleRole,
		PositionsPerSide: in.PositionsPerSide,
	})
	if err != nil {
		return dto.AxleDefinitionResponse{}, err
	}
	return toAxleDefinitionResponse(r), nil
}

func (s *AxleDefinitionStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteAxleDefinition(ctx, gen.DeleteAxleDefinitionParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toAxleDefinitionResponse(r gen.AxleDefinition) dto.AxleDefinitionResponse {
	return dto.AxleDefinitionResponse{
		ID:               r.ID,
		TemplateID:       r.TemplateID,
		PositionIndex:    r.PositionIndex,
		Label:            r.Label,
		AxleRole:         r.AxleRole,
		PositionsPerSide: r.PositionsPerSide,
	}
}
