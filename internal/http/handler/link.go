package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

// Stores for the m2m join tables. Every Link starts by resolving the target
// through a "candidate" query that also proves the parent belongs to the
// caller's company; when it finds nothing, pgx.ErrNoRows surfaces as 404 and no
// write happens. That keeps cross-tenant linking indistinguishable from a
// missing record, as it is everywhere else in the API.

// --------------------------------------------------------------------------
// issue -> assignees
// --------------------------------------------------------------------------

type IssueAssigneeStore struct{ q *gen.Queries }

func NewIssueAssigneeStore(q *gen.Queries) *IssueAssigneeStore { return &IssueAssigneeStore{q: q} }

func (s *IssueAssigneeStore) List(ctx context.Context, issueID int64, p paginate.Params) ([]dto.EmployeeResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListIssueAssignees(ctx, gen.ListIssueAssigneesParams{
		IssueID: issueID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountIssueAssignees(ctx, gen.CountIssueAssigneesParams{IssueID: issueID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.EmployeeResponse, len(rows))
	for i, r := range rows {
		out[i] = toEmployeeResponse(r)
	}
	return out, total, nil
}

func (s *IssueAssigneeStore) Link(ctx context.Context, issueID int64, in dto.LinkEmployeeRequest) (dto.EmployeeResponse, error) {
	company := middleware.CompanyFromContext(ctx)
	employee, err := s.q.FindIssueAssigneeCandidate(ctx, gen.FindIssueAssigneeCandidateParams{
		EmployeeID: in.EmployeeID, CompanyID: company, IssueID: issueID,
	})
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	if err := s.q.AddIssueAssignee(ctx, gen.AddIssueAssigneeParams{IssueID: issueID, EmployeeID: employee.ID}); err != nil {
		return dto.EmployeeResponse{}, err
	}
	return toEmployeeResponse(employee), nil
}

func (s *IssueAssigneeStore) Unlink(ctx context.Context, issueID, employeeID int64) error {
	return s.q.RemoveIssueAssignee(ctx, gen.RemoveIssueAssigneeParams{
		IssueID: issueID, EmployeeID: employeeID, CompanyID: middleware.CompanyFromContext(ctx),
	})
}

// --------------------------------------------------------------------------
// issue -> watchers
// --------------------------------------------------------------------------

type IssueWatcherStore struct{ q *gen.Queries }

func NewIssueWatcherStore(q *gen.Queries) *IssueWatcherStore { return &IssueWatcherStore{q: q} }

func (s *IssueWatcherStore) List(ctx context.Context, issueID int64, p paginate.Params) ([]dto.EmployeeResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListIssueWatchers(ctx, gen.ListIssueWatchersParams{
		IssueID: issueID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountIssueWatchers(ctx, gen.CountIssueWatchersParams{IssueID: issueID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.EmployeeResponse, len(rows))
	for i, r := range rows {
		out[i] = toEmployeeResponse(r)
	}
	return out, total, nil
}

func (s *IssueWatcherStore) Link(ctx context.Context, issueID int64, in dto.LinkEmployeeRequest) (dto.EmployeeResponse, error) {
	company := middleware.CompanyFromContext(ctx)
	employee, err := s.q.FindIssueWatcherCandidate(ctx, gen.FindIssueWatcherCandidateParams{
		EmployeeID: in.EmployeeID, CompanyID: company, IssueID: issueID,
	})
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	if err := s.q.AddIssueWatcher(ctx, gen.AddIssueWatcherParams{IssueID: issueID, EmployeeID: employee.ID}); err != nil {
		return dto.EmployeeResponse{}, err
	}
	return toEmployeeResponse(employee), nil
}

func (s *IssueWatcherStore) Unlink(ctx context.Context, issueID, employeeID int64) error {
	return s.q.RemoveIssueWatcher(ctx, gen.RemoveIssueWatcherParams{
		IssueID: issueID, EmployeeID: employeeID, CompanyID: middleware.CompanyFromContext(ctx),
	})
}

// --------------------------------------------------------------------------
// work order -> issues
// --------------------------------------------------------------------------

type WorkOrderIssueStore struct{ q *gen.Queries }

func NewWorkOrderIssueStore(q *gen.Queries) *WorkOrderIssueStore { return &WorkOrderIssueStore{q: q} }

func (s *WorkOrderIssueStore) List(ctx context.Context, workOrderID int64, p paginate.Params) ([]dto.IssueResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWorkOrderIssues(ctx, gen.ListWorkOrderIssuesParams{
		WorkOrderID: workOrderID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWorkOrderIssues(ctx, gen.CountWorkOrderIssuesParams{WorkOrderID: workOrderID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.IssueResponse, len(rows))
	for i, r := range rows {
		out[i] = toIssueResponse(r)
	}
	return out, total, nil
}

func (s *WorkOrderIssueStore) Link(ctx context.Context, workOrderID int64, in dto.LinkIssueRequest) (dto.IssueResponse, error) {
	company := middleware.CompanyFromContext(ctx)
	issue, err := s.q.FindWorkOrderIssueCandidate(ctx, gen.FindWorkOrderIssueCandidateParams{
		IssueID: in.IssueID, CompanyID: company, WorkOrderID: workOrderID,
	})
	if err != nil {
		return dto.IssueResponse{}, err
	}
	if err := s.q.AddWorkOrderIssue(ctx, gen.AddWorkOrderIssueParams{WorkOrderID: workOrderID, IssueID: issue.ID}); err != nil {
		return dto.IssueResponse{}, err
	}
	return toIssueResponse(issue), nil
}

func (s *WorkOrderIssueStore) Unlink(ctx context.Context, workOrderID, issueID int64) error {
	return s.q.RemoveWorkOrderIssue(ctx, gen.RemoveWorkOrderIssueParams{
		WorkOrderID: workOrderID, IssueID: issueID, CompanyID: middleware.CompanyFromContext(ctx),
	})
}

// --------------------------------------------------------------------------
// work order -> faults
// --------------------------------------------------------------------------

type WorkOrderFaultStore struct{ q *gen.Queries }

func NewWorkOrderFaultStore(q *gen.Queries) *WorkOrderFaultStore { return &WorkOrderFaultStore{q: q} }

func (s *WorkOrderFaultStore) List(ctx context.Context, workOrderID int64, p paginate.Params) ([]dto.FaultResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWorkOrderFaults(ctx, gen.ListWorkOrderFaultsParams{
		WorkOrderID: workOrderID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWorkOrderFaults(ctx, gen.CountWorkOrderFaultsParams{WorkOrderID: workOrderID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.FaultResponse, len(rows))
	for i, r := range rows {
		out[i] = toFaultResponse(r)
	}
	return out, total, nil
}

func (s *WorkOrderFaultStore) Link(ctx context.Context, workOrderID int64, in dto.LinkFaultRequest) (dto.FaultResponse, error) {
	company := middleware.CompanyFromContext(ctx)
	fault, err := s.q.FindWorkOrderFaultCandidate(ctx, gen.FindWorkOrderFaultCandidateParams{
		FaultID: in.FaultID, CompanyID: company, WorkOrderID: workOrderID,
	})
	if err != nil {
		return dto.FaultResponse{}, err
	}
	if err := s.q.AddWorkOrderFault(ctx, gen.AddWorkOrderFaultParams{WorkOrderID: workOrderID, FaultID: fault.ID}); err != nil {
		return dto.FaultResponse{}, err
	}
	return toFaultResponse(fault), nil
}

func (s *WorkOrderFaultStore) Unlink(ctx context.Context, workOrderID, faultID int64) error {
	return s.q.RemoveWorkOrderFault(ctx, gen.RemoveWorkOrderFaultParams{
		WorkOrderID: workOrderID, FaultID: faultID, CompanyID: middleware.CompanyFromContext(ctx),
	})
}

// --------------------------------------------------------------------------
// service entry line item -> issues
// --------------------------------------------------------------------------

type ServiceEntryLineItemIssueStore struct{ q *gen.Queries }

func NewServiceEntryLineItemIssueStore(q *gen.Queries) *ServiceEntryLineItemIssueStore {
	return &ServiceEntryLineItemIssueStore{q: q}
}

func (s *ServiceEntryLineItemIssueStore) List(ctx context.Context, lineItemID int64, p paginate.Params) ([]dto.IssueResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListServiceEntryLineItemIssues(ctx, gen.ListServiceEntryLineItemIssuesParams{
		LineItemID: lineItemID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountServiceEntryLineItemIssues(ctx, gen.CountServiceEntryLineItemIssuesParams{
		LineItemID: lineItemID, CompanyID: company,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.IssueResponse, len(rows))
	for i, r := range rows {
		out[i] = toIssueResponse(r)
	}
	return out, total, nil
}

func (s *ServiceEntryLineItemIssueStore) Link(ctx context.Context, lineItemID int64, in dto.LinkIssueRequest) (dto.IssueResponse, error) {
	company := middleware.CompanyFromContext(ctx)
	issue, err := s.q.FindServiceEntryLineItemIssueCandidate(ctx, gen.FindServiceEntryLineItemIssueCandidateParams{
		IssueID: in.IssueID, CompanyID: company, LineItemID: lineItemID,
	})
	if err != nil {
		return dto.IssueResponse{}, err
	}
	if err := s.q.AddServiceEntryLineItemIssue(ctx, gen.AddServiceEntryLineItemIssueParams{
		LineItemID: lineItemID, IssueID: issue.ID,
	}); err != nil {
		return dto.IssueResponse{}, err
	}
	return toIssueResponse(issue), nil
}

func (s *ServiceEntryLineItemIssueStore) Unlink(ctx context.Context, lineItemID, issueID int64) error {
	return s.q.RemoveServiceEntryLineItemIssue(ctx, gen.RemoveServiceEntryLineItemIssueParams{
		LineItemID: lineItemID, IssueID: issueID, CompanyID: middleware.CompanyFromContext(ctx),
	})
}

// --------------------------------------------------------------------------
// work order line item -> issues
// --------------------------------------------------------------------------

type WorkOrderLineItemIssueStore struct{ q *gen.Queries }

func NewWorkOrderLineItemIssueStore(q *gen.Queries) *WorkOrderLineItemIssueStore {
	return &WorkOrderLineItemIssueStore{q: q}
}

func (s *WorkOrderLineItemIssueStore) List(ctx context.Context, lineItemID int64, p paginate.Params) ([]dto.IssueResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWorkOrderLineItemIssues(ctx, gen.ListWorkOrderLineItemIssuesParams{
		LineItemID: lineItemID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWorkOrderLineItemIssues(ctx, gen.CountWorkOrderLineItemIssuesParams{
		LineItemID: lineItemID, CompanyID: company,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.IssueResponse, len(rows))
	for i, r := range rows {
		out[i] = toIssueResponse(r)
	}
	return out, total, nil
}

func (s *WorkOrderLineItemIssueStore) Link(ctx context.Context, lineItemID int64, in dto.LinkIssueRequest) (dto.IssueResponse, error) {
	company := middleware.CompanyFromContext(ctx)
	issue, err := s.q.FindWorkOrderLineItemIssueCandidate(ctx, gen.FindWorkOrderLineItemIssueCandidateParams{
		IssueID: in.IssueID, CompanyID: company, LineItemID: lineItemID,
	})
	if err != nil {
		return dto.IssueResponse{}, err
	}
	if err := s.q.AddWorkOrderLineItemIssue(ctx, gen.AddWorkOrderLineItemIssueParams{
		LineItemID: lineItemID, IssueID: issue.ID,
	}); err != nil {
		return dto.IssueResponse{}, err
	}
	return toIssueResponse(issue), nil
}

func (s *WorkOrderLineItemIssueStore) Unlink(ctx context.Context, lineItemID, issueID int64) error {
	return s.q.RemoveWorkOrderLineItemIssue(ctx, gen.RemoveWorkOrderLineItemIssueParams{
		LineItemID: lineItemID, IssueID: issueID, CompanyID: middleware.CompanyFromContext(ctx),
	})
}
