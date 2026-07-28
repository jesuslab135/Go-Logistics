package dto

import (
	"github.com/shopspring/decimal"
)

type UpsertVehicleRequest struct {
	EngineSerial            string           `json:"engine_serial" binding:"omitempty,max=100"`
	EngineDescription       string           `json:"engine_description" binding:"omitempty,max=200"`
	EngineBrand             string           `json:"engine_brand" binding:"omitempty,max=100"`
	EngineCylinders         *int32           `json:"engine_cylinders"`
	EngineDisplacement      string           `json:"engine_displacement" binding:"omitempty,max=50"`
	MaxHp                   string           `json:"max_hp" binding:"omitempty,max=50"`
	MaxTorque               string           `json:"max_torque" binding:"omitempty,max=50"`
	OilCapacity             string           `json:"oil_capacity" binding:"omitempty,max=50"`
	EngineAspiration        string           `json:"engine_aspiration" binding:"omitempty,max=50"`
	EngineBlockType         string           `json:"engine_block_type" binding:"omitempty,max=50"`
	EngineCompression       string           `json:"engine_compression" binding:"omitempty,max=50"`
	FuelInduction           string           `json:"fuel_induction" binding:"omitempty,max=50"`
	TransmissionDescription string           `json:"transmission_description" binding:"omitempty,max=200"`
	TransmissionBrand       string           `json:"transmission_brand" binding:"omitempty,max=100"`
	TransmissionType        string           `json:"transmission_type" binding:"omitempty,max=50"`
	TransmissionGears       *int32           `json:"transmission_gears"`
	DriveType               string           `json:"drive_type" binding:"omitempty,max=20"`
	BrakeSystem             string           `json:"brake_system" binding:"omitempty,max=20"`
	Axles                   int32            `json:"axles"`
	Differential            string           `json:"differential" binding:"omitempty,max=100"`
	DifferentialRatio       string           `json:"differential_ratio" binding:"omitempty,max=20"`
	Suspension              string           `json:"suspension" binding:"omitempty,max=100"`
	TireSize                string           `json:"tire_size" binding:"omitempty,max=50"`
	Height                  string           `json:"height" binding:"omitempty,max=50"`
	Length                  string           `json:"length" binding:"omitempty,max=50"`
	Width                   string           `json:"width" binding:"omitempty,max=50"`
	Wheelbase               string           `json:"wheelbase" binding:"omitempty,max=50"`
	CurbWeight              *decimal.Decimal `json:"curb_weight"`
	Gvwr                    *decimal.Decimal `json:"gvwr"`
	MaxWeightCapacity       *decimal.Decimal `json:"max_weight_capacity"`
	MaxPayload              *decimal.Decimal `json:"max_payload"`
	TowingCapacity          *decimal.Decimal `json:"towing_capacity"`
	FuelTankCapacity        *decimal.Decimal `json:"fuel_tank_capacity"`
	FuelTank2Capacity       *decimal.Decimal `json:"fuel_tank_2_capacity"`
	Tank1Security           string           `json:"tank_1_security" binding:"omitempty,max=10"`
	Tank2Security           string           `json:"tank_2_security" binding:"omitempty,max=10"`
	InteriorVolume          string           `json:"interior_volume" binding:"omitempty,max=50"`
	PassengerVolume         string           `json:"passenger_volume" binding:"omitempty,max=50"`
	GroundClearance         string           `json:"ground_clearance" binding:"omitempty,max=50"`
	EngineBore              string           `json:"engine_bore" binding:"omitempty,max=50"`
	RedlineRpm              string           `json:"redline_rpm" binding:"omitempty,max=50"`
	Stroke                  string           `json:"stroke" binding:"omitempty,max=50"`
	Valves                  *int32           `json:"valves"`
	FrontTrackWidth         string           `json:"front_track_width" binding:"omitempty,max=50"`
	RearTrackWidth          string           `json:"rear_track_width" binding:"omitempty,max=50"`
	FrontWheelDiameter      string           `json:"front_wheel_diameter" binding:"omitempty,max=50"`
	RearWheelDiameter       string           `json:"rear_wheel_diameter" binding:"omitempty,max=50"`
	FuelQuality             string           `json:"fuel_quality" binding:"omitempty,max=100"`
	CabType                 string           `json:"cab_type" binding:"omitempty,max=20"`
	TruckConfig             string           `json:"truck_config" binding:"omitempty,max=20"`
	TelematicsSystem        string           `json:"telematics_system" binding:"omitempty,max=100"`
	EmissionStandard        string           `json:"emission_standard" binding:"omitempty,max=50"`
	EmissionActive          *bool            `json:"emission_active"`
	SweetspotRpm            string           `json:"sweetspot_rpm" binding:"omitempty,max=50"`
	PedalSpeedLimit         *int32           `json:"pedal_speed_limit"`
	CruiseSpeedLimit        *int32           `json:"cruise_speed_limit"`
	IdleShutdown            string           `json:"idle_shutdown" binding:"omitempty,max=50"`
	ApuType                 string           `json:"apu_type" binding:"omitempty,max=50"`
	HasSpareTireRack        bool             `json:"has_spare_tire_rack"`
	FuelGroup               string           `json:"fuel_group" binding:"omitempty,max=100"`
	Cell                    string           `json:"cell" binding:"omitempty,max=100"`
	ServiceType             string           `json:"service_type" binding:"omitempty,max=100"`
	Supervisor              string           `json:"supervisor" binding:"omitempty,max=200"`
	Management              string           `json:"management" binding:"omitempty,max=200"`
	IsActiveDispatch        *bool            `json:"is_active_dispatch"`
	IsActiveCompany         *bool            `json:"is_active_company"`
	DutyType                string           `json:"duty_type" binding:"omitempty,max=20"`
	WeightClass             string           `json:"weight_class" binding:"omitempty,max=50"`
	CargoVolume             *decimal.Decimal `json:"cargo_volume"`
	BedLength               *decimal.Decimal `json:"bed_length"`
	FrontTirePsi            *decimal.Decimal `json:"front_tire_psi"`
	RearTirePsi             *decimal.Decimal `json:"rear_tire_psi"`
	FrontTireType           string           `json:"front_tire_type" binding:"omitempty,max=50"`
	RearTireType            string           `json:"rear_tire_type" binding:"omitempty,max=50"`
	RearAxleType            string           `json:"rear_axle_type" binding:"omitempty,max=100"`
	Operator                string           `json:"operator" binding:"omitempty,max=255"`
	EpaCity                 *decimal.Decimal `json:"epa_city"`
	EpaHighway              *decimal.Decimal `json:"epa_highway"`
	EpaCombined             *decimal.Decimal `json:"epa_combined"`
}

type VehicleResponse struct {
	AssetID                 int64            `json:"asset_id"`
	EngineSerial            string           `json:"engine_serial" binding:"omitempty,max=100"`
	EngineDescription       string           `json:"engine_description" binding:"omitempty,max=200"`
	EngineBrand             string           `json:"engine_brand" binding:"omitempty,max=100"`
	EngineCylinders         *int32           `json:"engine_cylinders"`
	EngineDisplacement      string           `json:"engine_displacement" binding:"omitempty,max=50"`
	MaxHp                   string           `json:"max_hp" binding:"omitempty,max=50"`
	MaxTorque               string           `json:"max_torque" binding:"omitempty,max=50"`
	OilCapacity             string           `json:"oil_capacity" binding:"omitempty,max=50"`
	EngineAspiration        string           `json:"engine_aspiration" binding:"omitempty,max=50"`
	EngineBlockType         string           `json:"engine_block_type" binding:"omitempty,max=50"`
	EngineCompression       string           `json:"engine_compression" binding:"omitempty,max=50"`
	FuelInduction           string           `json:"fuel_induction" binding:"omitempty,max=50"`
	TransmissionDescription string           `json:"transmission_description" binding:"omitempty,max=200"`
	TransmissionBrand       string           `json:"transmission_brand" binding:"omitempty,max=100"`
	TransmissionType        string           `json:"transmission_type" binding:"omitempty,max=50"`
	TransmissionGears       *int32           `json:"transmission_gears"`
	DriveType               string           `json:"drive_type" binding:"omitempty,max=20"`
	BrakeSystem             string           `json:"brake_system" binding:"omitempty,max=20"`
	Axles                   int32            `json:"axles"`
	Differential            string           `json:"differential" binding:"omitempty,max=100"`
	DifferentialRatio       string           `json:"differential_ratio" binding:"omitempty,max=20"`
	Suspension              string           `json:"suspension" binding:"omitempty,max=100"`
	TireSize                string           `json:"tire_size" binding:"omitempty,max=50"`
	Height                  string           `json:"height" binding:"omitempty,max=50"`
	Length                  string           `json:"length" binding:"omitempty,max=50"`
	Width                   string           `json:"width" binding:"omitempty,max=50"`
	Wheelbase               string           `json:"wheelbase" binding:"omitempty,max=50"`
	CurbWeight              *decimal.Decimal `json:"curb_weight"`
	Gvwr                    *decimal.Decimal `json:"gvwr"`
	MaxWeightCapacity       *decimal.Decimal `json:"max_weight_capacity"`
	MaxPayload              *decimal.Decimal `json:"max_payload"`
	TowingCapacity          *decimal.Decimal `json:"towing_capacity"`
	FuelTankCapacity        *decimal.Decimal `json:"fuel_tank_capacity"`
	FuelTank2Capacity       *decimal.Decimal `json:"fuel_tank_2_capacity"`
	Tank1Security           string           `json:"tank_1_security" binding:"omitempty,max=10"`
	Tank2Security           string           `json:"tank_2_security" binding:"omitempty,max=10"`
	InteriorVolume          string           `json:"interior_volume" binding:"omitempty,max=50"`
	PassengerVolume         string           `json:"passenger_volume" binding:"omitempty,max=50"`
	GroundClearance         string           `json:"ground_clearance" binding:"omitempty,max=50"`
	EngineBore              string           `json:"engine_bore" binding:"omitempty,max=50"`
	RedlineRpm              string           `json:"redline_rpm" binding:"omitempty,max=50"`
	Stroke                  string           `json:"stroke" binding:"omitempty,max=50"`
	Valves                  *int32           `json:"valves"`
	FrontTrackWidth         string           `json:"front_track_width" binding:"omitempty,max=50"`
	RearTrackWidth          string           `json:"rear_track_width" binding:"omitempty,max=50"`
	FrontWheelDiameter      string           `json:"front_wheel_diameter" binding:"omitempty,max=50"`
	RearWheelDiameter       string           `json:"rear_wheel_diameter" binding:"omitempty,max=50"`
	FuelQuality             string           `json:"fuel_quality" binding:"omitempty,max=100"`
	CabType                 string           `json:"cab_type" binding:"omitempty,max=20"`
	TruckConfig             string           `json:"truck_config" binding:"omitempty,max=20"`
	TelematicsSystem        string           `json:"telematics_system" binding:"omitempty,max=100"`
	EmissionStandard        string           `json:"emission_standard" binding:"omitempty,max=50"`
	EmissionActive          *bool            `json:"emission_active"`
	SweetspotRpm            string           `json:"sweetspot_rpm" binding:"omitempty,max=50"`
	PedalSpeedLimit         *int32           `json:"pedal_speed_limit"`
	CruiseSpeedLimit        *int32           `json:"cruise_speed_limit"`
	IdleShutdown            string           `json:"idle_shutdown" binding:"omitempty,max=50"`
	ApuType                 string           `json:"apu_type" binding:"omitempty,max=50"`
	HasSpareTireRack        bool             `json:"has_spare_tire_rack"`
	FuelGroup               string           `json:"fuel_group" binding:"omitempty,max=100"`
	Cell                    string           `json:"cell" binding:"omitempty,max=100"`
	ServiceType             string           `json:"service_type" binding:"omitempty,max=100"`
	Supervisor              string           `json:"supervisor" binding:"omitempty,max=200"`
	Management              string           `json:"management" binding:"omitempty,max=200"`
	IsActiveDispatch        bool             `json:"is_active_dispatch"`
	IsActiveCompany         bool             `json:"is_active_company"`
	DutyType                string           `json:"duty_type" binding:"omitempty,max=20"`
	WeightClass             string           `json:"weight_class" binding:"omitempty,max=50"`
	CargoVolume             *decimal.Decimal `json:"cargo_volume"`
	BedLength               *decimal.Decimal `json:"bed_length"`
	FrontTirePsi            *decimal.Decimal `json:"front_tire_psi"`
	RearTirePsi             *decimal.Decimal `json:"rear_tire_psi"`
	FrontTireType           string           `json:"front_tire_type" binding:"omitempty,max=50"`
	RearTireType            string           `json:"rear_tire_type" binding:"omitempty,max=50"`
	RearAxleType            string           `json:"rear_axle_type" binding:"omitempty,max=100"`
	Operator                string           `json:"operator" binding:"omitempty,max=255"`
	EpaCity                 *decimal.Decimal `json:"epa_city"`
	EpaHighway              *decimal.Decimal `json:"epa_highway"`
	EpaCombined             *decimal.Decimal `json:"epa_combined"`
}
