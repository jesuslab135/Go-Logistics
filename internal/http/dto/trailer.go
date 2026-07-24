package dto

import (
	"time"
)

type UpsertTrailerRequest struct {
	TrailerType          string     `json:"trailer_type" binding:"omitempty,max=20"`
	Classification       string     `json:"classification" binding:"omitempty,max=100"`
	Classification2      string     `json:"classification_2" binding:"omitempty,max=100"`
	Size                 string     `json:"size" binding:"omitempty,max=20"`
	Suspension           string     `json:"suspension" binding:"omitempty,max=100"`
	OwnerName            string     `json:"owner_name" binding:"omitempty,max=200"`
	Financing            string     `json:"financing" binding:"omitempty,max=20"`
	Supplier             string     `json:"supplier" binding:"omitempty,max=200"`
	LicensePlateUs       string     `json:"license_plate_us" binding:"omitempty,max=20"`
	LicensePlateUsState  string     `json:"license_plate_us_state" binding:"omitempty,max=10"`
	LicensePlateMx       string     `json:"license_plate_mx" binding:"omitempty,max=20"`
	Doors                string     `json:"doors" binding:"omitempty,max=50"`
	Walls                string     `json:"walls" binding:"omitempty,max=50"`
	RailPost             string     `json:"rail_post" binding:"omitempty,max=50"`
	Skylight             bool       `json:"skylight"`
	FloorType            string     `json:"floor_type" binding:"omitempty,max=50"`
	RoofType             string     `json:"roof_type" binding:"omitempty,max=50"`
	HazmatType           string     `json:"hazmat_type" binding:"omitempty,max=100"`
	AeroKitType          string     `json:"aero_kit_type" binding:"omitempty,max=50"`
	GpsProvider          string     `json:"gps_provider" binding:"omitempty,max=100"`
	GpsSerial            string     `json:"gps_serial" binding:"omitempty,max=100"`
	GpsSignalStatus      string     `json:"gps_signal_status" binding:"omitempty,max=50"`
	GpsContractStart     *time.Time `json:"gps_contract_start"`
	GpsContractEnd       *time.Time `json:"gps_contract_end"`
	GpsContractReference string     `json:"gps_contract_reference" binding:"omitempty,max=200"`
	ContractStart        *time.Time `json:"contract_start"`
	ContractEnd          *time.Time `json:"contract_end"`
	ContractPeriod       string     `json:"contract_period" binding:"omitempty,max=50"`
	ContractReference    string     `json:"contract_reference" binding:"omitempty,max=100"`
	FumigationDate       *time.Time `json:"fumigation_date"`
	FumigationCert       bool       `json:"fumigation_cert"`
	RegistrationDate     *time.Time `json:"registration_date"`
	WaterproofingDate    *time.Time `json:"waterproofing_date"`
	DecommissionReason   string     `json:"decommission_reason"`
	DecommissionDate     *time.Time `json:"decommission_date"`
	OperationalUse       string     `json:"operational_use" binding:"omitempty,max=20"`
	OperationZone        string     `json:"operation_zone" binding:"omitempty,max=100"`
}

type TrailerResponse struct {
	AssetID              int64      `json:"asset_id"`
	TrailerType          string     `json:"trailer_type" binding:"omitempty,max=20"`
	Classification       string     `json:"classification" binding:"omitempty,max=100"`
	Classification2      string     `json:"classification_2" binding:"omitempty,max=100"`
	Size                 string     `json:"size" binding:"omitempty,max=20"`
	Suspension           string     `json:"suspension" binding:"omitempty,max=100"`
	OwnerName            string     `json:"owner_name" binding:"omitempty,max=200"`
	Financing            string     `json:"financing" binding:"omitempty,max=20"`
	Supplier             string     `json:"supplier" binding:"omitempty,max=200"`
	LicensePlateUs       string     `json:"license_plate_us" binding:"omitempty,max=20"`
	LicensePlateUsState  string     `json:"license_plate_us_state" binding:"omitempty,max=10"`
	LicensePlateMx       string     `json:"license_plate_mx" binding:"omitempty,max=20"`
	Doors                string     `json:"doors" binding:"omitempty,max=50"`
	Walls                string     `json:"walls" binding:"omitempty,max=50"`
	RailPost             string     `json:"rail_post" binding:"omitempty,max=50"`
	Skylight             bool       `json:"skylight"`
	FloorType            string     `json:"floor_type" binding:"omitempty,max=50"`
	RoofType             string     `json:"roof_type" binding:"omitempty,max=50"`
	HazmatType           string     `json:"hazmat_type" binding:"omitempty,max=100"`
	AeroKitType          string     `json:"aero_kit_type" binding:"omitempty,max=50"`
	GpsProvider          string     `json:"gps_provider" binding:"omitempty,max=100"`
	GpsSerial            string     `json:"gps_serial" binding:"omitempty,max=100"`
	GpsSignalStatus      string     `json:"gps_signal_status" binding:"omitempty,max=50"`
	GpsContractStart     *time.Time `json:"gps_contract_start"`
	GpsContractEnd       *time.Time `json:"gps_contract_end"`
	GpsContractReference string     `json:"gps_contract_reference" binding:"omitempty,max=200"`
	ContractStart        *time.Time `json:"contract_start"`
	ContractEnd          *time.Time `json:"contract_end"`
	ContractPeriod       string     `json:"contract_period" binding:"omitempty,max=50"`
	ContractReference    string     `json:"contract_reference" binding:"omitempty,max=100"`
	FumigationDate       *time.Time `json:"fumigation_date"`
	FumigationCert       bool       `json:"fumigation_cert"`
	RegistrationDate     *time.Time `json:"registration_date"`
	WaterproofingDate    *time.Time `json:"waterproofing_date"`
	DecommissionReason   string     `json:"decommission_reason"`
	DecommissionDate     *time.Time `json:"decommission_date"`
	OperationalUse       string     `json:"operational_use" binding:"omitempty,max=20"`
	OperationZone        string     `json:"operation_zone" binding:"omitempty,max=100"`
}
