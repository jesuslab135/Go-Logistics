package asset

// AssetStatus — port of api/models/asset_model.py:18

type AssetStatus struct {
	ID        int64
	CompanyID int64
	Name      string
	ColorCode string
}
