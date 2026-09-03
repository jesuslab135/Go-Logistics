package handler

// docCreateAsset godoc
//
//	@Summary	Create assets
//	@Description	Supplying a nested "vehicle" or "trailer" object makes the write atomic: the asset row and its subtype are created in one transaction, so a validation or write failure on the subtype rolls back the asset too. Omit both for an asset-only create.
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
//	@Description	Supplying a nested "vehicle" or "trailer" object makes the write atomic: the asset row and its subtype are updated in one transaction, so a validation or write failure on the subtype rolls back the asset update too. Omit both for an asset-only edit (status, photo, etc.), which keeps the single-write path unchanged.
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
