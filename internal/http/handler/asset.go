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

type AssetStore struct{ q *gen.Queries }

func NewAssetStore(q *gen.Queries) *AssetStore { return &AssetStore{q: q} }

func (s *AssetStore) List(ctx context.Context, p paginate.Params) ([]dto.AssetResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListAssets(ctx, gen.ListAssetsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountAssets(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.AssetResponse, len(rows))
	for i, r := range rows {
		out[i] = toAssetResponse(r)
	}
	return out, total, nil
}

func (s *AssetStore) Get(ctx context.Context, id int64) (dto.AssetResponse, error) {
	r, err := s.q.GetAsset(ctx, gen.GetAssetParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.AssetResponse{}, err
	}
	return s.withSubtypes(ctx, toAssetResponse(r))
}

// withSubtypes fills the denormalized vehicle/trailer columns so a single asset
// carries the same shape the list projection returns. The lookup is a primary
// key hit on two 1:1 tables; a missing subtype simply leaves the fields null.
func (s *AssetStore) withSubtypes(ctx context.Context, resp dto.AssetResponse) (dto.AssetResponse, error) {
	sub, err := s.q.GetAssetSubtypeFields(ctx, gen.GetAssetSubtypeFieldsParams{
		ID:        resp.ID,
		CompanyID: middleware.CompanyFromContext(ctx),
	})
	if err != nil {
		return dto.AssetResponse{}, err
	}
	resp.Operator = sub.Operator
	resp.TrailerType = sub.TrailerType
	resp.TrailerClassification = sub.TrailerClassification
	resp.TrailerSize = sub.TrailerSize
	return resp, nil
}

func (s *AssetStore) Create(ctx context.Context, in dto.CreateAssetRequest) (dto.AssetResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateAsset(ctx, gen.CreateAssetParams{
		CompanyID:                   middleware.CompanyFromContext(ctx),
		Name:                        in.Name,
		VinSn:                       in.VinSn,
		Msrp:                        in.Msrp,
		GenerateExpenses:            in.GenerateExpenses,
		AssetTypeID:                 in.AssetTypeID,
		StatusID:                    in.StatusID,
		LeaseVendorID:               in.LeaseVendorID,
		VehicleType:                 orDefault(in.VehicleType, "CAR"),
		OwnershipType:               orDefault(in.OwnershipType, "OWNED"),
		Labels:                      jsonbOrDefault(in.Labels, "[]"),
		LinkedVehicles:              jsonbOrDefault(in.LinkedVehicles, "[]"),
		LoanStartDate:               in.LoanStartDate,
		LoanEndDate:                 in.LoanEndDate,
		MonthlyPayment:              in.MonthlyPayment,
		NumberOfPayments:            in.NumberOfPayments,
		LeaseNumber:                 in.LeaseNumber,
		LeaseStartDate:              in.LeaseStartDate,
		LeaseEndDate:                in.LeaseEndDate,
		ExcessMileageCharge:         in.ExcessMileageCharge,
		OwnerCompanyID:              in.OwnerCompanyID,
		Year:                        in.Year,
		Make:                        in.Make,
		Model:                       in.Model,
		Trim:                        in.Trim,
		Color:                       in.Color,
		LicensePlate:                in.LicensePlate,
		Group:                       in.Group,
		Photo:                       in.Photo,
		MeterUnit:                   orDefault(in.MeterUnit, "mi"),
		CurrentMeter:                in.CurrentMeter,
		SecondaryMeterUnit:          in.SecondaryMeterUnit,
		SecondaryMeterValue:         in.SecondaryMeterValue,
		FuelType:                    in.FuelType,
		BodyType:                    in.BodyType,
		BodySubtype:                 in.BodySubtype,
		RegistrationState:           in.RegistrationState,
		PurchaseDate:                in.PurchaseDate,
		PurchasePrice:               in.PurchasePrice,
		PurchaseVendor:              in.PurchaseVendor,
		PurchaseMeter:               in.PurchaseMeter,
		InServiceDate:               in.InServiceDate,
		InServiceMeter:              in.InServiceMeter,
		OutOfServiceDate:            in.OutOfServiceDate,
		OutOfServiceMeter:           in.OutOfServiceMeter,
		EstimatedServiceMonths:      in.EstimatedServiceMonths,
		EstimatedReplacementMileage: in.EstimatedReplacementMileage,
		EstimatedResalePrice:        in.EstimatedResalePrice,
		AcquisitionType:             in.AcquisitionType,
		MonthlyCost:                 in.MonthlyCost,
		AcquisitionDate:             in.AcquisitionDate,
		LoanAmount:                  in.LoanAmount,
		CapitalizedCost:             in.CapitalizedCost,
		DownPayment:                 in.DownPayment,
		AnnualPercentageRate:        in.AnnualPercentageRate,
		FirstPaymentDate:            in.FirstPaymentDate,
		ResidualValue:               in.ResidualValue,
		MileageCap:                  in.MileageCap,
		Notes:                       in.Notes,
		ArchivedAt:                  in.ArchivedAt,
		ExternalID:                  in.ExternalID,
		CustomFields:                jsonbOrDefault(in.CustomFields, "{}"),
		FuelVolumeUnits:             orDefault(in.FuelVolumeUnits, "liters"),
		CurrentMeterDate:            in.CurrentMeterDate,
		LoanAccountNumber:           in.LoanAccountNumber,
		LoanNotes:                   in.LoanNotes,
		LoanVendorID:                in.LoanVendorID,
		LoanStartedAt:               in.LoanStartedAt,
		LoanEndedAt:                 in.LoanEndedAt,
		UpdatedAt:                   now,
	})
	if err != nil {
		return dto.AssetResponse{}, err
	}
	return s.withSubtypes(ctx, toAssetResponse(r))
}

func (s *AssetStore) Update(ctx context.Context, id int64, in dto.UpdateAssetRequest) (dto.AssetResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateAsset(ctx, gen.UpdateAssetParams{
		ID:                          id,
		CompanyID:                   middleware.CompanyFromContext(ctx),
		Name:                        in.Name,
		VinSn:                       in.VinSn,
		Msrp:                        in.Msrp,
		GenerateExpenses:            in.GenerateExpenses,
		AssetTypeID:                 in.AssetTypeID,
		StatusID:                    in.StatusID,
		LeaseVendorID:               in.LeaseVendorID,
		VehicleType:                 in.VehicleType,
		OwnershipType:               in.OwnershipType,
		Labels:                      jsonbOrDefault(in.Labels, "[]"),
		LinkedVehicles:              jsonbOrDefault(in.LinkedVehicles, "[]"),
		LoanStartDate:               in.LoanStartDate,
		LoanEndDate:                 in.LoanEndDate,
		MonthlyPayment:              in.MonthlyPayment,
		NumberOfPayments:            in.NumberOfPayments,
		LeaseNumber:                 in.LeaseNumber,
		LeaseStartDate:              in.LeaseStartDate,
		LeaseEndDate:                in.LeaseEndDate,
		ExcessMileageCharge:         in.ExcessMileageCharge,
		OwnerCompanyID:              in.OwnerCompanyID,
		Year:                        in.Year,
		Make:                        in.Make,
		Model:                       in.Model,
		Trim:                        in.Trim,
		Color:                       in.Color,
		LicensePlate:                in.LicensePlate,
		Group:                       in.Group,
		Photo:                       in.Photo,
		MeterUnit:                   in.MeterUnit,
		CurrentMeter:                in.CurrentMeter,
		SecondaryMeterUnit:          in.SecondaryMeterUnit,
		SecondaryMeterValue:         in.SecondaryMeterValue,
		FuelType:                    in.FuelType,
		BodyType:                    in.BodyType,
		BodySubtype:                 in.BodySubtype,
		RegistrationState:           in.RegistrationState,
		PurchaseDate:                in.PurchaseDate,
		PurchasePrice:               in.PurchasePrice,
		PurchaseVendor:              in.PurchaseVendor,
		PurchaseMeter:               in.PurchaseMeter,
		InServiceDate:               in.InServiceDate,
		InServiceMeter:              in.InServiceMeter,
		OutOfServiceDate:            in.OutOfServiceDate,
		OutOfServiceMeter:           in.OutOfServiceMeter,
		EstimatedServiceMonths:      in.EstimatedServiceMonths,
		EstimatedReplacementMileage: in.EstimatedReplacementMileage,
		EstimatedResalePrice:        in.EstimatedResalePrice,
		AcquisitionType:             in.AcquisitionType,
		MonthlyCost:                 in.MonthlyCost,
		AcquisitionDate:             in.AcquisitionDate,
		LoanAmount:                  in.LoanAmount,
		CapitalizedCost:             in.CapitalizedCost,
		DownPayment:                 in.DownPayment,
		AnnualPercentageRate:        in.AnnualPercentageRate,
		FirstPaymentDate:            in.FirstPaymentDate,
		ResidualValue:               in.ResidualValue,
		MileageCap:                  in.MileageCap,
		Notes:                       in.Notes,
		ArchivedAt:                  in.ArchivedAt,
		ExternalID:                  in.ExternalID,
		CustomFields:                jsonbOrDefault(in.CustomFields, "{}"),
		FuelVolumeUnits:             in.FuelVolumeUnits,
		CurrentMeterDate:            in.CurrentMeterDate,
		LoanAccountNumber:           in.LoanAccountNumber,
		LoanNotes:                   in.LoanNotes,
		LoanVendorID:                in.LoanVendorID,
		LoanStartedAt:               in.LoanStartedAt,
		LoanEndedAt:                 in.LoanEndedAt,
		UpdatedAt:                   now,
	})
	if err != nil {
		return dto.AssetResponse{}, err
	}
	return s.withSubtypes(ctx, toAssetResponse(r))
}

func (s *AssetStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteAsset(ctx, gen.DeleteAssetParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toAssetResponse(r gen.Asset) dto.AssetResponse {
	return dto.AssetResponse{
		ID:                          r.ID,
		CompanyID:                   r.CompanyID,
		Name:                        r.Name,
		VinSn:                       r.VinSn,
		Msrp:                        r.Msrp,
		GenerateExpenses:            r.GenerateExpenses,
		AssetTypeID:                 r.AssetTypeID,
		StatusID:                    r.StatusID,
		LeaseVendorID:               r.LeaseVendorID,
		VehicleType:                 r.VehicleType,
		OwnershipType:               r.OwnershipType,
		Labels:                      json.RawMessage(r.Labels),
		LinkedVehicles:              json.RawMessage(r.LinkedVehicles),
		LoanStartDate:               r.LoanStartDate,
		LoanEndDate:                 r.LoanEndDate,
		MonthlyPayment:              r.MonthlyPayment,
		NumberOfPayments:            r.NumberOfPayments,
		LeaseNumber:                 r.LeaseNumber,
		LeaseStartDate:              r.LeaseStartDate,
		LeaseEndDate:                r.LeaseEndDate,
		ExcessMileageCharge:         r.ExcessMileageCharge,
		OwnerCompanyID:              r.OwnerCompanyID,
		Year:                        r.Year,
		Make:                        r.Make,
		Model:                       r.Model,
		Trim:                        r.Trim,
		Color:                       r.Color,
		LicensePlate:                r.LicensePlate,
		Group:                       r.Group,
		Photo:                       r.Photo,
		MeterUnit:                   r.MeterUnit,
		CurrentMeter:                r.CurrentMeter,
		SecondaryMeterUnit:          r.SecondaryMeterUnit,
		SecondaryMeterValue:         r.SecondaryMeterValue,
		FuelType:                    r.FuelType,
		BodyType:                    r.BodyType,
		BodySubtype:                 r.BodySubtype,
		RegistrationState:           r.RegistrationState,
		PurchaseDate:                r.PurchaseDate,
		PurchasePrice:               r.PurchasePrice,
		PurchaseVendor:              r.PurchaseVendor,
		PurchaseMeter:               r.PurchaseMeter,
		InServiceDate:               r.InServiceDate,
		InServiceMeter:              r.InServiceMeter,
		OutOfServiceDate:            r.OutOfServiceDate,
		OutOfServiceMeter:           r.OutOfServiceMeter,
		EstimatedServiceMonths:      r.EstimatedServiceMonths,
		EstimatedReplacementMileage: r.EstimatedReplacementMileage,
		EstimatedResalePrice:        r.EstimatedResalePrice,
		AcquisitionType:             r.AcquisitionType,
		MonthlyCost:                 r.MonthlyCost,
		AcquisitionDate:             r.AcquisitionDate,
		LoanAmount:                  r.LoanAmount,
		CapitalizedCost:             r.CapitalizedCost,
		DownPayment:                 r.DownPayment,
		AnnualPercentageRate:        r.AnnualPercentageRate,
		FirstPaymentDate:            r.FirstPaymentDate,
		ResidualValue:               r.ResidualValue,
		MileageCap:                  r.MileageCap,
		Notes:                       r.Notes,
		ArchivedAt:                  r.ArchivedAt,
		ExternalID:                  r.ExternalID,
		CustomFields:                json.RawMessage(r.CustomFields),
		FuelVolumeUnits:             r.FuelVolumeUnits,
		CurrentMeterDate:            r.CurrentMeterDate,
		LoanAccountNumber:           r.LoanAccountNumber,
		LoanNotes:                   r.LoanNotes,
		LoanVendorID:                r.LoanVendorID,
		LoanStartedAt:               r.LoanStartedAt,
		LoanEndedAt:                 r.LoanEndedAt,
		UpdatedAt:                   r.UpdatedAt,
	}
}
