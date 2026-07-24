package dto

import (
	"github.com/shopspring/decimal"
)

type CreateTireModelRequest struct {
	Brand                  string           `json:"brand" binding:"omitempty,max=100"`
	ModelName              string           `json:"model_name" binding:"omitempty,max=150"`
	Size                   string           `json:"size" binding:"omitempty,max=50"`
	FactoryTreadDepth32nds *int32           `json:"factory_tread_depth_32nds"`
	MinimumTreadDepth32nds *int32           `json:"minimum_tread_depth_32nds"`
	LifeExpectancyMiles    *int32           `json:"life_expectancy_miles"`
	RecommendedPsi         *decimal.Decimal `json:"recommended_psi"`
}

type UpdateTireModelRequest struct {
	Brand                  string           `json:"brand" binding:"omitempty,max=100"`
	ModelName              string           `json:"model_name" binding:"omitempty,max=150"`
	Size                   string           `json:"size" binding:"omitempty,max=50"`
	FactoryTreadDepth32nds *int32           `json:"factory_tread_depth_32nds"`
	MinimumTreadDepth32nds *int32           `json:"minimum_tread_depth_32nds"`
	LifeExpectancyMiles    *int32           `json:"life_expectancy_miles"`
	RecommendedPsi         *decimal.Decimal `json:"recommended_psi"`
}

type TireModelResponse struct {
	ID                     int64            `json:"id"`
	CompanyID              int64            `json:"company_id"`
	Brand                  string           `json:"brand" binding:"omitempty,max=100"`
	ModelName              string           `json:"model_name" binding:"omitempty,max=150"`
	Size                   string           `json:"size" binding:"omitempty,max=50"`
	FactoryTreadDepth32nds *int32           `json:"factory_tread_depth_32nds"`
	MinimumTreadDepth32nds *int32           `json:"minimum_tread_depth_32nds"`
	LifeExpectancyMiles    *int32           `json:"life_expectancy_miles"`
	RecommendedPsi         *decimal.Decimal `json:"recommended_psi"`
}

type TireModelPage struct {
	Data    []TireModelResponse `json:"data"`
	Total   int64               `json:"total"`
	Limit   int                 `json:"limit"`
	Offset  int                 `json:"offset"`
	HasNext bool                `json:"has_next"`
}
