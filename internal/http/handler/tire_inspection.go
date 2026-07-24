package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type TireInspectionStore struct{ q *gen.Queries }

func NewTireInspectionStore(q *gen.Queries) *TireInspectionStore { return &TireInspectionStore{q: q} }

func (s *TireInspectionStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.TireInspectionResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListTireInspections(ctx, gen.ListTireInspectionsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountTireInspections(ctx, gen.CountTireInspectionsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.TireInspectionResponse, len(rows))
	for i, r := range rows {
		out[i] = toTireInspectionResponse(r)
	}
	return out, total, nil
}

func (s *TireInspectionStore) Get(ctx context.Context, parentID, id int64) (dto.TireInspectionResponse, error) {
	r, err := s.q.GetTireInspection(ctx, gen.GetTireInspectionParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.TireInspectionResponse{}, err
	}
	return toTireInspectionResponse(r), nil
}

func (s *TireInspectionStore) Create(ctx context.Context, parentID int64, in dto.CreateTireInspectionRequest) (dto.TireInspectionResponse, error) {
	r, err := s.q.CreateTireInspection(ctx, gen.CreateTireInspectionParams{
		ParentID:        parentID,
		CompanyID:       middleware.CompanyFromContext(ctx),
		InspectionDate:  in.InspectionDate,
		Odometer:        in.Odometer,
		TreadDepth32nds: in.TreadDepth32nds,
		Psi:             in.Psi,
		MeasuredByID:    in.MeasuredByID,
		Notes:           in.Notes,
	})
	if err != nil {
		return dto.TireInspectionResponse{}, err
	}
	return toTireInspectionResponse(r), nil
}

func (s *TireInspectionStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateTireInspectionRequest) (dto.TireInspectionResponse, error) {
	r, err := s.q.UpdateTireInspection(ctx, gen.UpdateTireInspectionParams{
		ID:              id,
		ParentID:        parentID,
		CompanyID:       middleware.CompanyFromContext(ctx),
		InspectionDate:  in.InspectionDate,
		Odometer:        in.Odometer,
		TreadDepth32nds: in.TreadDepth32nds,
		Psi:             in.Psi,
		MeasuredByID:    in.MeasuredByID,
		Notes:           in.Notes,
	})
	if err != nil {
		return dto.TireInspectionResponse{}, err
	}
	return toTireInspectionResponse(r), nil
}

func (s *TireInspectionStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteTireInspection(ctx, gen.DeleteTireInspectionParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toTireInspectionResponse(r gen.TireInspection) dto.TireInspectionResponse {
	return dto.TireInspectionResponse{
		ID:              r.ID,
		TireID:          r.TireID,
		InspectionDate:  r.InspectionDate,
		Odometer:        r.Odometer,
		TreadDepth32nds: r.TreadDepth32nds,
		Psi:             r.Psi,
		MeasuredByID:    r.MeasuredByID,
		Notes:           r.Notes,
	}
}
