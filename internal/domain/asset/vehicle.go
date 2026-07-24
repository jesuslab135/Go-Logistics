package asset

import "github.com/shopspring/decimal"

// Vehicle — port of api/models/asset_model.py:186
// Shared primary key: api_vehicle.asset_id is both PK and FK, so Asset is
// embedded and the two must always be loaded together via join.

type Vehicle struct {
	Asset

	EngineSerial            string
	EngineDescription       string
	EngineBrand             string
	EngineCylinders         *int32
	EngineDisplacement      string
	MaxHP                   string
	MaxTorque               string
	OilCapacity             string
	EngineAspiration        string
	EngineBlockType         string
	EngineCompression       string
	FuelInduction           string
	TransmissionDescription string
	TransmissionBrand       string
	TransmissionType        string
	TransmissionGears       *int32
	DriveType               string
	BrakeSystem             string
	Axles                   int32
	Differential            string
	DifferentialRatio       string
	Suspension              string
	TireSize                string
	Height                  string
	Length                  string
	Width                   string
	Wheelbase               string
	CurbWeight              *decimal.Decimal
	GVWR                    *decimal.Decimal
	MaxWeightCapacity       *decimal.Decimal
	MaxPayload              *decimal.Decimal
	TowingCapacity          *decimal.Decimal
	FuelTankCapacity        *decimal.Decimal
	FuelTank2Capacity       *decimal.Decimal
	Tank1Security           string
	Tank2Security           string
	InteriorVolume          string
	PassengerVolume         string
	GroundClearance         string
	EngineBore              string
	RedlineRPM              string
	Stroke                  string
	Valves                  *int32
	FrontTrackWidth         string
	RearTrackWidth          string
	FrontWheelDiameter      string
	RearWheelDiameter       string
	FuelQuality             string
	CabType                 string
	TruckConfig             string
	TelematicsSystem        string
	EmissionStandard        string
	EmissionActive          *bool
	SweetspotRPM            string
	PedalSpeedLimit         *int32
	CruiseSpeedLimit        *int32
	IdleShutdown            string
	APUType                 string
	HasSpareTireRack        bool
	FuelGroup               string
	Cell                    string
	ServiceType             string
	Supervisor              string
	Management              string
	IsActiveDispatch        bool
	IsActiveCompany         bool
	DutyType                string
	WeightClass             string
	CargoVolume             *decimal.Decimal
	BedLength               *decimal.Decimal
	FrontTirePSI            *decimal.Decimal
	RearTirePSI             *decimal.Decimal
	FrontTireType           string
	RearTireType            string
	RearAxleType            string
	Operator                string
	EPACity                 *decimal.Decimal
	EPAHighway              *decimal.Decimal
	EPACombined             *decimal.Decimal
}
