package dto

import "time"

// A custom field definition describes what may go in one key of a resource's
// custom_fields document. Without them the column was a free-for-all, and the
// only way to build a form over it was a raw JSON editor — which is not a
// business UI, it is the problem moved downstream.

type CreateCustomFieldDefinitionRequest struct {
	// Resource is the API's own plural name: "assets", "parts", "work-orders".
	Resource string `json:"resource" binding:"required,max=50"`
	// Key is what the value is stored under. It cannot be changed afterwards:
	// every stored document is keyed by it.
	Key   string `json:"key" binding:"required,max=50"`
	Label string `json:"label" binding:"required,max=100"`
	// FieldType is text, number, date, boolean or select.
	FieldType string `json:"field_type" binding:"required,max=20"`
	Required  bool   `json:"required"`
	// Options are the permitted values for a select, and are ignored otherwise.
	Options  []string `json:"options"`
	Position int32    `json:"position"`
}

// UpdateCustomFieldDefinitionRequest has no key or resource: changing either
// would orphan every value already stored under the old one.
type UpdateCustomFieldDefinitionRequest struct {
	Label     string   `json:"label" binding:"required,max=100"`
	FieldType string   `json:"field_type" binding:"required,max=20"`
	Required  bool     `json:"required"`
	Options   []string `json:"options"`
	Position  int32    `json:"position"`
}

type CustomFieldDefinitionResponse struct {
	ID        int64    `json:"id"`
	CompanyID int64    `json:"company_id"`
	Resource  string   `json:"resource"`
	Key       string   `json:"key"`
	Label     string   `json:"label"`
	FieldType string   `json:"field_type"`
	Required  bool     `json:"required"`
	Options   []string `json:"options"`
	Position  int32    `json:"position"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CustomFieldDefinitionPage struct {
	Data    []CustomFieldDefinitionResponse `json:"data"`
	Total   int64                           `json:"total"`
	Limit   int                             `json:"limit"`
	Offset  int                             `json:"offset"`
	HasNext bool                            `json:"has_next"`
}
