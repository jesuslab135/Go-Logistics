package inspection

import "time"

// InspectionFormItem — port of api/models/inspection_model.py:41
// TypeConfig shape varies by ItemType (NUMERIC: min/max/unit, DROPDOWN: options).

type InspectionFormItem struct {
	ID                                 int64
	FormID                             int64
	ItemType                           string
	Label                              string
	ShortDescription                   string
	Instructions                       string
	Position                           int32
	IsRequired                         bool
	PassLabel                          string
	FailLabel                          string
	NALabel                            string
	EnableNAOption                     bool
	RequireRemarkOnFail                bool
	RequireRemarkOnPass                bool
	RequirePhotoOnFail                 bool
	RequireMeterEntryPhotoVerification bool
	RequireSecondaryMeterIfOneExists   bool
	TypeConfig                         map[string]any
	CreatedAt                          time.Time
	UpdatedAt                          time.Time
}
