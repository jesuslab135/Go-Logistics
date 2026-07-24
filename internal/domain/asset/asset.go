package asset

import (
	"time"

	"github.com/shopspring/decimal"
)

// Asset — port of api/models/asset_model.py:32
// GenerateExpenses and NumberOfPayments are each declared twice in the Django
// source (:37/:46 and :79/:148); Python keeps the last, so each is one column.

type Asset struct {
	ID                          int64
	CompanyID                   int64
	Name                        string
	VinSN                       string
	MSRP                        *decimal.Decimal
	GenerateExpenses            bool
	AssetTypeID                 *int64
	StatusID                    *int64
	LeaseVendorID               *int64
	VehicleType                 string
	OwnershipType               string
	Labels                      []string
	LinkedVehicles              []int64
	LoanStartDate               *time.Time
	LoanEndDate                 *time.Time
	MonthlyPayment              *decimal.Decimal
	NumberOfPayments            *int32
	LeaseNumber                 string
	LeaseStartDate              *time.Time
	LeaseEndDate                *time.Time
	ExcessMileageCharge         *decimal.Decimal
	OwnerCompanyID              *int64
	Year                        *int32
	Make                        string
	Model                       string
	Trim                        string
	Color                       string
	LicensePlate                string
	Group                       string
	Photo                       *string
	MeterUnit                   string
	CurrentMeter                *decimal.Decimal
	SecondaryMeterUnit          string
	SecondaryMeterValue         *decimal.Decimal
	FuelType                    string
	BodyType                    string
	BodySubtype                 string
	RegistrationState           string
	PurchaseDate                *time.Time
	PurchasePrice               *decimal.Decimal
	PurchaseVendor              string
	PurchaseMeter               *decimal.Decimal
	InServiceDate               *time.Time
	InServiceMeter              *decimal.Decimal
	OutOfServiceDate            *time.Time
	OutOfServiceMeter           *decimal.Decimal
	EstimatedServiceMonths      *int32
	EstimatedReplacementMileage *int32
	EstimatedResalePrice        *decimal.Decimal
	AcquisitionType             string
	MonthlyCost                 *decimal.Decimal
	AcquisitionDate             *time.Time
	LoanAmount                  *decimal.Decimal
	CapitalizedCost             *decimal.Decimal
	DownPayment                 *decimal.Decimal
	AnnualPercentageRate        *decimal.Decimal
	FirstPaymentDate            *time.Time
	ResidualValue               *decimal.Decimal
	MileageCap                  *int32
	Notes                       string
	ArchivedAt                  *time.Time
	ExternalID                  string
	CustomFields                map[string]any
	FuelVolumeUnits             string
	CurrentMeterDate            *time.Time
	LoanAccountNumber           string
	LoanNotes                   string
	LoanVendorID                *int64
	LoanStartedAt               *time.Time
	LoanEndedAt                 *time.Time
	UpdatedAt                   time.Time
}
