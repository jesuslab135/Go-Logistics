package handler

// docSGetTrailer godoc
//
//	@Summary	Get trailer
//	@Tags	trailer
//	@Security	BearerAuth
//	@Param	id	path	int	true	"asset id"
//	@Produce	json
//	@Success	200	{object}	dto.TrailerResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/trailer [get]
func docSGetTrailer() {}

// docSUpsertTrailer godoc
//
//	@Summary	Create or update trailer
//	@Tags	trailer
//	@Security	BearerAuth
//	@Param	id	path	int	true	"asset id"
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.UpsertTrailerRequest	true	"body"
//	@Success	200	{object}	dto.TrailerResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/trailer [put]
func docSUpsertTrailer() {}

// docSDeleteTrailer godoc
//
//	@Summary	Delete trailer
//	@Tags	trailer
//	@Security	BearerAuth
//	@Param	id	path	int	true	"asset id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/trailer [delete]
func docSDeleteTrailer() {}
