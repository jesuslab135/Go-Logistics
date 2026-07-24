package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type ServiceReminderStore struct{ q *gen.Queries }

func NewServiceReminderStore(q *gen.Queries) *ServiceReminderStore {
	return &ServiceReminderStore{q: q}
}

func (s *ServiceReminderStore) List(ctx context.Context, p paginate.Params) ([]dto.ServiceReminderResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListServiceReminders(ctx, gen.ListServiceRemindersParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountServiceReminders(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ServiceReminderResponse, len(rows))
	for i, r := range rows {
		out[i] = toServiceReminderResponse(r)
	}
	return out, total, nil
}

func (s *ServiceReminderStore) Get(ctx context.Context, id int64) (dto.ServiceReminderResponse, error) {
	r, err := s.q.GetServiceReminder(ctx, gen.GetServiceReminderParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.ServiceReminderResponse{}, err
	}
	return toServiceReminderResponse(r), nil
}

func (s *ServiceReminderStore) Create(ctx context.Context, in dto.CreateServiceReminderRequest) (dto.ServiceReminderResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateServiceReminder(ctx, gen.CreateServiceReminderParams{
		CompanyID:             middleware.CompanyFromContext(ctx),
		AssetID:               in.AssetID,
		ServiceTaskID:         in.ServiceTaskID,
		IsActive:              in.IsActive,
		Status:                in.Status,
		TimeInterval:          in.TimeInterval,
		TimeFrequency:         in.TimeFrequency,
		NextDueAt:             in.NextDueAt,
		DueSoonAt:             in.DueSoonAt,
		DueSoonTimeThreshold:  in.DueSoonTimeThreshold,
		MeterInterval:         in.MeterInterval,
		NextDueMeterValue:     in.NextDueMeterValue,
		DueSoonMeterValue:     in.DueSoonMeterValue,
		DueSoonMeterThreshold: in.DueSoonMeterThreshold,
		SnoozeUntil:           in.SnoozeUntil,
		LastServiceEntryID:    in.LastServiceEntryID,
		CreatedAt:             now,
		UpdatedAt:             now,
	})
	if err != nil {
		return dto.ServiceReminderResponse{}, err
	}
	return toServiceReminderResponse(r), nil
}

func (s *ServiceReminderStore) Update(ctx context.Context, id int64, in dto.UpdateServiceReminderRequest) (dto.ServiceReminderResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateServiceReminder(ctx, gen.UpdateServiceReminderParams{
		ID:                    id,
		CompanyID:             middleware.CompanyFromContext(ctx),
		AssetID:               in.AssetID,
		ServiceTaskID:         in.ServiceTaskID,
		IsActive:              in.IsActive,
		Status:                in.Status,
		TimeInterval:          in.TimeInterval,
		TimeFrequency:         in.TimeFrequency,
		NextDueAt:             in.NextDueAt,
		DueSoonAt:             in.DueSoonAt,
		DueSoonTimeThreshold:  in.DueSoonTimeThreshold,
		MeterInterval:         in.MeterInterval,
		NextDueMeterValue:     in.NextDueMeterValue,
		DueSoonMeterValue:     in.DueSoonMeterValue,
		DueSoonMeterThreshold: in.DueSoonMeterThreshold,
		SnoozeUntil:           in.SnoozeUntil,
		LastServiceEntryID:    in.LastServiceEntryID,
		UpdatedAt:             now,
	})
	if err != nil {
		return dto.ServiceReminderResponse{}, err
	}
	return toServiceReminderResponse(r), nil
}

func (s *ServiceReminderStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteServiceReminder(ctx, gen.DeleteServiceReminderParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toServiceReminderResponse(r gen.ServiceReminder) dto.ServiceReminderResponse {
	return dto.ServiceReminderResponse{
		ID:                    r.ID,
		CompanyID:             r.CompanyID,
		AssetID:               r.AssetID,
		ServiceTaskID:         r.ServiceTaskID,
		IsActive:              r.IsActive,
		Status:                r.Status,
		TimeInterval:          r.TimeInterval,
		TimeFrequency:         r.TimeFrequency,
		NextDueAt:             r.NextDueAt,
		DueSoonAt:             r.DueSoonAt,
		DueSoonTimeThreshold:  r.DueSoonTimeThreshold,
		MeterInterval:         r.MeterInterval,
		NextDueMeterValue:     r.NextDueMeterValue,
		DueSoonMeterValue:     r.DueSoonMeterValue,
		DueSoonMeterThreshold: r.DueSoonMeterThreshold,
		SnoozeUntil:           r.SnoozeUntil,
		LastServiceEntryID:    r.LastServiceEntryID,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}
