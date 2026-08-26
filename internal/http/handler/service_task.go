package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type ServiceTaskStore struct{ q *gen.Queries }

func NewServiceTaskStore(q *gen.Queries) *ServiceTaskStore { return &ServiceTaskStore{q: q} }

func (s *ServiceTaskStore) List(ctx context.Context, p paginate.Params) ([]dto.ServiceTaskResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListServiceTasks(ctx, gen.ListServiceTasksParams{CompanyID: company, IncludeArchived: includeArchived(ctx), Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountServiceTasks(ctx, gen.CountServiceTasksParams{CompanyID: company, IncludeArchived: includeArchived(ctx)})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ServiceTaskResponse, len(rows))
	for i, r := range rows {
		out[i] = toServiceTaskResponse(r)
	}
	return out, total, nil
}

func (s *ServiceTaskStore) Get(ctx context.Context, id int64) (dto.ServiceTaskResponse, error) {
	r, err := s.q.GetServiceTask(ctx, gen.GetServiceTaskParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.ServiceTaskResponse{}, err
	}
	return toServiceTaskResponse(r), nil
}

func (s *ServiceTaskStore) Create(ctx context.Context, in dto.CreateServiceTaskRequest) (dto.ServiceTaskResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateServiceTask(ctx, gen.CreateServiceTaskParams{
		CompanyID:               middleware.CompanyFromContext(ctx),
		Name:                    in.Name,
		Description:             in.Description,
		ExpectedDurationSeconds: in.ExpectedDurationSeconds,
		ParentTaskID:            in.ParentTaskID,
		CreatedAt:               now,
		UpdatedAt:               now,
	})
	if err != nil {
		return dto.ServiceTaskResponse{}, err
	}
	return toServiceTaskResponse(r), nil
}

func (s *ServiceTaskStore) Update(ctx context.Context, id int64, in dto.UpdateServiceTaskRequest) (dto.ServiceTaskResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateServiceTask(ctx, gen.UpdateServiceTaskParams{
		ID:                      id,
		CompanyID:               middleware.CompanyFromContext(ctx),
		Name:                    in.Name,
		Description:             in.Description,
		ExpectedDurationSeconds: in.ExpectedDurationSeconds,
		ParentTaskID:            in.ParentTaskID,
		UpdatedAt:               now,
	})
	if err != nil {
		return dto.ServiceTaskResponse{}, err
	}
	return toServiceTaskResponse(r), nil
}

func (s *ServiceTaskStore) Delete(ctx context.Context, id int64) error {
	if err := guardDelete(ctx, "service task", id, s.q.ServiceTaskReferences); err != nil {
		return err
	}
	return s.q.DeleteServiceTask(ctx, gen.DeleteServiceTaskParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toServiceTaskResponse(r gen.ServiceTask) dto.ServiceTaskResponse {
	return dto.ServiceTaskResponse{
		ID:                      r.ID,
		CompanyID:               r.CompanyID,
		Name:                    r.Name,
		Description:             r.Description,
		ExpectedDurationSeconds: r.ExpectedDurationSeconds,
		ParentTaskID:            r.ParentTaskID,
		ArchivedAt:              r.ArchivedAt,
		CreatedAt:               r.CreatedAt,
		UpdatedAt:               r.UpdatedAt,
	}
}
