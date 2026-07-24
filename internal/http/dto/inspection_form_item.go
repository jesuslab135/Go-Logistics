package dto

import (
	"encoding/json"
	"time"
)

type CreateInspectionFormItemRequest struct {
	ItemType                           string          `json:"item_type" binding:"omitempty,max=20"`
	Label                              string          `json:"label" binding:"omitempty,max=255"`
	ShortDescription                   string          `json:"short_description" binding:"omitempty,max=500"`
	Instructions                       string          `json:"instructions"`
	Position                           int32           `json:"position"`
	IsRequired                         bool            `json:"is_required"`
	PassLabel                          string          `json:"pass_label" binding:"omitempty,max=50"`
	FailLabel                          string          `json:"fail_label" binding:"omitempty,max=50"`
	NaLabel                            string          `json:"na_label" binding:"omitempty,max=50"`
	EnableNaOption                     bool            `json:"enable_na_option"`
	RequireRemarkOnFail                bool            `json:"require_remark_on_fail"`
	RequireRemarkOnPass                bool            `json:"require_remark_on_pass"`
	RequirePhotoOnFail                 bool            `json:"require_photo_on_fail"`
	RequireMeterEntryPhotoVerification bool            `json:"require_meter_entry_photo_verification"`
	RequireSecondaryMeterIfOneExists   bool            `json:"require_secondary_meter_if_one_exists"`
	TypeConfig                         json.RawMessage `json:"type_config"`
}

type UpdateInspectionFormItemRequest struct {
	ItemType                           string          `json:"item_type" binding:"omitempty,max=20"`
	Label                              string          `json:"label" binding:"omitempty,max=255"`
	ShortDescription                   string          `json:"short_description" binding:"omitempty,max=500"`
	Instructions                       string          `json:"instructions"`
	Position                           int32           `json:"position"`
	IsRequired                         bool            `json:"is_required"`
	PassLabel                          string          `json:"pass_label" binding:"omitempty,max=50"`
	FailLabel                          string          `json:"fail_label" binding:"omitempty,max=50"`
	NaLabel                            string          `json:"na_label" binding:"omitempty,max=50"`
	EnableNaOption                     bool            `json:"enable_na_option"`
	RequireRemarkOnFail                bool            `json:"require_remark_on_fail"`
	RequireRemarkOnPass                bool            `json:"require_remark_on_pass"`
	RequirePhotoOnFail                 bool            `json:"require_photo_on_fail"`
	RequireMeterEntryPhotoVerification bool            `json:"require_meter_entry_photo_verification"`
	RequireSecondaryMeterIfOneExists   bool            `json:"require_secondary_meter_if_one_exists"`
	TypeConfig                         json.RawMessage `json:"type_config"`
}

type InspectionFormItemResponse struct {
	ID                                 int64           `json:"id"`
	FormID                             int64           `json:"form_id"`
	ItemType                           string          `json:"item_type" binding:"omitempty,max=20"`
	Label                              string          `json:"label" binding:"omitempty,max=255"`
	ShortDescription                   string          `json:"short_description" binding:"omitempty,max=500"`
	Instructions                       string          `json:"instructions"`
	Position                           int32           `json:"position"`
	IsRequired                         bool            `json:"is_required"`
	PassLabel                          string          `json:"pass_label" binding:"omitempty,max=50"`
	FailLabel                          string          `json:"fail_label" binding:"omitempty,max=50"`
	NaLabel                            string          `json:"na_label" binding:"omitempty,max=50"`
	EnableNaOption                     bool            `json:"enable_na_option"`
	RequireRemarkOnFail                bool            `json:"require_remark_on_fail"`
	RequireRemarkOnPass                bool            `json:"require_remark_on_pass"`
	RequirePhotoOnFail                 bool            `json:"require_photo_on_fail"`
	RequireMeterEntryPhotoVerification bool            `json:"require_meter_entry_photo_verification"`
	RequireSecondaryMeterIfOneExists   bool            `json:"require_secondary_meter_if_one_exists"`
	TypeConfig                         json.RawMessage `json:"type_config"`
	CreatedAt                          time.Time       `json:"created_at"`
	UpdatedAt                          time.Time       `json:"updated_at"`
}

type InspectionFormItemPage struct {
	Data    []InspectionFormItemResponse `json:"data"`
	Total   int64                        `json:"total"`
	Limit   int                          `json:"limit"`
	Offset  int                          `json:"offset"`
	HasNext bool                         `json:"has_next"`
}
