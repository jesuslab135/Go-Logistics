package handler

// docDListFuelComments godoc
//
//	@Summary	List fuel-comments
//	@Tags	fuel-comments
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.FuelCommentPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-entries/{id}/comments [get]
func docDListFuelComments() {}

// docDCreateFuelComment godoc
//
//	@Summary	Create fuel-comments
//	@Tags	fuel-comments
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateFuelCommentRequest	true	"body"
//	@Success	201	{object}	dto.FuelCommentResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-entries/{id}/comments [post]
func docDCreateFuelComment() {}

// docDGetFuelComment godoc
//
//	@Summary	Get fuel-comments
//	@Tags	fuel-comments
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.FuelCommentResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-entries/{id}/comments/{child_id} [get]
func docDGetFuelComment() {}

// docDUpdateFuelComment godoc
//
//	@Summary	Update fuel-comments
//	@Tags	fuel-comments
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateFuelCommentRequest	true	"body"
//	@Success	200	{object}	dto.FuelCommentResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-entries/{id}/comments/{child_id} [put]
func docDUpdateFuelComment() {}

// docDDeleteFuelComment godoc
//
//	@Summary	Delete fuel-comments
//	@Tags	fuel-comments
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-entries/{id}/comments/{child_id} [delete]
func docDDeleteFuelComment() {}
