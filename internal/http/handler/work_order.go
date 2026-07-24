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

type WorkOrderStore struct{ q *gen.Queries }

func NewWorkOrderStore(q *gen.Queries) *WorkOrderStore { return &WorkOrderStore{q: q} }

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
	r, err := s.q.CreateWorkOrder(ctx, gen.CreateWorkOrderParams{
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
		PartsMarkupType:       in.PartsMarkupType,
		PartsMarkup:           in.PartsMarkup,
		PartsMarkupPercentage: in.PartsMarkupPercentage,
		LaborMarkupType:       in.LaborMarkupType,
		LaborMarkup:           in.LaborMarkup,
		LaborMarkupPercentage: in.LaborMarkupPercentage,
		PartsSubtotal:         in.PartsSubtotal,
		LaborSubtotal:         in.LaborSubtotal,
		Subtotal:              in.Subtotal,
		Discount:              in.Discount,
		DiscountType:          in.DiscountType,
		Tax1:                  in.Tax1,
		Tax1Type:              in.Tax1Type,
		Tax1Percentage:        in.Tax1Percentage,
		Tax2:                  in.Tax2,
		Tax2Type:              in.Tax2Type,
		Tax2Percentage:        in.Tax2Percentage,
		TotalAmount:           in.TotalAmount,
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
	return toWorkOrderResponse(r), nil
}

func (s *WorkOrderStore) Update(ctx context.Context, id int64, in dto.UpdateWorkOrderRequest) (dto.WorkOrderResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateWorkOrder(ctx, gen.UpdateWorkOrderParams{
		ID:                    id,
		CompanyID:             middleware.CompanyFromContext(ctx),
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
		PartsSubtotal:         in.PartsSubtotal,
		LaborSubtotal:         in.LaborSubtotal,
		Subtotal:              in.Subtotal,
		Discount:              in.Discount,
		DiscountType:          in.DiscountType,
		Tax1:                  in.Tax1,
		Tax1Type:              in.Tax1Type,
		Tax1Percentage:        in.Tax1Percentage,
		Tax2:                  in.Tax2,
		Tax2Type:              in.Tax2Type,
		Tax2Percentage:        in.Tax2Percentage,
		TotalAmount:           in.TotalAmount,
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
