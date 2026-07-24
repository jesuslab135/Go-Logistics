package asset

// CatalogOption — port of api/models/asset_model.py:437

type CatalogOption struct {
	ID        int64
	CompanyID int64
	Category  string
	Value     string
}
