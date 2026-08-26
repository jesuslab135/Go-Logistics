package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type WorkOrderStore struct {
	q *gen.Queries
	// pool is needed because a status change and the row recording it must land
	// together. Django did this with a post_save signal; the transaction is the
	// same guarantee without the indirection.
	pool *pgxpool.Pool
}

func NewWorkOrderStore(q *gen.Queries, pool *pgxpool.Pool) *WorkOrderStore {
	return &WorkOrderStore{q: q, pool: pool}
}

// logStatusChange appends to the work order's history. It is unexported and
// takes a transaction on purpose: the log is written by whatever changed the
// status, never by a route, so the record can never disagree with the record it
// describes.
//
// The actor is the authenticated caller when there is one. A transition with no
// caller is a system transition, which is why actor_employee_id is nullable and
// actor_type carries the distinction rather than leaving a null to interpret.
func logStatusChange(ctx context.Context, qtx *gen.Queries, workOrderID, statusID, companyID int64, at time.Time) error {
	actorID, actorType := actorOf(ctx)

	_, err := qtx.CreateWorkOrderStatusLog(ctx, gen.CreateWorkOrderStatusLogParams{
		ParentID:        workOrderID,
		CompanyID:       companyID,
		StatusID:        statusID,
		ChangedAt:       at,
		ActorEmployeeID: actorID,
		ActorType:       actorType,
	})
	return err
}

func (s *WorkOrderStore) List(ctx context.Context, p paginate.Params) ([]dto.WorkOrderResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWorkOrders(ctx, gen.ListWorkOrdersParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWorkOrders(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.WorkOrderResponse, len(rows))
	for i, r := range rows {
		out[i] = toWorkOrderResponse(r)
	}
	return out, total, nil
}

func (s *WorkOrderStore) Get(ctx context.Context, id int64) (dto.WorkOrderResponse, error) {
	r, err := s.q.GetWorkOrder(ctx, gen.GetWorkOrderParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.WorkOrderResponse{}, err
	}
	return toWorkOrderResponse(r), nil
}

func (s *WorkOrderStore) Create(ctx context.Context, in dto.CreateWorkOrderRequest) (dto.WorkOrderResponse, error) {
	now := time.Now().UTC()
	company := middleware.CompanyFromContext(ctx)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return dto.WorkOrderResponse{}, err
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	r, err := qtx.CreateWorkOrder(ctx, gen.CreateWorkOrderParams{
		LocationID:            in.LocationID,
		CompanyID:             middleware.CompanyFromContext(ctx),
		Number:                in.Number,
		Description:           in.Description,
		AssetID:               in.AssetID,
		StatusID:              in.StatusID,
		VendorID:              in.VendorID,
		AssignedToID:          in.AssignedToID,
		IssuedByID:            in.IssuedByID,
		FaultID:               in.FaultID,
		IssuedAt:              in.IssuedAt,
		ScheduledAt:           in.ScheduledAt,
		StartedAt:             in.StartedAt,
		ExpectedCompletedAt:   in.ExpectedCompletedAt,
		CompletedAt:           in.CompletedAt,
		StartingMeter:         in.StartingMeter,
		EndingMeter:           in.EndingMeter,
		DurationSeconds:       in.DurationSeconds,
		LaborTimeSeconds:      in.LaborTimeSeconds,
		PartsMarkupType:       orDefault(in.PartsMarkupType, "PERCENTAGE"),
		PartsMarkup:           in.PartsMarkup,
		PartsMarkupPercentage: in.PartsMarkupPercentage,
		LaborMarkupType:       orDefault(in.LaborMarkupType, "PERCENTAGE"),
		LaborMarkup:           in.LaborMarkup,
		LaborMarkupPercentage: in.LaborMarkupPercentage,
		Discount:              in.Discount,
		DiscountType:          orDefault(in.DiscountType, "FIXED"),
		Tax1:                  in.Tax1,
		Tax1Type:              orDefault(in.Tax1Type, "PERCENTAGE"),
		Tax1Percentage:        in.Tax1Percentage,
		Tax2:                  in.Tax2,
		Tax2Type:              orDefault(in.Tax2Type, "PERCENTAGE"),
		Tax2Percentage:        in.Tax2Percentage,
		InvoiceNumber:         in.InvoiceNumber,
		PurchaseOrderNumber:   in.PurchaseOrderNumber,
		CommentsCount:         in.CommentsCount,
		ImagesCount:           in.ImagesCount,
		DocumentsCount:        in.DocumentsCount,
		Labels:                jsonbOrDefault(in.Labels, "[]"),
		CustomFields:          jsonbOrDefault(in.CustomFields, "{}"),
		CreatedAt:             now,
		UpdatedAt:             now,
	})
	if err != nil {
		return dto.WorkOrderResponse{}, err
	}

	// A work order opens in a status, and that is the first thing its history has
	// to show — Django logged on create for the same reason.
	if err := logStatusChange(ctx, qtx, r.ID, r.StatusID, company, now); err != nil {
		return dto.WorkOrderResponse{}, err
	}
	// A new order has no lines, so this stores zeros — but it stores computed
	// zeros rather than whatever the request happened to send.
	if err := recalcWorkOrder(ctx, qtx, r.ID, now); err != nil {
		return dto.WorkOrderResponse{}, err
	}
	r, err = qtx.GetWorkOrder(ctx, gen.GetWorkOrderParams{ID: r.ID, CompanyID: company})
	if err != nil {
		return dto.WorkOrderResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.WorkOrderResponse{}, err
	}
	return toWorkOrderResponse(r), nil
}

func (s *WorkOrderStore) Update(ctx context.Context, id int64, in dto.UpdateWorkOrderRequest) (dto.WorkOrderResponse, error) {
	now := time.Now().UTC()
	company := middleware.CompanyFromContext(ctx)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return dto.WorkOrderResponse{}, err
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	// Read the stored status inside the transaction: whether this write is a
	// transition is decided against the row being replaced, not against whatever
	// a concurrent request left behind.
	previous, err := qtx.GetWorkOrder(ctx, gen.GetWorkOrderParams{ID: id, CompanyID: company})
	if err != nil {
		return dto.WorkOrderResponse{}, err
	}

	r, err := qtx.UpdateWorkOrder(ctx, gen.UpdateWorkOrderParams{
		ID:                    id,
		CompanyID:             company,
		LocationID:            in.LocationID,
		Number:                in.Number,
		Description:           in.Description,
		AssetID:               in.AssetID,
		StatusID:              in.StatusID,
		VendorID:              in.VendorID,
		AssignedToID:          in.AssignedToID,
		IssuedByID:            in.IssuedByID,
		FaultID:               in.FaultID,
		IssuedAt:              in.IssuedAt,
		ScheduledAt:           in.ScheduledAt,
		StartedAt:             in.StartedAt,
		ExpectedCompletedAt:   in.ExpectedCompletedAt,
		CompletedAt:           in.CompletedAt,
		StartingMeter:         in.StartingMeter,
		EndingMeter:           in.EndingMeter,
		DurationSeconds:       in.DurationSeconds,
		LaborTimeSeconds:      in.LaborTimeSeconds,
		PartsMarkupType:       in.PartsMarkupType,
		PartsMarkup:           in.PartsMarkup,
		PartsMarkupPercentage: in.PartsMarkupPercentage,
		LaborMarkupType:       in.LaborMarkupType,
		LaborMarkup:           in.LaborMarkup,
		LaborMarkupPercentage: in.LaborMarkupPercentage,
		Discount:              in.Discount,
		DiscountType:          in.DiscountType,
		Tax1:                  in.Tax1,
		Tax1Type:              in.Tax1Type,
		Tax1Percentage:        in.Tax1Percentage,
		Tax2:                  in.Tax2,
		Tax2Type:              in.Tax2Type,
		Tax2Percentage:        in.Tax2Percentage,
		InvoiceNumber:         in.InvoiceNumber,
		PurchaseOrderNumber:   in.PurchaseOrderNumber,
		CommentsCount:         in.CommentsCount,
		ImagesCount:           in.ImagesCount,
		DocumentsCount:        in.DocumentsCount,
		Labels:                jsonbOrDefault(in.Labels, "[]"),
		CustomFields:          jsonbOrDefault(in.CustomFields, "{}"),
		UpdatedAt:             now,
	})
	if err != nil {
		return dto.WorkOrderResponse{}, err
	}

	// Only a transition is history. A save that leaves the status alone would
	// otherwise file a row saying nothing happened.
	if r.StatusID != previous.StatusID {
		if err := logStatusChange(ctx, qtx, r.ID, r.StatusID, company, now); err != nil {
			return dto.WorkOrderResponse{}, err
		}
	}
	// Being handed work is news to exactly one person, and only when the
	// assignee actually changed — re-saving a form does not re-announce it.
	if err := notifyWorkOrderAssignee(ctx, qtx, r, previous.AssignedToID, company, now); err != nil {
		return dto.WorkOrderResponse{}, err
	}
	// This write can change the markup, discount and tax terms, so the totals
	// are recomputed from them and re-read.
	if err := recalcWorkOrder(ctx, qtx, id, now); err != nil {
		return dto.WorkOrderResponse{}, err
	}
	r, err = qtx.GetWorkOrder(ctx, gen.GetWorkOrderParams{ID: id, CompanyID: company})
	if err != nil {
		return dto.WorkOrderResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.WorkOrderResponse{}, err
	}
	return toWorkOrderResponse(r), nil
}

func (s *WorkOrderStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteWorkOrder(ctx, gen.DeleteWorkOrderParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toWorkOrderResponse(r gen.WorkOrder) dto.WorkOrderResponse {
	return dto.WorkOrderResponse{
		ID:                    r.ID,
		LocationID:            r.LocationID,
		CompanyID:             r.CompanyID,
		Number:                r.Number,
		Description:           r.Description,
		AssetID:               r.AssetID,
		StatusID:              r.StatusID,
		VendorID:              r.VendorID,
		AssignedToID:          r.AssignedToID,
		IssuedByID:            r.IssuedByID,
		FaultID:               r.FaultID,
		IssuedAt:              r.IssuedAt,
		ScheduledAt:           r.ScheduledAt,
		StartedAt:             r.StartedAt,
		ExpectedCompletedAt:   r.ExpectedCompletedAt,
		CompletedAt:           r.CompletedAt,
		StartingMeter:         r.StartingMeter,
		EndingMeter:           r.EndingMeter,
		DurationSeconds:       r.DurationSeconds,
		LaborTimeSeconds:      r.LaborTimeSeconds,
		PartsMarkupType:       r.PartsMarkupType,
		PartsMarkup:           r.PartsMarkup,
		PartsMarkupPercentage: r.PartsMarkupPercentage,
		LaborMarkupType:       r.LaborMarkupType,
		LaborMarkup:           r.LaborMarkup,
		LaborMarkupPercentage: r.LaborMarkupPercentage,
		PartsSubtotal:         r.PartsSubtotal,
		LaborSubtotal:         r.LaborSubtotal,
		Subtotal:              r.Subtotal,
		Discount:              r.Discount,
		DiscountType:          r.DiscountType,
		Tax1:                  r.Tax1,
		Tax1Type:              r.Tax1Type,
		Tax1Percentage:        r.Tax1Percentage,
		Tax2:                  r.Tax2,
		Tax2Type:              r.Tax2Type,
		Tax2Percentage:        r.Tax2Percentage,
		TotalAmount:           r.TotalAmount,
		InvoiceNumber:         r.InvoiceNumber,
		PurchaseOrderNumber:   r.PurchaseOrderNumber,
		CommentsCount:         r.CommentsCount,
		ImagesCount:           r.ImagesCount,
		DocumentsCount:        r.DocumentsCount,
		Labels:                json.RawMessage(r.Labels),
		CustomFields:          json.RawMessage(r.CustomFields),
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}

// notifyWorkOrderAssignee tells the new assignee that work is theirs.
//
// Nothing is sent when the assignee did not change, when the order is assigned
// to nobody, or when the caller assigned it to themselves — being told about
// your own action is not a notification, it is an echo.
//
// It runs in the transaction that made the assignment, so a notification that
// exists always describes an assignment that happened.
func notifyWorkOrderAssignee(ctx context.Context, qtx *gen.Queries, order gen.WorkOrder, previous *int64, companyID int64, at time.Time) error {
	assignee := order.AssignedToID
	if assignee == nil {
		return nil
	}
	if previous != nil && *previous == *assignee {
		return nil
	}
	if actor := middleware.EmployeeFromContext(ctx); actor == *assignee {
		return nil
	}

	url := "/app/maintenance/work-orders/" + strconv.FormatInt(order.ID, 10)
	if !ValidNotificationURL(url) {
		// Unreachable with a numeric id, but the check is where the rule lives:
		// a producer must not be the one place that can store an off-site link.
		return nil
	}

	return qtx.NotifyWorkOrderAssignee(ctx, gen.NotifyWorkOrderAssigneeParams{
		CompanyID:  companyID,
		EmployeeID: *assignee,
		Kind:       NotificationWorkOrderAssigned,
		Title:      "Work order " + order.Number + " assigned to you",
		Body:       order.Description,
		Url:        url,
		CreatedAt:  at,
	})
}
