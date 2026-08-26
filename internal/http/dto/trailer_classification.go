package dto

// A company's vocabulary for classifying trailers. It replaced two free-text
// columns, where "Dry Van", "DRY VAN" and "dry van" were three classifications
// as far as any filter or report was concerned.

type CreateTrailerClassificationRequest struct {
	Name string `json:"name" binding:"required,max=100"`
	// Position orders the picker. Ties fall back to name.
	Position int32 `json:"position"`
}

type UpdateTrailerClassificationRequest struct {
	Name     string `json:"name" binding:"required,max=100"`
	Position int32  `json:"position"`
}

type TrailerClassificationResponse struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"company_id"`
	Name      string `json:"name"`
	Position  int32  `json:"position"`
}

type TrailerClassificationPage struct {
	Data    []TrailerClassificationResponse `json:"data"`
	Total   int64                           `json:"total"`
	Limit   int                             `json:"limit"`
	Offset  int                             `json:"offset"`
	HasNext bool                            `json:"has_next"`
}
