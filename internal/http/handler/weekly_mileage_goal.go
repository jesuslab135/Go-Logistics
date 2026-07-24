package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type WeeklyMileageGoalStore struct{ q *gen.Queries }

func NewWeeklyMileageGoalStore(q *gen.Queries) *WeeklyMileageGoalStore {
	return &WeeklyMileageGoalStore{q: q}
}

func (s *WeeklyMileageGoalStore) List(ctx context.Context, p paginate.Params) ([]dto.WeeklyMileageGoalResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWeeklyMileageGoals(ctx, gen.ListWeeklyMileageGoalsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWeeklyMileageGoals(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.WeeklyMileageGoalResponse, len(rows))
	for i, r := range rows {
		out[i] = toWeeklyMileageGoalResponse(r)
	}
	return out, total, nil
}

func (s *WeeklyMileageGoalStore) Get(ctx context.Context, id int64) (dto.WeeklyMileageGoalResponse, error) {
	r, err := s.q.GetWeeklyMileageGoal(ctx, gen.GetWeeklyMileageGoalParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.WeeklyMileageGoalResponse{}, err
	}
	return toWeeklyMileageGoalResponse(r), nil
}

func (s *WeeklyMileageGoalStore) Create(ctx context.Context, in dto.CreateWeeklyMileageGoalRequest) (dto.WeeklyMileageGoalResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateWeeklyMileageGoal(ctx, gen.CreateWeeklyMileageGoalParams{
		CompanyID:          middleware.CompanyFromContext(ctx),
		ServiceType:        in.ServiceType,
		RatePerMile:        in.RatePerMile,
		WeeklyMileageGoal:  in.WeeklyMileageGoal,
		MpgGoal:            in.MpgGoal,
		UnitsPerService:    in.UnitsPerService,
		MachinesInWorkshop: in.MachinesInWorkshop,
		MissingMiles:       in.MissingMiles,
		SortOrder:          in.SortOrder,
		IsActive:           in.IsActive,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if err != nil {
		return dto.WeeklyMileageGoalResponse{}, err
	}
	return toWeeklyMileageGoalResponse(r), nil
}

func (s *WeeklyMileageGoalStore) Update(ctx context.Context, id int64, in dto.UpdateWeeklyMileageGoalRequest) (dto.WeeklyMileageGoalResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateWeeklyMileageGoal(ctx, gen.UpdateWeeklyMileageGoalParams{
		ID:                 id,
		CompanyID:          middleware.CompanyFromContext(ctx),
		ServiceType:        in.ServiceType,
		RatePerMile:        in.RatePerMile,
		WeeklyMileageGoal:  in.WeeklyMileageGoal,
		MpgGoal:            in.MpgGoal,
		UnitsPerService:    in.UnitsPerService,
		MachinesInWorkshop: in.MachinesInWorkshop,
		MissingMiles:       in.MissingMiles,
		SortOrder:          in.SortOrder,
		IsActive:           in.IsActive,
		UpdatedAt:          now,
	})
	if err != nil {
		return dto.WeeklyMileageGoalResponse{}, err
	}
	return toWeeklyMileageGoalResponse(r), nil
}

func (s *WeeklyMileageGoalStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteWeeklyMileageGoal(ctx, gen.DeleteWeeklyMileageGoalParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toWeeklyMileageGoalResponse(r gen.WeeklyMileageGoal) dto.WeeklyMileageGoalResponse {
	return dto.WeeklyMileageGoalResponse{
		ID:                 r.ID,
		CompanyID:          r.CompanyID,
		ServiceType:        r.ServiceType,
		RatePerMile:        r.RatePerMile,
		WeeklyMileageGoal:  r.WeeklyMileageGoal,
		MpgGoal:            r.MpgGoal,
		UnitsPerService:    r.UnitsPerService,
		MachinesInWorkshop: r.MachinesInWorkshop,
		MissingMiles:       r.MissingMiles,
		SortOrder:          r.SortOrder,
		IsActive:           r.IsActive,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}
