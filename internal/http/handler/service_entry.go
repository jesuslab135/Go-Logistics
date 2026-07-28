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

type ServiceEntryStore struct{ q *gen.Queries }

func NewServiceEntryStore(q *gen.Queries) *ServiceEntryStore { return &ServiceEntryStore{q: q} }

func (s *ServiceEntryStore) List(ctx context.Context, p paginate.Params) ([]dto.ServiceEntryResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListServiceEntries(ctx, gen.ListServiceEntriesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountServiceEntries(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ServiceEntryResponse, len(rows))
	for i, r := range rows {
		out[i] = toServiceEntryResponse(r)
	}
	return out, total, nil
}

func (s *ServiceEntryStore) Get(ctx context.Context, id int64) (dto.ServiceEntryResponse, error) {
	r, err := s.q.GetServiceEntry(ctx, gen.GetServiceEntryParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.ServiceEntryResponse{}, err
	}
	return toServiceEntryResponse(r), nil
}

func (s *ServiceEntryStore) Create(ctx context.Context, in dto.CreateServiceEntryRequest) (dto.ServiceEntryResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateServiceEntry(ctx, gen.CreateServiceEntryParams{
		CompanyID:            middleware.CompanyFromContext(ctx),
		Reference:            in.Reference,
		Status:               orDefault(in.Status, "COMPLETED"),
		AssetID:              in.AssetID,
		VendorID:             in.VendorID,
		WorkOrderID:          in.WorkOrderID,
		StartedAt:            in.StartedAt,
		CompletedAt:          in.CompletedAt,
		MeterValue:           in.MeterValue,
		PartsSubtotal:        in.PartsSubtotal,
		LaborSubtotal:        in.LaborSubtotal,
		Subtotal:             in.Subtotal,
		Discount:             in.Discount,
		DiscountType:         orDefault(in.DiscountType, "FIXED"),
		Tax1:                 in.Tax1,
		Tax1Type:             orDefault(in.Tax1Type, "PERCENTAGE"),
		Tax1Percentage:       in.Tax1Percentage,
		Tax2:                 in.Tax2,
		Tax2Type:             orDefault(in.Tax2Type, "PERCENTAGE"),
		Tax2Percentage:       in.Tax2Percentage,
		TotalAmount:          in.TotalAmount,
		GeneralNotes:         in.GeneralNotes,
		IsRoadsideAssistance: in.IsRoadsideAssistance,
		LaborTimeSeconds:     in.LaborTimeSeconds,
		Labels:               jsonbOrDefault(in.Labels, "[]"),
		CustomFields:         jsonbOrDefault(in.CustomFields, "{}"),
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		return dto.ServiceEntryResponse{}, err
	}
	return toServiceEntryResponse(r), nil
}

func (s *ServiceEntryStore) Update(ctx context.Context, id int64, in dto.UpdateServiceEntryRequest) (dto.ServiceEntryResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateServiceEntry(ctx, gen.UpdateServiceEntryParams{
		ID:                   id,
		CompanyID:            middleware.CompanyFromContext(ctx),
		Reference:            in.Reference,
		Status:               in.Status,
		AssetID:              in.AssetID,
		VendorID:             in.VendorID,
		WorkOrderID:          in.WorkOrderID,
		StartedAt:            in.StartedAt,
		CompletedAt:          in.CompletedAt,
		MeterValue:           in.MeterValue,
		PartsSubtotal:        in.PartsSubtotal,
		LaborSubtotal:        in.LaborSubtotal,
		Subtotal:             in.Subtotal,
		Discount:             in.Discount,
		DiscountType:         in.DiscountType,
		Tax1:                 in.Tax1,
		Tax1Type:             in.Tax1Type,
		Tax1Percentage:       in.Tax1Percentage,
		Tax2:                 in.Tax2,
		Tax2Type:             in.Tax2Type,
		Tax2Percentage:       in.Tax2Percentage,
		TotalAmount:          in.TotalAmount,
		GeneralNotes:         in.GeneralNotes,
		IsRoadsideAssistance: in.IsRoadsideAssistance,
		LaborTimeSeconds:     in.LaborTimeSeconds,
		Labels:               jsonbOrDefault(in.Labels, "[]"),
		CustomFields:         jsonbOrDefault(in.CustomFields, "{}"),
		UpdatedAt:            now,
	})
	if err != nil {
		return dto.ServiceEntryResponse{}, err
	}
	return toServiceEntryResponse(r), nil
}

func (s *ServiceEntryStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteServiceEntry(ctx, gen.DeleteServiceEntryParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toServiceEntryResponse(r gen.ServiceEntry) dto.ServiceEntryResponse {
	return dto.ServiceEntryResponse{
		ID:                   r.ID,
		CompanyID:            r.CompanyID,
		Reference:            r.Reference,
		Status:               r.Status,
		AssetID:              r.AssetID,
		VendorID:             r.VendorID,
		WorkOrderID:          r.WorkOrderID,
		StartedAt:            r.StartedAt,
		CompletedAt:          r.CompletedAt,
		MeterValue:           r.MeterValue,
		PartsSubtotal:        r.PartsSubtotal,
		LaborSubtotal:        r.LaborSubtotal,
		Subtotal:             r.Subtotal,
		Discount:             r.Discount,
		DiscountType:         r.DiscountType,
		Tax1:                 r.Tax1,
		Tax1Type:             r.Tax1Type,
		Tax1Percentage:       r.Tax1Percentage,
		Tax2:                 r.Tax2,
		Tax2Type:             r.Tax2Type,
		Tax2Percentage:       r.Tax2Percentage,
		TotalAmount:          r.TotalAmount,
		GeneralNotes:         r.GeneralNotes,
		IsRoadsideAssistance: r.IsRoadsideAssistance,
		LaborTimeSeconds:     r.LaborTimeSeconds,
		Labels:               json.RawMessage(r.Labels),
		CustomFields:         json.RawMessage(r.CustomFields),
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}
