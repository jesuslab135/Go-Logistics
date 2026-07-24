package tire

import "github.com/shopspring/decimal"

// TireModel — port of api/models/tire_model.py:145

type TireModel struct {
	ID                     int64
	CompanyID              int64
	Brand                  string
	ModelName              string
	Size                   string
	FactoryTreadDepth32nds *int32
	MinimumTreadDepth32nds *int32
	LifeExpectancyMiles    *int32
	RecommendedPSI         *decimal.Decimal
}
