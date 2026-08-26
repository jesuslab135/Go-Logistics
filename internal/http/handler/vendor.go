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

type VendorStore struct{ q *gen.Queries }

func NewVendorStore(q *gen.Queries) *VendorStore { return &VendorStore{q: q} }

func (s *VendorStore) List(ctx context.Context, p paginate.Params) ([]dto.VendorResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListVendors(ctx, gen.ListVendorsParams{CompanyID: company, IncludeArchived: includeArchived(ctx), Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountVendors(ctx, gen.CountVendorsParams{CompanyID: company, IncludeArchived: includeArchived(ctx)})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.VendorResponse, len(rows))
	for i, r := range rows {
		out[i] = toVendorResponse(r)
	}
	return out, total, nil
}

func (s *VendorStore) Get(ctx context.Context, id int64) (dto.VendorResponse, error) {
	r, err := s.q.GetVendor(ctx, gen.GetVendorParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.VendorResponse{}, err
	}
	return toVendorResponse(r), nil
}

func (s *VendorStore) Create(ctx context.Context, in dto.CreateVendorRequest) (dto.VendorResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateVendor(ctx, gen.CreateVendorParams{
		CompanyID:          middleware.CompanyFromContext(ctx),
		Name:               in.Name,
		IsMobileService:    in.IsMobileService,
		StreetAddress:      in.StreetAddress,
		StreetAddressLine2: in.StreetAddressLine2,
		City:               in.City,
		Region:             in.Region,
		PostalCode:         in.PostalCode,
		Country:            in.Country,
		Phone:              in.Phone,
		Website:            in.Website,
		ContactName:        in.ContactName,
		ContactPhone:       in.ContactPhone,
		ContactEmail:       in.ContactEmail,
		ExternalID:         in.ExternalID,
		Latitude:           in.Latitude,
		Longitude:          in.Longitude,
		IsFuelVendor:       in.IsFuelVendor,
		IsServiceVendor:    in.IsServiceVendor,
		IsPartsVendor:      in.IsPartsVendor,
		Labels:             jsonbOrDefault(in.Labels, "[]"),
		CustomFields:       jsonbOrDefault(in.CustomFields, "{}"),
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if err != nil {
		return dto.VendorResponse{}, err
	}
	return toVendorResponse(r), nil
}

func (s *VendorStore) Update(ctx context.Context, id int64, in dto.UpdateVendorRequest) (dto.VendorResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateVendor(ctx, gen.UpdateVendorParams{
		ID:                 id,
		CompanyID:          middleware.CompanyFromContext(ctx),
		Name:               in.Name,
		IsMobileService:    in.IsMobileService,
		StreetAddress:      in.StreetAddress,
		StreetAddressLine2: in.StreetAddressLine2,
		City:               in.City,
		Region:             in.Region,
		PostalCode:         in.PostalCode,
		Country:            in.Country,
		Phone:              in.Phone,
		Website:            in.Website,
		ContactName:        in.ContactName,
		ContactPhone:       in.ContactPhone,
		ContactEmail:       in.ContactEmail,
		ExternalID:         in.ExternalID,
		Latitude:           in.Latitude,
		Longitude:          in.Longitude,
		IsFuelVendor:       in.IsFuelVendor,
		IsServiceVendor:    in.IsServiceVendor,
		IsPartsVendor:      in.IsPartsVendor,
		Labels:             jsonbOrDefault(in.Labels, "[]"),
		CustomFields:       jsonbOrDefault(in.CustomFields, "{}"),
		UpdatedAt:          now,
	})
	if err != nil {
		return dto.VendorResponse{}, err
	}
	return toVendorResponse(r), nil
}

func (s *VendorStore) Delete(ctx context.Context, id int64) error {
	if err := guardDelete(ctx, "vendor", id, s.q.VendorReferences); err != nil {
		return err
	}
	return s.q.DeleteVendor(ctx, gen.DeleteVendorParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toVendorResponse(r gen.Vendor) dto.VendorResponse {
	return dto.VendorResponse{
		ID:                 r.ID,
		CompanyID:          r.CompanyID,
		Name:               r.Name,
		IsMobileService:    r.IsMobileService,
		StreetAddress:      r.StreetAddress,
		StreetAddressLine2: r.StreetAddressLine2,
		City:               r.City,
		Region:             r.Region,
		PostalCode:         r.PostalCode,
		Country:            r.Country,
		Phone:              r.Phone,
		Website:            r.Website,
		ContactName:        r.ContactName,
		ContactPhone:       r.ContactPhone,
		ContactEmail:       r.ContactEmail,
		ExternalID:         r.ExternalID,
		Latitude:           r.Latitude,
		Longitude:          r.Longitude,
		IsFuelVendor:       r.IsFuelVendor,
		IsServiceVendor:    r.IsServiceVendor,
		IsPartsVendor:      r.IsPartsVendor,
		Labels:             json.RawMessage(r.Labels),
		ArchivedAt:         r.ArchivedAt,
		CustomFields:       json.RawMessage(r.CustomFields),
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}
