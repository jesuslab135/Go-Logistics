package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type TireMountLogStore struct{ q *gen.Queries }

func NewTireMountLogStore(q *gen.Queries) *TireMountLogStore { return &TireMountLogStore{q: q} }

func (s *TireMountLogStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.TireMountLogResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListTireMountLogs(ctx, gen.ListTireMountLogsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountTireMountLogs(ctx, gen.CountTireMountLogsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.TireMountLogResponse, len(rows))
	for i, r := range rows {
		out[i] = toTireMountLogResponse(r)
	}
	return out, total, nil
}

func (s *TireMountLogStore) Get(ctx context.Context, parentID, id int64) (dto.TireMountLogResponse, error) {
	r, err := s.q.GetTireMountLog(ctx, gen.GetTireMountLogParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.TireMountLogResponse{}, err
	}
	return toTireMountLogResponse(r), nil
}

func (s *TireMountLogStore) Create(ctx context.Context, parentID int64, in dto.CreateTireMountLogRequest) (dto.TireMountLogResponse, error) {
	r, err := s.q.CreateTireMountLog(ctx, gen.CreateTireMountLogParams{
		ParentID:        parentID,
		CompanyID:       middleware.CompanyFromContext(ctx),
		VehicleID:       in.VehicleID,
		PositionCode:    in.PositionCode,
		EventType:       in.EventType,
		EventDate:       in.EventDate,
		Odometer:        in.Odometer,
		TreadDepth32nds: in.TreadDepth32nds,
		Psi:             in.Psi,
		PerformedByID:   in.PerformedByID,
		Reason:          in.Reason,
	})
	if err != nil {
		return dto.TireMountLogResponse{}, err
	}
	return toTireMountLogResponse(r), nil
}

func (s *TireMountLogStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateTireMountLogRequest) (dto.TireMountLogResponse, error) {
	r, err := s.q.UpdateTireMountLog(ctx, gen.UpdateTireMountLogParams{
		ID:              id,
		ParentID:        parentID,
		CompanyID:       middleware.CompanyFromContext(ctx),
		VehicleID:       in.VehicleID,
		PositionCode:    in.PositionCode,
		EventType:       in.EventType,
		EventDate:       in.EventDate,
		Odometer:        in.Odometer,
		TreadDepth32nds: in.TreadDepth32nds,
		Psi:             in.Psi,
		PerformedByID:   in.PerformedByID,
		Reason:          in.Reason,
	})
	if err != nil {
		return dto.TireMountLogResponse{}, err
	}
	return toTireMountLogResponse(r), nil
}

func (s *TireMountLogStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteTireMountLog(ctx, gen.DeleteTireMountLogParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toTireMountLogResponse(r gen.TireMountLog) dto.TireMountLogResponse {
	return dto.TireMountLogResponse{
		ID:              r.ID,
		TireID:          r.TireID,
		VehicleID:       r.VehicleID,
		PositionCode:    r.PositionCode,
		EventType:       r.EventType,
		EventDate:       r.EventDate,
		Odometer:        r.Odometer,
		TreadDepth32nds: r.TreadDepth32nds,
		Psi:             r.Psi,
		PerformedByID:   r.PerformedByID,
		Reason:          r.Reason,
	}
}
