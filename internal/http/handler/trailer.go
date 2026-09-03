package handler

import (
	"context"
	"slices"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
)

type TrailerStore struct{ q *gen.Queries }

func NewTrailerStore(q *gen.Queries) *TrailerStore { return &TrailerStore{q: q} }

func (s *TrailerStore) Get(ctx context.Context, parentID int64) (dto.TrailerResponse, error) {
	r, err := s.q.GetTrailer(ctx, gen.GetTrailerParams{ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.TrailerResponse{}, err
	}
	out := toTrailerResponse(r)
	// Names resolved separately; a trailer with no classification set simply has
	// none, which is not an error.
	if names, err := s.q.GetTrailerClassificationNames(ctx, parentID); err == nil {
		out.ClassificationName = names.ClassificationName
		out.Classification2Name = names.Classification2Name
	}
	return out, nil
}

func (s *TrailerStore) Upsert(ctx context.Context, parentID int64, in dto.UpsertTrailerRequest) (dto.TrailerResponse, error) {
	return upsertTrailerTx(ctx, s.q, parentID, middleware.CompanyFromContext(ctx), in)
}

// upsertTrailerTx is the tx-friendly core of TrailerStore.Upsert: it takes its
// *gen.Queries and company explicitly so it can run against either the
// singleton store's own queries or a transaction-scoped one, e.g. when an
// asset create/update needs the trailer write to commit or roll back with it.
func upsertTrailerTx(ctx context.Context, q *gen.Queries, parentID, company int64, in dto.UpsertTrailerRequest) (dto.TrailerResponse, error) {
	// The classification ids must belong to the caller's company. Without this a
	// trailer could be linked to another tenant's vocabulary — a value its own
	// fleet cannot see, edit or filter by, and which leaks that tenant's terms.
	if err := validateTrailerClassificationsTx(ctx, q, company, in.ClassificationID, in.Classification2ID); err != nil {
		return dto.TrailerResponse{}, err
	}

	r, err := q.UpsertTrailer(ctx, gen.UpsertTrailerParams{
		ParentID:             parentID,
		CompanyID:            company,
		TrailerType:          in.TrailerType,
		ClassificationID:     in.ClassificationID,
		Classification2ID:    in.Classification2ID,
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
		ClassificationID:     r.ClassificationID,
		Classification2ID:    r.Classification2ID,
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

// validateClassifications refuses ids that do not name a classification in this
// company, as a 422 against the offending field rather than the 409 a foreign
// key would raise from somewhere the caller cannot see.
func (s *TrailerStore) validateClassifications(ctx context.Context, companyID int64, ids ...*int64) error {
	return validateTrailerClassificationsTx(ctx, s.q, companyID, ids...)
}

// validateTrailerClassificationsTx is the tx-friendly core of
// TrailerStore.validateClassifications: it takes its *gen.Queries explicitly
// so it can run inside an asset create/update transaction, where a bad
// classification id must roll back the asset row along with the trailer.
func validateTrailerClassificationsTx(ctx context.Context, q *gen.Queries, companyID int64, ids ...*int64) error {
	var want []int64
	fields := []string{"classification_id", "classification_2_id"}
	details := map[string]string{}

	for i, id := range ids {
		if id != nil {
			want = append(want, *id)
			details[fields[i]] = "must name a classification in your company"
		}
	}
	if len(want) == 0 {
		return nil
	}

	n, err := q.CountTrailerClassificationsByIDs(ctx, gen.CountTrailerClassificationsByIDsParams{
		CompanyID: companyID,
		Ids:       want,
	})
	if err != nil {
		return err
	}
	if int(n) != len(dedupe(want)) {
		return apierr.Validation(details)
	}
	return nil
}

// dedupe is needed because both columns may legitimately name the same
// classification, and the count would then be one where two ids were sent.
func dedupe(ids []int64) []int64 {
	out := slices.Clone(ids)
	slices.Sort(out)
	return slices.Compact(out)
}
