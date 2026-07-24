package handler

// docListAssets godoc
//
//	@Summary	List assets
//	@Tags	assets
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.AssetPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets [get]
func docListAssets() {}

// docCreateAsset godoc
//
//	@Summary	Create assets
//	@Tags	assets
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateAssetRequest	true	"body"
//	@Success	201	{object}	dto.AssetResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets [post]
func docCreateAsset() {}

// docGetAsset godoc
//
//	@Summary	Get assets
//	@Tags	assets
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.AssetResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id} [get]
func docGetAsset() {}

// docUpdateAsset godoc
//
//	@Summary	Update assets
//	@Tags	assets
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateAssetRequest	true	"body"
//	@Success	200	{object}	dto.AssetResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id} [put]
func docUpdateAsset() {}

// docDeleteAsset godoc
//
//	@Summary	Delete assets
//	@Tags	assets
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id} [delete]
func docDeleteAsset() {}
