package handler

// docNListAssetTrailerAssignments godoc
//
//	@Summary	List asset-trailer-assignments
//	@Tags	asset-trailer-assignments
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.AssetTrailerAssignmentPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/trailer-assignments [get]
func docNListAssetTrailerAssignments() {}

// docNCreateAssetTrailerAssignment godoc
//
//	@Summary	Create asset-trailer-assignments
//	@Tags	asset-trailer-assignments
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateAssetTrailerAssignmentRequest	true	"body"
//	@Success	201	{object}	dto.AssetTrailerAssignmentResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/trailer-assignments [post]
func docNCreateAssetTrailerAssignment() {}

// docNGetAssetTrailerAssignment godoc
//
//	@Summary	Get asset-trailer-assignments
//	@Tags	asset-trailer-assignments
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.AssetTrailerAssignmentResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/trailer-assignments/{child_id} [get]
func docNGetAssetTrailerAssignment() {}

// docNUpdateAssetTrailerAssignment godoc
//
//	@Summary	Update asset-trailer-assignments
//	@Tags	asset-trailer-assignments
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateAssetTrailerAssignmentRequest	true	"body"
//	@Success	200	{object}	dto.AssetTrailerAssignmentResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/trailer-assignments/{child_id} [put]
func docNUpdateAssetTrailerAssignment() {}

// docNDeleteAssetTrailerAssignment godoc
//
//	@Summary	Delete asset-trailer-assignments
//	@Tags	asset-trailer-assignments
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/trailer-assignments/{child_id} [delete]
func docNDeleteAssetTrailerAssignment() {}
