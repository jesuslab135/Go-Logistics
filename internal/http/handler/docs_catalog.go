package handler

// Phase 1 resources are served by generic crud.Handlers; their OpenAPI operations
// are documented here as annotation-only stubs.

// asset-types

// docListAssetTypes godoc
//
//	@Summary	List asset types
//	@Tags		asset-types
//	@Security	BearerAuth
//	@Produce	json
//	@Param		page		query		int	false	"Page number"
//	@Param		page_size	query		int	false	"Items per page"
//	@Success	200			{object}	dto.AssetTypePage
//	@Failure	401			{object}	dto.ErrorResponse
//	@Router		/api/v1/asset-types [get]
func docListAssetTypes() {}

// docCreateAssetType godoc
//
//	@Summary	Create an asset type
//	@Tags		asset-types
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.CreateAssetTypeRequest	true	"Asset type"
//	@Success	201		{object}	dto.AssetTypeResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Router		/api/v1/asset-types [post]
func docCreateAssetType() {}

// docGetAssetType godoc
//
//	@Summary	Get an asset type
//	@Tags		asset-types
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		int	true	"Asset type id"
//	@Success	200	{object}	dto.AssetTypeResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/api/v1/asset-types/{id} [get]
func docGetAssetType() {}

// docUpdateAssetType godoc
//
//	@Summary	Update an asset type
//	@Tags		asset-types
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int							true	"Asset type id"
//	@Param		body	body		dto.UpdateAssetTypeRequest	true	"Asset type"
//	@Success	200		{object}	dto.AssetTypeResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/api/v1/asset-types/{id} [put]
func docUpdateAssetType() {}

// docDeleteAssetType godoc
//
//	@Summary	Delete an asset type
//	@Tags		asset-types
//	@Security	BearerAuth
//	@Param		id	path	int	true	"Asset type id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router		/api/v1/asset-types/{id} [delete]
func docDeleteAssetType() {}

// asset-statuses

// docListAssetStatuses godoc
//
//	@Summary	List asset statuses
//	@Tags		asset-statuses
//	@Security	BearerAuth
//	@Produce	json
//	@Param		page		query		int	false	"Page number"
//	@Param		page_size	query		int	false	"Items per page"
//	@Success	200			{object}	dto.AssetStatusPage
//	@Failure	401			{object}	dto.ErrorResponse
//	@Router		/api/v1/asset-statuses [get]
func docListAssetStatuses() {}

// docCreateAssetStatus godoc
//
//	@Summary	Create an asset status
//	@Tags		asset-statuses
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.CreateAssetStatusRequest	true	"Asset status"
//	@Success	201		{object}	dto.AssetStatusResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Router		/api/v1/asset-statuses [post]
func docCreateAssetStatus() {}

// docGetAssetStatus godoc
//
//	@Summary	Get an asset status
//	@Tags		asset-statuses
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		int	true	"Asset status id"
//	@Success	200	{object}	dto.AssetStatusResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/api/v1/asset-statuses/{id} [get]
func docGetAssetStatus() {}

// docUpdateAssetStatus godoc
//
//	@Summary	Update an asset status
//	@Tags		asset-statuses
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int								true	"Asset status id"
//	@Param		body	body		dto.UpdateAssetStatusRequest	true	"Asset status"
//	@Success	200		{object}	dto.AssetStatusResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/api/v1/asset-statuses/{id} [put]
func docUpdateAssetStatus() {}

// docDeleteAssetStatus godoc
//
//	@Summary	Delete an asset status
//	@Tags		asset-statuses
//	@Security	BearerAuth
//	@Param		id	path	int	true	"Asset status id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router		/api/v1/asset-statuses/{id} [delete]
func docDeleteAssetStatus() {}

// docCreateCatalogOption godoc
//
//	@Summary	Create a catalog option
//	@Tags		catalog-options
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.CreateCatalogOptionRequest	true	"Catalog option"
//	@Success	201		{object}	dto.CatalogOptionResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Router		/api/v1/catalog-options [post]
func docCreateCatalogOption() {}

// docGetCatalogOption godoc
//
//	@Summary	Get a catalog option
//	@Tags		catalog-options
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		int	true	"Catalog option id"
//	@Success	200	{object}	dto.CatalogOptionResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/api/v1/catalog-options/{id} [get]
func docGetCatalogOption() {}

// docUpdateCatalogOption godoc
//
//	@Summary	Update a catalog option
//	@Tags		catalog-options
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int								true	"Catalog option id"
//	@Param		body	body		dto.UpdateCatalogOptionRequest	true	"Catalog option"
//	@Success	200		{object}	dto.CatalogOptionResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/api/v1/catalog-options/{id} [put]
func docUpdateCatalogOption() {}

// docDeleteCatalogOption godoc
//
//	@Summary	Delete a catalog option
//	@Tags		catalog-options
//	@Security	BearerAuth
//	@Param		id	path	int	true	"Catalog option id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router		/api/v1/catalog-options/{id} [delete]
func docDeleteCatalogOption() {}

// vehicle-makes

// docListVehicleMakes godoc
//
//	@Summary	List vehicle makes
//	@Tags		vehicle-makes
//	@Security	BearerAuth
//	@Produce	json
//	@Param		page		query		int	false	"Page number"
//	@Param		page_size	query		int	false	"Items per page"
//	@Success	200			{object}	dto.VehicleMakePage
//	@Failure	401			{object}	dto.ErrorResponse
//	@Router		/api/v1/vehicle-makes [get]
func docListVehicleMakes() {}

// docCreateVehicleMake godoc
//
//	@Summary	Create a vehicle make
//	@Tags		vehicle-makes
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.CreateVehicleMakeRequest	true	"Vehicle make"
//	@Success	201		{object}	dto.VehicleMakeResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Router		/api/v1/vehicle-makes [post]
func docCreateVehicleMake() {}

// docGetVehicleMake godoc
//
//	@Summary	Get a vehicle make
//	@Tags		vehicle-makes
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		int	true	"Vehicle make id"
//	@Success	200	{object}	dto.VehicleMakeResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/api/v1/vehicle-makes/{id} [get]
func docGetVehicleMake() {}

// docUpdateVehicleMake godoc
//
//	@Summary	Update a vehicle make
//	@Tags		vehicle-makes
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int								true	"Vehicle make id"
//	@Param		body	body		dto.UpdateVehicleMakeRequest	true	"Vehicle make"
//	@Success	200		{object}	dto.VehicleMakeResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/api/v1/vehicle-makes/{id} [put]
func docUpdateVehicleMake() {}

// docDeleteVehicleMake godoc
//
//	@Summary	Delete a vehicle make
//	@Tags		vehicle-makes
//	@Security	BearerAuth
//	@Param		id	path	int	true	"Vehicle make id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router		/api/v1/vehicle-makes/{id} [delete]
func docDeleteVehicleMake() {}

// vehicle-models

// docListVehicleModels godoc
//
//	@Summary	List vehicle models
//	@Tags		vehicle-models
//	@Security	BearerAuth
//	@Produce	json
//	@Param		page		query		int	false	"Page number"
//	@Param		page_size	query		int	false	"Items per page"
//	@Success	200			{object}	dto.VehicleModelPage
//	@Failure	401			{object}	dto.ErrorResponse
//	@Router		/api/v1/vehicle-models [get]
func docListVehicleModels() {}

// docCreateVehicleModel godoc
//
//	@Summary	Create a vehicle model
//	@Tags		vehicle-models
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.CreateVehicleModelRequest	true	"Vehicle model"
//	@Success	201		{object}	dto.VehicleModelResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Router		/api/v1/vehicle-models [post]
func docCreateVehicleModel() {}

// docGetVehicleModel godoc
//
//	@Summary	Get a vehicle model
//	@Tags		vehicle-models
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		int	true	"Vehicle model id"
//	@Success	200	{object}	dto.VehicleModelResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/api/v1/vehicle-models/{id} [get]
func docGetVehicleModel() {}

// docUpdateVehicleModel godoc
//
//	@Summary	Update a vehicle model
//	@Tags		vehicle-models
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int								true	"Vehicle model id"
//	@Param		body	body		dto.UpdateVehicleModelRequest	true	"Vehicle model"
//	@Success	200		{object}	dto.VehicleModelResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/api/v1/vehicle-models/{id} [put]
func docUpdateVehicleModel() {}

// docDeleteVehicleModel godoc
//
//	@Summary	Delete a vehicle model
//	@Tags		vehicle-models
//	@Security	BearerAuth
//	@Param		id	path	int	true	"Vehicle model id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router		/api/v1/vehicle-models/{id} [delete]
func docDeleteVehicleModel() {}
