package asset

// AssetType — port of api/models/asset_model.py:4
// DEPRECATED upstream: classification now lives in Vehicle / Trailer.TrailerType.

type AssetType struct {
	ID          int64
	CompanyID   int64
	Name        string
	Category    string
	Description string
}
