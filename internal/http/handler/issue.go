package handler

import (
	"context"
	"encoding/json"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type IssueStore struct{ q *gen.Queries }

func NewIssueStore(q *gen.Queries) *IssueStore { return &IssueStore{q: q} }

func (s *IssueStore) List(ctx context.Context, p paginate.Params) ([]dto.IssueResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListIssues(ctx, gen.ListIssuesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountIssues(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.IssueResponse, len(rows))
	for i, r := range rows {
		out[i] = toIssueResponse(r)
	}
	return out, total, nil
}

func (s *IssueStore) Get(ctx context.Context, id int64) (dto.IssueResponse, error) {
	r, err := s.q.GetIssue(ctx, gen.GetIssueParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.IssueResponse{}, err
	}
	return toIssueResponse(r), nil
}

func (s *IssueStore) Create(ctx context.Context, in dto.CreateIssueRequest) (dto.IssueResponse, error) {
	// The document has to satisfy whatever this company declared for issues;
	// a company that declared nothing pays one indexed lookup.
	if err := validateCustomFields(ctx, s.q, "issues", in.CustomFields); err != nil {
		return dto.IssueResponse{}, err
	}

	now := time.Now().UTC()
	r, err := s.q.CreateIssue(ctx, gen.CreateIssueParams{
		CompanyID:              middleware.CompanyFromContext(ctx),
		Number:                 in.Number,
		AssetID:                in.AssetID,
		AssetType:              orDefault(in.AssetType, "VEHICLE"),
		Name:                   in.Name,
		Summary:                in.Summary,
		Description:            in.Description,
		State:                  orDefault(in.State, "OPEN"),
		PriorityID:             in.PriorityID,
		FaultID:                in.FaultID,
		SourceType:             orDefault(in.SourceType, "MANUAL"),
		InspectionSubmissionID: in.InspectionSubmissionID,
		ReportedAt:             in.ReportedAt,
		ReportedByID:           in.ReportedByID,
		DueDate:                in.DueDate,
		DueMeterValue:          in.DueMeterValue,
		DueSecondaryMeterValue: in.DueSecondaryMeterValue,
		Overdue:                in.Overdue,
		ResolvedAt:             in.ResolvedAt,
		ResolvedByID:           in.ResolvedByID,
		ResolutionNote:         in.ResolutionNote,
		ReopenedAt:             in.ReopenedAt,
		ReopenedByID:           in.ReopenedByID,
		ResolvableType:         in.ResolvableType,
		ResolvableID:           in.ResolvableID,
		ClosedAt:               in.ClosedAt,
		ClosedByID:             in.ClosedByID,
		ClosedNote:             in.ClosedNote,
		ExternalID:             in.ExternalID,
		CreatedByWorkflow:      in.CreatedByWorkflow,
		CommentsCount:          in.CommentsCount,
		ImagesCount:            in.ImagesCount,
		DocumentsCount:         in.DocumentsCount,
		Labels:                 jsonbOrDefault(in.Labels, "[]"),
		CustomFields:           jsonbOrDefault(in.CustomFields, "{}"),
		CreatedAt:              now,
		UpdatedAt:              now,
	})
	if err != nil {
		return dto.IssueResponse{}, err
	}
	return toIssueResponse(r), nil
}

func (s *IssueStore) Update(ctx context.Context, id int64, in dto.UpdateIssueRequest) (dto.IssueResponse, error) {
	// The document has to satisfy whatever this company declared for issues;
	// a company that declared nothing pays one indexed lookup.
	if err := validateCustomFields(ctx, s.q, "issues", in.CustomFields); err != nil {
		return dto.IssueResponse{}, err
	}

	now := time.Now().UTC()
	r, err := s.q.UpdateIssue(ctx, gen.UpdateIssueParams{
		ID:                     id,
		CompanyID:              middleware.CompanyFromContext(ctx),
		Number:                 in.Number,
		AssetID:                in.AssetID,
		AssetType:              in.AssetType,
		Name:                   in.Name,
		Summary:                in.Summary,
		Description:            in.Description,
		State:                  in.State,
		PriorityID:             in.PriorityID,
		FaultID:                in.FaultID,
		SourceType:             in.SourceType,
		InspectionSubmissionID: in.InspectionSubmissionID,
		ReportedAt:             in.ReportedAt,
		ReportedByID:           in.ReportedByID,
		DueDate:                in.DueDate,
		DueMeterValue:          in.DueMeterValue,
		DueSecondaryMeterValue: in.DueSecondaryMeterValue,
		Overdue:                in.Overdue,
		ResolvedAt:             in.ResolvedAt,
		ResolvedByID:           in.ResolvedByID,
		ResolutionNote:         in.ResolutionNote,
		ReopenedAt:             in.ReopenedAt,
		ReopenedByID:           in.ReopenedByID,
		ResolvableType:         in.ResolvableType,
		ResolvableID:           in.ResolvableID,
		ClosedAt:               in.ClosedAt,
		ClosedByID:             in.ClosedByID,
		ClosedNote:             in.ClosedNote,
		ExternalID:             in.ExternalID,
		CreatedByWorkflow:      in.CreatedByWorkflow,
		CommentsCount:          in.CommentsCount,
		ImagesCount:            in.ImagesCount,
		DocumentsCount:         in.DocumentsCount,
		Labels:                 jsonbOrDefault(in.Labels, "[]"),
		CustomFields:           jsonbOrDefault(in.CustomFields, "{}"),
		UpdatedAt:              now,
	})
	if err != nil {
		return dto.IssueResponse{}, err
	}
	return toIssueResponse(r), nil
}

func (s *IssueStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteIssue(ctx, gen.DeleteIssueParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toIssueResponse(r gen.Issue) dto.IssueResponse {
	return dto.IssueResponse{
		ID:                     r.ID,
		CompanyID:              r.CompanyID,
		Number:                 r.Number,
		AssetID:                r.AssetID,
		AssetType:              r.AssetType,
		Name:                   r.Name,
		Summary:                r.Summary,
		Description:            r.Description,
		State:                  r.State,
		PriorityID:             r.PriorityID,
		FaultID:                r.FaultID,
		SourceType:             r.SourceType,
		InspectionSubmissionID: r.InspectionSubmissionID,
		ReportedAt:             r.ReportedAt,
		ReportedByID:           r.ReportedByID,
		DueDate:                r.DueDate,
		DueMeterValue:          r.DueMeterValue,
		DueSecondaryMeterValue: r.DueSecondaryMeterValue,
		Overdue:                r.Overdue,
		ResolvedAt:             r.ResolvedAt,
		ResolvedByID:           r.ResolvedByID,
		ResolutionNote:         r.ResolutionNote,
		ReopenedAt:             r.ReopenedAt,
		ReopenedByID:           r.ReopenedByID,
		ResolvableType:         r.ResolvableType,
		ResolvableID:           r.ResolvableID,
		ClosedAt:               r.ClosedAt,
		ClosedByID:             r.ClosedByID,
		ClosedNote:             r.ClosedNote,
		ExternalID:             r.ExternalID,
		CreatedByWorkflow:      r.CreatedByWorkflow,
		CommentsCount:          r.CommentsCount,
		ImagesCount:            r.ImagesCount,
		DocumentsCount:         r.DocumentsCount,
		Labels:                 json.RawMessage(r.Labels),
		CustomFields:           json.RawMessage(r.CustomFields),
		CreatedAt:              r.CreatedAt,
		UpdatedAt:              r.UpdatedAt,
	}
}
