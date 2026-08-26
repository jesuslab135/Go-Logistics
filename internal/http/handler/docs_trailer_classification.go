package handler

// docListTrailerClassifications godoc
//
//	@Summary		List trailer classifications
//	@Description	The company's own vocabulary for classifying trailers. It replaced two free-text columns, where "Dry Van", "DRY VAN" and "dry van" counted as three classifications to every filter and report. Each company's catalog was seeded from the values that company was already using; no list is seeded on top of that.
//	@Tags			trailer-classifications
//	@Security		BearerAuth
//	@Produce		json
//	@Param			limit	query	int	false	"Page size"
//	@Param			offset	query	int	false	"Offset"
//	@Success		200		{object}	dto.TrailerClassificationPage
//	@Failure		401		{object}	dto.ErrorResponse
//	@Router			/api/v1/trailer-classifications [get]
func docListTrailerClassifications() {}

// docCreateTrailerClassification godoc
//
//	@Summary	Create a trailer classification
//	@Tags		trailer-classifications
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		body	body	dto.CreateTrailerClassificationRequest	true	"body"
//	@Success	201		{object}	dto.TrailerClassificationResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	409		{object}	dto.ErrorResponse
//	@Router		/api/v1/trailer-classifications [post]
func docCreateTrailerClassification() {}

// docGetTrailerClassification godoc
//
//	@Summary	Get a trailer classification
//	@Tags		trailer-classifications
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	int	true	"id"
//	@Success	200	{object}	dto.TrailerClassificationResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/api/v1/trailer-classifications/{id} [get]
func docGetTrailerClassification() {}

// docUpdateTrailerClassification godoc
//
//	@Summary	Update a trailer classification
//	@Tags		trailer-classifications
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int										true	"id"
//	@Param		body	body	dto.UpdateTrailerClassificationRequest	true	"body"
//	@Success	200		{object}	dto.TrailerClassificationResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/api/v1/trailer-classifications/{id} [put]
func docUpdateTrailerClassification() {}

// docDeleteTrailerClassification godoc
//
//	@Summary		Delete a trailer classification
//	@Description	Trailers referencing it become unclassified rather than blocking the delete: retiring a term nobody uses should not require re-classifying every trailer that ever carried it.
//	@Tags			trailer-classifications
//	@Security		BearerAuth
//	@Param			id	path	int	true	"id"
//	@Success		204	"No Content"
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/api/v1/trailer-classifications/{id} [delete]
func docDeleteTrailerClassification() {}
