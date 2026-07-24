package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateTireInstallationRequest struct {
	VehicleID                int64            `json:"vehicle_id"`
	PositionCode             string           `json:"position_code" binding:"omitempty,max=10"`
	InstallDate              time.Time        `json:"install_date"`
	OdometerAtInstall        int32            `json:"odometer_at_install"`
	TreadDepthAtInstall32nds *int32           `json:"tread_depth_at_install_32nds"`
	PsiAtInstall             *decimal.Decimal `json:"psi_at_install"`
	InstalledByID            *int64           `json:"installed_by_id"`
	Notes                    string           `json:"notes"`
}

type UpdateTireInstallationRequest struct {
	VehicleID                int64            `json:"vehicle_id"`
	PositionCode             string           `json:"position_code" binding:"omitempty,max=10"`
	InstallDate              time.Time        `json:"install_date"`
	OdometerAtInstall        int32            `json:"odometer_at_install"`
	TreadDepthAtInstall32nds *int32           `json:"tread_depth_at_install_32nds"`
	PsiAtInstall             *decimal.Decimal `json:"psi_at_install"`
	InstalledByID            *int64           `json:"installed_by_id"`
	Notes                    string           `json:"notes"`
}

type TireInstallationResponse struct {
	ID                       int64            `json:"id"`
	VehicleID                int64            `json:"vehicle_id"`
	TireID                   int64            `json:"tire_id"`
	PositionCode             string           `json:"position_code" binding:"omitempty,max=10"`
	InstallDate              time.Time        `json:"install_date"`
	OdometerAtInstall        int32            `json:"odometer_at_install"`
	TreadDepthAtInstall32nds *int32           `json:"tread_depth_at_install_32nds"`
	PsiAtInstall             *decimal.Decimal `json:"psi_at_install"`
	InstalledByID            *int64           `json:"installed_by_id"`
	Notes                    string           `json:"notes"`
}

type TireInstallationPage struct {
	Data    []TireInstallationResponse `json:"data"`
	Total   int64                      `json:"total"`
	Limit   int                        `json:"limit"`
	Offset  int                        `json:"offset"`
	HasNext bool                       `json:"has_next"`
}
