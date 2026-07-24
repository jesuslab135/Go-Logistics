package dto

// --- AssetType ---

type CreateAssetTypeRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Category    string `json:"category" binding:"max=50"`
	Description string `json:"description"`
}

type UpdateAssetTypeRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Category    string `json:"category" binding:"max=50"`
	Description string `json:"description"`
}

type AssetTypeResponse struct {
	ID          int64  `json:"id"`
	CompanyID   int64  `json:"company_id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type AssetTypePage struct {
	Data    []AssetTypeResponse `json:"data"`
	Total   int64               `json:"total"`
	Limit   int                 `json:"limit"`
	Offset  int                 `json:"offset"`
	HasNext bool                `json:"has_next"`
}

// --- AssetStatus ---

type CreateAssetStatusRequest struct {
	Name      string `json:"name" binding:"required,max=50"`
	ColorCode string `json:"color_code" binding:"max=7"`
}

type UpdateAssetStatusRequest struct {
	Name      string `json:"name" binding:"required,max=50"`
	ColorCode string `json:"color_code" binding:"max=7"`
}

type AssetStatusResponse struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"company_id"`
	Name      string `json:"name"`
	ColorCode string `json:"color_code"`
}

type AssetStatusPage struct {
	Data    []AssetStatusResponse `json:"data"`
	Total   int64                 `json:"total"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
	HasNext bool                  `json:"has_next"`
}

// --- CatalogOption ---

type CreateCatalogOptionRequest struct {
	Category string `json:"category" binding:"required,max=30"`
	Value    string `json:"value" binding:"required,max=100"`
}

type UpdateCatalogOptionRequest struct {
	Category string `json:"category" binding:"required,max=30"`
	Value    string `json:"value" binding:"required,max=100"`
}

type CatalogOptionResponse struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"company_id"`
	Category  string `json:"category"`
	Value     string `json:"value"`
}

type CatalogOptionPage struct {
	Data    []CatalogOptionResponse `json:"data"`
	Total   int64                   `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	HasNext bool                    `json:"has_next"`
}
