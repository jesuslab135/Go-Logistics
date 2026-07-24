package asset

import "time"

// Trailer — port of api/models/asset_model.py:319
// Shared primary key with Asset, same as Vehicle.

type Trailer struct {
	Asset

	TrailerType          string
	Classification       string
	Classification2      string
	Size                 string
	Suspension           string
	OwnerName            string
	Financing            string
	Supplier             string
	LicensePlateUS       string
	LicensePlateUSState  string
	LicensePlateMX       string
	Doors                string
	Walls                string
	RailPost             string
	Skylight             bool
	FloorType            string
	RoofType             string
	HazmatType           string
	AeroKitType          string
	GPSProvider          string
	GPSSerial            string
	GPSSignalStatus      string
	GPSContractStart     *time.Time
	GPSContractEnd       *time.Time
	GPSContractReference string
	ContractStart        *time.Time
	ContractEnd          *time.Time
	ContractPeriod       string
	ContractReference    string
	FumigationDate       *time.Time
	FumigationCert       bool
	RegistrationDate     *time.Time
	WaterproofingDate    *time.Time
	DecommissionReason   string
	DecommissionDate     *time.Time
	OperationalUse       string
	OperationZone        string
}
