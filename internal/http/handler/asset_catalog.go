package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

// ===================== AssetType =====================

type AssetTypeStore struct{ q *gen.Queries }

func NewAssetTypeStore(q *gen.Queries) *AssetTypeStore { return &AssetTypeStore{q: q} }

func (s *AssetTypeStore) List(ctx context.Context, p paginate.Params) ([]dto.AssetTypeResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListAssetTypes(ctx, gen.ListAssetTypesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountAssetTypes(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.AssetTypeResponse, len(rows))
	for i, r := range rows {
		out[i] = toAssetTypeResponse(r)
	}
	return out, total, nil
}

func (s *AssetTypeStore) Get(ctx context.Context, id int64) (dto.AssetTypeResponse, error) {
	row, err := s.q.GetAssetType(ctx, gen.GetAssetTypeParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.AssetTypeResponse{}, err
	}
	return toAssetTypeResponse(row), nil
}

func (s *AssetTypeStore) Create(ctx context.Context, in dto.CreateAssetTypeRequest) (dto.AssetTypeResponse, error) {
	row, err := s.q.CreateAssetType(ctx, gen.CreateAssetTypeParams{
		CompanyID:   middleware.CompanyFromContext(ctx),
		Name:        in.Name,
		Category:    in.Category,
		Description: in.Description,
	})
	if err != nil {
		return dto.AssetTypeResponse{}, err
	}
	return toAssetTypeResponse(row), nil
}

func (s *AssetTypeStore) Update(ctx context.Context, id int64, in dto.UpdateAssetTypeRequest) (dto.AssetTypeResponse, error) {
	row, err := s.q.UpdateAssetType(ctx, gen.UpdateAssetTypeParams{
		ID:          id,
		CompanyID:   middleware.CompanyFromContext(ctx),
		Name:        in.Name,
		Category:    in.Category,
		Description: in.Description,
	})
	if err != nil {
		return dto.AssetTypeResponse{}, err
	}
	return toAssetTypeResponse(row), nil
}

func (s *AssetTypeStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteAssetType(ctx, gen.DeleteAssetTypeParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toAssetTypeResponse(r gen.AssetType) dto.AssetTypeResponse {
	return dto.AssetTypeResponse{ID: r.ID, CompanyID: r.CompanyID, Name: r.Name, Category: r.Category, Description: r.Description}
}

// ===================== AssetStatus =====================

type AssetStatusStore struct{ q *gen.Queries }

func NewAssetStatusStore(q *gen.Queries) *AssetStatusStore { return &AssetStatusStore{q: q} }

func (s *AssetStatusStore) List(ctx context.Context, p paginate.Params) ([]dto.AssetStatusResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListAssetStatuses(ctx, gen.ListAssetStatusesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountAssetStatuses(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.AssetStatusResponse, len(rows))
	for i, r := range rows {
		out[i] = toAssetStatusResponse(r)
	}
	return out, total, nil
}

func (s *AssetStatusStore) Get(ctx context.Context, id int64) (dto.AssetStatusResponse, error) {
	row, err := s.q.GetAssetStatus(ctx, gen.GetAssetStatusParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.AssetStatusResponse{}, err
	}
	return toAssetStatusResponse(row), nil
}

func (s *AssetStatusStore) Create(ctx context.Context, in dto.CreateAssetStatusRequest) (dto.AssetStatusResponse, error) {
	row, err := s.q.CreateAssetStatus(ctx, gen.CreateAssetStatusParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		ColorCode: orDefault(in.ColorCode, "#28a745"),
	})
	if err != nil {
		return dto.AssetStatusResponse{}, err
	}
	return toAssetStatusResponse(row), nil
}

func (s *AssetStatusStore) Update(ctx context.Context, id int64, in dto.UpdateAssetStatusRequest) (dto.AssetStatusResponse, error) {
	row, err := s.q.UpdateAssetStatus(ctx, gen.UpdateAssetStatusParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		ColorCode: orDefault(in.ColorCode, "#28a745"),
	})
	if err != nil {
		return dto.AssetStatusResponse{}, err
	}
	return toAssetStatusResponse(row), nil
}

func (s *AssetStatusStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteAssetStatus(ctx, gen.DeleteAssetStatusParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toAssetStatusResponse(r gen.AssetStatus) dto.AssetStatusResponse {
	return dto.AssetStatusResponse{ID: r.ID, CompanyID: r.CompanyID, Name: r.Name, ColorCode: r.ColorCode}
}

// ===================== CatalogOption =====================

type CatalogOptionStore struct{ q *gen.Queries }

func NewCatalogOptionStore(q *gen.Queries) *CatalogOptionStore { return &CatalogOptionStore{q: q} }

func (s *CatalogOptionStore) List(ctx context.Context, p paginate.Params) ([]dto.CatalogOptionResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListCatalogOptions(ctx, gen.ListCatalogOptionsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountCatalogOptions(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.CatalogOptionResponse, len(rows))
	for i, r := range rows {
		out[i] = toCatalogOptionResponse(r)
	}
	return out, total, nil
}

func (s *CatalogOptionStore) Get(ctx context.Context, id int64) (dto.CatalogOptionResponse, error) {
	row, err := s.q.GetCatalogOption(ctx, gen.GetCatalogOptionParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.CatalogOptionResponse{}, err
	}
	return toCatalogOptionResponse(row), nil
}

func (s *CatalogOptionStore) Create(ctx context.Context, in dto.CreateCatalogOptionRequest) (dto.CatalogOptionResponse, error) {
	row, err := s.q.CreateCatalogOption(ctx, gen.CreateCatalogOptionParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Category:  in.Category,
		Value:     in.Value,
	})
	if err != nil {
		return dto.CatalogOptionResponse{}, err
	}
	return toCatalogOptionResponse(row), nil
}

func (s *CatalogOptionStore) Update(ctx context.Context, id int64, in dto.UpdateCatalogOptionRequest) (dto.CatalogOptionResponse, error) {
	row, err := s.q.UpdateCatalogOption(ctx, gen.UpdateCatalogOptionParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Category:  in.Category,
		Value:     in.Value,
	})
	if err != nil {
		return dto.CatalogOptionResponse{}, err
	}
	return toCatalogOptionResponse(row), nil
}

func (s *CatalogOptionStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteCatalogOption(ctx, gen.DeleteCatalogOptionParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toCatalogOptionResponse(r gen.CatalogOption) dto.CatalogOptionResponse {
	return dto.CatalogOptionResponse{ID: r.ID, CompanyID: r.CompanyID, Category: r.Category, Value: r.Value}
}
