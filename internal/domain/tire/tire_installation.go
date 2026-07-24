package tire

import (
	"time"

	"github.com/shopspring/decimal"
)

// TireInstallation — port of api/models/tire_model.py:292
// VehicleID references Asset. TireID is a OneToOne, so a tire can only be
// installed in one place at a time.

type TireInstallation struct {
	ID                       int64
	VehicleID                int64
	TireID                   int64
	PositionCode             string
	InstallDate              time.Time
	OdometerAtInstall        int32
	TreadDepthAtInstall32nds *int32
	PSIAtInstall             *decimal.Decimal
	InstalledByID            *int64
	Notes                    string
}
