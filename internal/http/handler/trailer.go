package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
)

type TrailerStore struct{ q *gen.Queries }

func NewTrailerStore(q *gen.Queries) *TrailerStore { return &TrailerStore{q: q} }

func (s *TrailerStore) Get(ctx context.Context, parentID int64) (dto.TrailerResponse, error) {
	r, err := s.q.GetTrailer(ctx, gen.GetTrailerParams{ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.TrailerResponse{}, err
	}
	return toTrailerResponse(r), nil
}

func (s *TrailerStore) Upsert(ctx context.Context, parentID int64, in dto.UpsertTrailerRequest) (dto.TrailerResponse, error) {
	r, err := s.q.UpsertTrailer(ctx, gen.UpsertTrailerParams{
		ParentID:             parentID,
		CompanyID:            middleware.CompanyFromContext(ctx),
		TrailerType:          in.TrailerType,
		Classification:       in.Classification,
		Classification2:      in.Classification2,
		Size:                 in.Size,
		Suspension:           in.Suspension,
		OwnerName:            in.OwnerName,
		Financing:            in.Financing,
		Supplier:             in.Supplier,
		LicensePlateUs:       in.LicensePlateUs,
		LicensePlateUsState:  in.LicensePlateUsState,
		LicensePlateMx:       in.LicensePlateMx,
		Doors:                in.Doors,
		Walls:                in.Walls,
		RailPost:             in.RailPost,
		Skylight:             in.Skylight,
		FloorType:            in.FloorType,
		RoofType:             in.RoofType,
		HazmatType:           in.HazmatType,
		AeroKitType:          in.AeroKitType,
		GpsProvider:          in.GpsProvider,
		GpsSerial:            in.GpsSerial,
		GpsSignalStatus:      in.GpsSignalStatus,
		GpsContractStart:     in.GpsContractStart,
		GpsContractEnd:       in.GpsContractEnd,
		GpsContractReference: in.GpsContractReference,
		ContractStart:        in.ContractStart,
		ContractEnd:          in.ContractEnd,
		ContractPeriod:       in.ContractPeriod,
		ContractReference:    in.ContractReference,
		FumigationDate:       in.FumigationDate,
		FumigationCert:       in.FumigationCert,
		RegistrationDate:     in.RegistrationDate,
		WaterproofingDate:    in.WaterproofingDate,
		DecommissionReason:   in.DecommissionReason,
		DecommissionDate:     in.DecommissionDate,
		OperationalUse:       in.OperationalUse,
		OperationZone:        in.OperationZone,
	})
	if err != nil {
		return dto.TrailerResponse{}, err
	}
	return toTrailerResponse(r), nil
}

func (s *TrailerStore) Delete(ctx context.Context, parentID int64) error {
	return s.q.DeleteTrailer(ctx, gen.DeleteTrailerParams{ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toTrailerResponse(r gen.Trailer) dto.TrailerResponse {
	return dto.TrailerResponse{
		AssetID:              r.AssetID,
		TrailerType:          r.TrailerType,
		Classification:       r.Classification,
		Classification2:      r.Classification2,
		Size:                 r.Size,
		Suspension:           r.Suspension,
		OwnerName:            r.OwnerName,
		Financing:            r.Financing,
		Supplier:             r.Supplier,
		LicensePlateUs:       r.LicensePlateUs,
		LicensePlateUsState:  r.LicensePlateUsState,
		LicensePlateMx:       r.LicensePlateMx,
		Doors:                r.Doors,
		Walls:                r.Walls,
		RailPost:             r.RailPost,
		Skylight:             r.Skylight,
		FloorType:            r.FloorType,
		RoofType:             r.RoofType,
		HazmatType:           r.HazmatType,
		AeroKitType:          r.AeroKitType,
		GpsProvider:          r.GpsProvider,
		GpsSerial:            r.GpsSerial,
		GpsSignalStatus:      r.GpsSignalStatus,
		GpsContractStart:     r.GpsContractStart,
		GpsContractEnd:       r.GpsContractEnd,
		GpsContractReference: r.GpsContractReference,
		ContractStart:        r.ContractStart,
		ContractEnd:          r.ContractEnd,
		ContractPeriod:       r.ContractPeriod,
		ContractReference:    r.ContractReference,
		FumigationDate:       r.FumigationDate,
		FumigationCert:       r.FumigationCert,
		RegistrationDate:     r.RegistrationDate,
		WaterproofingDate:    r.WaterproofingDate,
		DecommissionReason:   r.DecommissionReason,
		DecommissionDate:     r.DecommissionDate,
		OperationalUse:       r.OperationalUse,
		OperationZone:        r.OperationZone,
	}
}
