package handler

// docCreateMedium godoc
//
//	@Summary	Create media
//	@Tags	media
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateMediumRequest	true	"body"
//	@Success	201	{object}	dto.MediumResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/media [post]
func docCreateMedium() {}

// docGetMedium godoc
//
//	@Summary	Get media
//	@Tags	media
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.MediumResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/media/{id} [get]
func docGetMedium() {}

// docUpdateMedium godoc
//
//	@Summary	Update media
//	@Tags	media
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateMediumRequest	true	"body"
//	@Success	200	{object}	dto.MediumResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/media/{id} [put]
func docUpdateMedium() {}

// docDeleteMedium godoc
//
//	@Summary	Delete media
//	@Tags	media
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/media/{id} [delete]
func docDeleteMedium() {}
