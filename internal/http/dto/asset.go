package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CreateAssetRequest struct {
	// archived_at is not writable: archiving is an action
	// (POST .../archive and .../restore), because a referenced record must be
	// archived rather than deleted and that is a decision, not a field.
	Name                        string           `json:"name" binding:"omitempty,max=100"`
	VinSn                       string           `json:"vin_sn" binding:"omitempty,max=100"`
	Msrp                        *decimal.Decimal `json:"msrp"`
	GenerateExpenses            bool             `json:"generate_expenses"`
	AssetTypeID                 *int64           `json:"asset_type_id"`
	StatusID                    *int64           `json:"status_id"`
	LeaseVendorID               *int64           `json:"lease_vendor_id"`
	VehicleType                 string           `json:"vehicle_type" binding:"omitempty,max=20"`
	OwnershipType               string           `json:"ownership_type" binding:"omitempty,max=20"`
	Labels                      json.RawMessage  `json:"labels"`
	LinkedVehicles              json.RawMessage  `json:"linked_vehicles"`
	LoanStartDate               *time.Time       `json:"loan_start_date"`
	LoanEndDate                 *time.Time       `json:"loan_end_date"`
	MonthlyPayment              *decimal.Decimal `json:"monthly_payment"`
	NumberOfPayments            *int32           `json:"number_of_payments"`
	LeaseNumber                 string           `json:"lease_number" binding:"omitempty,max=100"`
	LeaseStartDate              *time.Time       `json:"lease_start_date"`
	LeaseEndDate                *time.Time       `json:"lease_end_date"`
	ExcessMileageCharge         *decimal.Decimal `json:"excess_mileage_charge"`
	OwnerCompanyID              *int64           `json:"owner_company_id"`
	Year                        *int32           `json:"year"`
	Make                        string           `json:"make" binding:"omitempty,max=100"`
	Model                       string           `json:"model" binding:"omitempty,max=100"`
	Trim                        string           `json:"trim" binding:"omitempty,max=100"`
	Color                       string           `json:"color" binding:"omitempty,max=50"`
	LicensePlate                string           `json:"license_plate" binding:"omitempty,max=20"`
	Group                       string           `json:"group" binding:"omitempty,max=100"`
	Photo                       *string          `json:"photo" binding:"omitempty,max=500"`
	MeterUnit                   string           `json:"meter_unit" binding:"omitempty,max=5"`
	CurrentMeter                *decimal.Decimal `json:"current_meter"`
	SecondaryMeterUnit          string           `json:"secondary_meter_unit" binding:"omitempty,max=20"`
	SecondaryMeterValue         *decimal.Decimal `json:"secondary_meter_value"`
	FuelType                    string           `json:"fuel_type" binding:"omitempty,max=20"`
	BodyType                    string           `json:"body_type" binding:"omitempty,max=100"`
	BodySubtype                 string           `json:"body_subtype" binding:"omitempty,max=100"`
	RegistrationState           string           `json:"registration_state" binding:"omitempty,max=50"`
	PurchaseDate                *time.Time       `json:"purchase_date"`
	PurchasePrice               *decimal.Decimal `json:"purchase_price"`
	PurchaseVendor              string           `json:"purchase_vendor" binding:"omitempty,max=200"`
	PurchaseMeter               *decimal.Decimal `json:"purchase_meter"`
	InServiceDate               *time.Time       `json:"in_service_date"`
	InServiceMeter              *decimal.Decimal `json:"in_service_meter"`
	OutOfServiceDate            *time.Time       `json:"out_of_service_date"`
	OutOfServiceMeter           *decimal.Decimal `json:"out_of_service_meter"`
	EstimatedServiceMonths      *int32           `json:"estimated_service_months"`
	EstimatedReplacementMileage *int32           `json:"estimated_replacement_mileage"`
	EstimatedResalePrice        *decimal.Decimal `json:"estimated_resale_price"`
	AcquisitionType             string           `json:"acquisition_type" binding:"omitempty,max=20"`
	MonthlyCost                 *decimal.Decimal `json:"monthly_cost"`
	AcquisitionDate             *time.Time       `json:"acquisition_date"`
	LoanAmount                  *decimal.Decimal `json:"loan_amount"`
	CapitalizedCost             *decimal.Decimal `json:"capitalized_cost"`
	DownPayment                 *decimal.Decimal `json:"down_payment"`
	AnnualPercentageRate        *decimal.Decimal `json:"annual_percentage_rate"`
	FirstPaymentDate            *time.Time       `json:"first_payment_date"`
	ResidualValue               *decimal.Decimal `json:"residual_value"`
	MileageCap                  *int32           `json:"mileage_cap"`
	Notes                       string           `json:"notes"`
	ExternalID                  string           `json:"external_id" binding:"omitempty,max=100"`
	CustomFields                json.RawMessage  `json:"custom_fields" swaggertype:"object"`
	FuelVolumeUnits             string           `json:"fuel_volume_units" binding:"omitempty,max=20"`
	CurrentMeterDate            *time.Time       `json:"current_meter_date"`
	LoanAccountNumber           string           `json:"loan_account_number" binding:"omitempty,max=100"`
	LoanNotes                   string           `json:"loan_notes"`
	LoanVendorID                *int64           `json:"loan_vendor_id"`
	LoanStartedAt               *time.Time       `json:"loan_started_at"`
	LoanEndedAt                 *time.Time       `json:"loan_ended_at"`

	// Vehicle/Trailer, when present, are upserted in the SAME transaction as the
	// asset write: a failure on either rolls the whole request back, so no asset
	// row is ever persisted without its subtype. Omit both for an asset-only edit
	// (status, photo), which keeps the single-write path unchanged.
	Vehicle *UpsertVehicleRequest `json:"vehicle,omitempty"`
	Trailer *UpsertTrailerRequest `json:"trailer,omitempty"`
}

type UpdateAssetRequest struct {
	// archived_at is not writable: archiving is an action
	// (POST .../archive and .../restore), because a referenced record must be
	// archived rather than deleted and that is a decision, not a field.
	Name                        string           `json:"name" binding:"omitempty,max=100"`
	VinSn                       string           `json:"vin_sn" binding:"omitempty,max=100"`
	Msrp                        *decimal.Decimal `json:"msrp"`
	GenerateExpenses            bool             `json:"generate_expenses"`
	AssetTypeID                 *int64           `json:"asset_type_id"`
	StatusID                    *int64           `json:"status_id"`
	LeaseVendorID               *int64           `json:"lease_vendor_id"`
	VehicleType                 string           `json:"vehicle_type" binding:"omitempty,max=20"`
	OwnershipType               string           `json:"ownership_type" binding:"omitempty,max=20"`
	Labels                      json.RawMessage  `json:"labels"`
	LinkedVehicles              json.RawMessage  `json:"linked_vehicles"`
	LoanStartDate               *time.Time       `json:"loan_start_date"`
	LoanEndDate                 *time.Time       `json:"loan_end_date"`
	MonthlyPayment              *decimal.Decimal `json:"monthly_payment"`
	NumberOfPayments            *int32           `json:"number_of_payments"`
	LeaseNumber                 string           `json:"lease_number" binding:"omitempty,max=100"`
	LeaseStartDate              *time.Time       `json:"lease_start_date"`
	LeaseEndDate                *time.Time       `json:"lease_end_date"`
	ExcessMileageCharge         *decimal.Decimal `json:"excess_mileage_charge"`
	OwnerCompanyID              *int64           `json:"owner_company_id"`
	Year                        *int32           `json:"year"`
	Make                        string           `json:"make" binding:"omitempty,max=100"`
	Model                       string           `json:"model" binding:"omitempty,max=100"`
	Trim                        string           `json:"trim" binding:"omitempty,max=100"`
	Color                       string           `json:"color" binding:"omitempty,max=50"`
	LicensePlate                string           `json:"license_plate" binding:"omitempty,max=20"`
	Group                       string           `json:"group" binding:"omitempty,max=100"`
	Photo                       *string          `json:"photo" binding:"omitempty,max=500"`
	MeterUnit                   string           `json:"meter_unit" binding:"omitempty,max=5"`
	CurrentMeter                *decimal.Decimal `json:"current_meter"`
	SecondaryMeterUnit          string           `json:"secondary_meter_unit" binding:"omitempty,max=20"`
	SecondaryMeterValue         *decimal.Decimal `json:"secondary_meter_value"`
	FuelType                    string           `json:"fuel_type" binding:"omitempty,max=20"`
	BodyType                    string           `json:"body_type" binding:"omitempty,max=100"`
	BodySubtype                 string           `json:"body_subtype" binding:"omitempty,max=100"`
	RegistrationState           string           `json:"registration_state" binding:"omitempty,max=50"`
	PurchaseDate                *time.Time       `json:"purchase_date"`
	PurchasePrice               *decimal.Decimal `json:"purchase_price"`
	PurchaseVendor              string           `json:"purchase_vendor" binding:"omitempty,max=200"`
	PurchaseMeter               *decimal.Decimal `json:"purchase_meter"`
	InServiceDate               *time.Time       `json:"in_service_date"`
	InServiceMeter              *decimal.Decimal `json:"in_service_meter"`
	OutOfServiceDate            *time.Time       `json:"out_of_service_date"`
	OutOfServiceMeter           *decimal.Decimal `json:"out_of_service_meter"`
	EstimatedServiceMonths      *int32           `json:"estimated_service_months"`
	EstimatedReplacementMileage *int32           `json:"estimated_replacement_mileage"`
	EstimatedResalePrice        *decimal.Decimal `json:"estimated_resale_price"`
	AcquisitionType             string           `json:"acquisition_type" binding:"omitempty,max=20"`
	MonthlyCost                 *decimal.Decimal `json:"monthly_cost"`
	AcquisitionDate             *time.Time       `json:"acquisition_date"`
	LoanAmount                  *decimal.Decimal `json:"loan_amount"`
	CapitalizedCost             *decimal.Decimal `json:"capitalized_cost"`
	DownPayment                 *decimal.Decimal `json:"down_payment"`
	AnnualPercentageRate        *decimal.Decimal `json:"annual_percentage_rate"`
	FirstPaymentDate            *time.Time       `json:"first_payment_date"`
	ResidualValue               *decimal.Decimal `json:"residual_value"`
	MileageCap                  *int32           `json:"mileage_cap"`
	Notes                       string           `json:"notes"`
	ExternalID                  string           `json:"external_id" binding:"omitempty,max=100"`
	CustomFields                json.RawMessage  `json:"custom_fields" swaggertype:"object"`
	FuelVolumeUnits             string           `json:"fuel_volume_units" binding:"omitempty,max=20"`
	CurrentMeterDate            *time.Time       `json:"current_meter_date"`
	LoanAccountNumber           string           `json:"loan_account_number" binding:"omitempty,max=100"`
	LoanNotes                   string           `json:"loan_notes"`
	LoanVendorID                *int64           `json:"loan_vendor_id"`
	LoanStartedAt               *time.Time       `json:"loan_started_at"`
	LoanEndedAt                 *time.Time       `json:"loan_ended_at"`

	// Vehicle/Trailer, when present, are upserted in the SAME transaction as the
	// asset write: a failure on either rolls the whole request back, so no asset
	// row is ever persisted without its subtype. Omit both for an asset-only edit
	// (status, photo), which keeps the single-write path unchanged.
	Vehicle *UpsertVehicleRequest `json:"vehicle,omitempty"`
	Trailer *UpsertTrailerRequest `json:"trailer,omitempty"`
}

type AssetResponse struct {
	ID                          int64            `json:"id"`
	CompanyID                   int64            `json:"company_id"`
	Name                        string           `json:"name" binding:"omitempty,max=100"`
	VinSn                       string           `json:"vin_sn" binding:"omitempty,max=100"`
	Msrp                        *decimal.Decimal `json:"msrp"`
	GenerateExpenses            bool             `json:"generate_expenses"`
	AssetTypeID                 *int64           `json:"asset_type_id"`
	StatusID                    *int64           `json:"status_id"`
	LeaseVendorID               *int64           `json:"lease_vendor_id"`
	VehicleType                 string           `json:"vehicle_type" binding:"omitempty,max=20"`
	OwnershipType               string           `json:"ownership_type" binding:"omitempty,max=20"`
	Labels                      json.RawMessage  `json:"labels"`
	LinkedVehicles              json.RawMessage  `json:"linked_vehicles"`
	LoanStartDate               *time.Time       `json:"loan_start_date"`
	LoanEndDate                 *time.Time       `json:"loan_end_date"`
	MonthlyPayment              *decimal.Decimal `json:"monthly_payment"`
	NumberOfPayments            *int32           `json:"number_of_payments"`
	LeaseNumber                 string           `json:"lease_number" binding:"omitempty,max=100"`
	LeaseStartDate              *time.Time       `json:"lease_start_date"`
	LeaseEndDate                *time.Time       `json:"lease_end_date"`
	ExcessMileageCharge         *decimal.Decimal `json:"excess_mileage_charge"`
	OwnerCompanyID              *int64           `json:"owner_company_id"`
	Year                        *int32           `json:"year"`
	Make                        string           `json:"make" binding:"omitempty,max=100"`
	Model                       string           `json:"model" binding:"omitempty,max=100"`
	Trim                        string           `json:"trim" binding:"omitempty,max=100"`
	Color                       string           `json:"color" binding:"omitempty,max=50"`
	LicensePlate                string           `json:"license_plate" binding:"omitempty,max=20"`
	Group                       string           `json:"group" binding:"omitempty,max=100"`
	Photo                       *string          `json:"photo" binding:"omitempty,max=500"`
	MeterUnit                   string           `json:"meter_unit" binding:"omitempty,max=5"`
	CurrentMeter                *decimal.Decimal `json:"current_meter"`
	SecondaryMeterUnit          string           `json:"secondary_meter_unit" binding:"omitempty,max=20"`
	SecondaryMeterValue         *decimal.Decimal `json:"secondary_meter_value"`
	FuelType                    string           `json:"fuel_type" binding:"omitempty,max=20"`
	BodyType                    string           `json:"body_type" binding:"omitempty,max=100"`
	BodySubtype                 string           `json:"body_subtype" binding:"omitempty,max=100"`
	RegistrationState           string           `json:"registration_state" binding:"omitempty,max=50"`
	PurchaseDate                *time.Time       `json:"purchase_date"`
	PurchasePrice               *decimal.Decimal `json:"purchase_price"`
	PurchaseVendor              string           `json:"purchase_vendor" binding:"omitempty,max=200"`
	PurchaseMeter               *decimal.Decimal `json:"purchase_meter"`
	InServiceDate               *time.Time       `json:"in_service_date"`
	InServiceMeter              *decimal.Decimal `json:"in_service_meter"`
	OutOfServiceDate            *time.Time       `json:"out_of_service_date"`
	OutOfServiceMeter           *decimal.Decimal `json:"out_of_service_meter"`
	EstimatedServiceMonths      *int32           `json:"estimated_service_months"`
	EstimatedReplacementMileage *int32           `json:"estimated_replacement_mileage"`
	EstimatedResalePrice        *decimal.Decimal `json:"estimated_resale_price"`
	AcquisitionType             string           `json:"acquisition_type" binding:"omitempty,max=20"`
	MonthlyCost                 *decimal.Decimal `json:"monthly_cost"`
	AcquisitionDate             *time.Time       `json:"acquisition_date"`
	LoanAmount                  *decimal.Decimal `json:"loan_amount"`
	CapitalizedCost             *decimal.Decimal `json:"capitalized_cost"`
	DownPayment                 *decimal.Decimal `json:"down_payment"`
	AnnualPercentageRate        *decimal.Decimal `json:"annual_percentage_rate"`
	FirstPaymentDate            *time.Time       `json:"first_payment_date"`
	ResidualValue               *decimal.Decimal `json:"residual_value"`
	MileageCap                  *int32           `json:"mileage_cap"`
	Notes                       string           `json:"notes"`
	ArchivedAt                  *time.Time       `json:"archived_at"`
	ExternalID                  string           `json:"external_id" binding:"omitempty,max=100"`
	CustomFields                json.RawMessage  `json:"custom_fields" swaggertype:"object"`
	FuelVolumeUnits             string           `json:"fuel_volume_units" binding:"omitempty,max=20"`
	CurrentMeterDate            *time.Time       `json:"current_meter_date"`
	LoanAccountNumber           string           `json:"loan_account_number" binding:"omitempty,max=100"`
	LoanNotes                   string           `json:"loan_notes"`
	LoanVendorID                *int64           `json:"loan_vendor_id"`
	LoanStartedAt               *time.Time       `json:"loan_started_at"`
	LoanEndedAt                 *time.Time       `json:"loan_ended_at"`
	UpdatedAt                   time.Time        `json:"updated_at"`

	// Denormalized from the 1:1 subtype rows so a registry can be drawn from
	// one paginated request. Null for assets without that subtype; the
	// authoritative records stay at /assets/{id}/vehicle and /assets/{id}/trailer.
	Operator              *string `json:"operator"`
	TrailerType           *string `json:"trailer_type"`
	TrailerClassification *string `json:"trailer_classification"`
	TrailerSize           *string `json:"trailer_size"`
}

type AssetPage struct {
	Data    []AssetResponse `json:"data"`
	Total   int64           `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasNext bool            `json:"has_next"`
}
