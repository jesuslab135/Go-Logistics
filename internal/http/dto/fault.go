package dto

type CreateFaultRequest struct {
	Family              string   `json:"family" binding:"omitempty,max=20"`
	Code                string   `json:"code" binding:"omitempty,max=50"`
	Name                string   `json:"name" binding:"omitempty,max=255"`
	Description         string   `json:"description"`
	AppliesToAssetTypes []string `json:"applies_to_asset_types" binding:"omitempty,max=20"`
}

type UpdateFaultRequest struct {
	Family              string   `json:"family" binding:"omitempty,max=20"`
	Code                string   `json:"code" binding:"omitempty,max=50"`
	Name                string   `json:"name" binding:"omitempty,max=255"`
	Description         string   `json:"description"`
	AppliesToAssetTypes []string `json:"applies_to_asset_types" binding:"omitempty,max=20"`
}

type FaultResponse struct {
	ID                  int64    `json:"id"`
	CompanyID           int64    `json:"company_id"`
	Family              string   `json:"family" binding:"omitempty,max=20"`
	Code                string   `json:"code" binding:"omitempty,max=50"`
	Name                string   `json:"name" binding:"omitempty,max=255"`
	Description         string   `json:"description"`
	AppliesToAssetTypes []string `json:"applies_to_asset_types" binding:"omitempty,max=20"`
}

type FaultPage struct {
	Data    []FaultResponse `json:"data"`
	Total   int64           `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasNext bool            `json:"has_next"`
}
