package asset

import "time"

// AssetTrailerAssignment — port of api/models/asset_model.py:393
// AssetID is the tractor, TrailerID is another Asset row carrying a Trailer.

type AssetTrailerAssignment struct {
	ID             int64
	AssetID        int64
	TrailerID      int64
	Position       int32
	AssignedDate   time.Time
	UnassignedDate *time.Time
	AssignedByID   *int64
	IsActive       bool
	Notes          string
}
