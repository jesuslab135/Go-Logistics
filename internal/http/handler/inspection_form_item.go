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

type InspectionFormItemStore struct{ q *gen.Queries }

func NewInspectionFormItemStore(q *gen.Queries) *InspectionFormItemStore {
	return &InspectionFormItemStore{q: q}
}

func (s *InspectionFormItemStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.InspectionFormItemResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListInspectionFormItems(ctx, gen.ListInspectionFormItemsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountInspectionFormItems(ctx, gen.CountInspectionFormItemsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InspectionFormItemResponse, len(rows))
	for i, r := range rows {
		out[i] = toInspectionFormItemResponse(r)
	}
	return out, total, nil
}

func (s *InspectionFormItemStore) Get(ctx context.Context, parentID, id int64) (dto.InspectionFormItemResponse, error) {
	r, err := s.q.GetInspectionFormItem(ctx, gen.GetInspectionFormItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.InspectionFormItemResponse{}, err
	}
	return toInspectionFormItemResponse(r), nil
}

func (s *InspectionFormItemStore) Create(ctx context.Context, parentID int64, in dto.CreateInspectionFormItemRequest) (dto.InspectionFormItemResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateInspectionFormItem(ctx, gen.CreateInspectionFormItemParams{
		ParentID:                           parentID,
		CompanyID:                          middleware.CompanyFromContext(ctx),
		ItemType:                           in.ItemType,
		Label:                              in.Label,
		ShortDescription:                   in.ShortDescription,
		Instructions:                       in.Instructions,
		Position:                           in.Position,
		IsRequired:                         boolOrDefault(in.IsRequired, true),
		PassLabel:                          orDefault(in.PassLabel, "Pass"),
		FailLabel:                          orDefault(in.FailLabel, "Fail"),
		NaLabel:                            orDefault(in.NaLabel, "N/A"),
		EnableNaOption:                     in.EnableNaOption,
		RequireRemarkOnFail:                boolOrDefault(in.RequireRemarkOnFail, true),
		RequireRemarkOnPass:                in.RequireRemarkOnPass,
		RequirePhotoOnFail:                 in.RequirePhotoOnFail,
		RequireMeterEntryPhotoVerification: in.RequireMeterEntryPhotoVerification,
		RequireSecondaryMeterIfOneExists:   in.RequireSecondaryMeterIfOneExists,
		TypeConfig:                         jsonbOrDefault(in.TypeConfig, "{}"),
		CreatedAt:                          now,
		UpdatedAt:                          now,
	})
	if err != nil {
		return dto.InspectionFormItemResponse{}, err
	}
	return toInspectionFormItemResponse(r), nil
}

func (s *InspectionFormItemStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateInspectionFormItemRequest) (dto.InspectionFormItemResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateInspectionFormItem(ctx, gen.UpdateInspectionFormItemParams{
		ID:                                 id,
		ParentID:                           parentID,
		CompanyID:                          middleware.CompanyFromContext(ctx),
		ItemType:                           in.ItemType,
		Label:                              in.Label,
		ShortDescription:                   in.ShortDescription,
		Instructions:                       in.Instructions,
		Position:                           in.Position,
		IsRequired:                         in.IsRequired,
		PassLabel:                          in.PassLabel,
		FailLabel:                          in.FailLabel,
		NaLabel:                            in.NaLabel,
		EnableNaOption:                     in.EnableNaOption,
		RequireRemarkOnFail:                in.RequireRemarkOnFail,
		RequireRemarkOnPass:                in.RequireRemarkOnPass,
		RequirePhotoOnFail:                 in.RequirePhotoOnFail,
		RequireMeterEntryPhotoVerification: in.RequireMeterEntryPhotoVerification,
		RequireSecondaryMeterIfOneExists:   in.RequireSecondaryMeterIfOneExists,
		TypeConfig:                         jsonbOrDefault(in.TypeConfig, "{}"),
		UpdatedAt:                          now,
	})
	if err != nil {
		return dto.InspectionFormItemResponse{}, err
	}
	return toInspectionFormItemResponse(r), nil
}

func (s *InspectionFormItemStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteInspectionFormItem(ctx, gen.DeleteInspectionFormItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toInspectionFormItemResponse(r gen.InspectionFormItem) dto.InspectionFormItemResponse {
	return dto.InspectionFormItemResponse{
		ID:                                 r.ID,
		FormID:                             r.FormID,
		ItemType:                           r.ItemType,
		Label:                              r.Label,
		ShortDescription:                   r.ShortDescription,
		Instructions:                       r.Instructions,
		Position:                           r.Position,
		IsRequired:                         r.IsRequired,
		PassLabel:                          r.PassLabel,
		FailLabel:                          r.FailLabel,
		NaLabel:                            r.NaLabel,
		EnableNaOption:                     r.EnableNaOption,
		RequireRemarkOnFail:                r.RequireRemarkOnFail,
		RequireRemarkOnPass:                r.RequireRemarkOnPass,
		RequirePhotoOnFail:                 r.RequirePhotoOnFail,
		RequireMeterEntryPhotoVerification: r.RequireMeterEntryPhotoVerification,
		RequireSecondaryMeterIfOneExists:   r.RequireSecondaryMeterIfOneExists,
		TypeConfig:                         json.RawMessage(r.TypeConfig),
		CreatedAt:                          r.CreatedAt,
		UpdatedAt:                          r.UpdatedAt,
	}
}
