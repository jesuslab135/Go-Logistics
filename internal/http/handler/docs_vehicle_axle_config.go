package handler

// docSGetVehicleAxleConfig godoc
//
//	@Summary	Get vehicle-axle-config
//	@Tags	vehicle-axle-config
//	@Security	BearerAuth
//	@Param	id	path	int	true	"asset id"
//	@Produce	json
//	@Success	200	{object}	dto.VehicleAxleConfigResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/axle-config [get]
func docSGetVehicleAxleConfig() {}

// docSUpsertVehicleAxleConfig godoc
//
//	@Summary	Create or update vehicle-axle-config
//	@Tags	vehicle-axle-config
//	@Security	BearerAuth
//	@Param	id	path	int	true	"asset id"
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.UpsertVehicleAxleConfigRequest	true	"body"
//	@Success	200	{object}	dto.VehicleAxleConfigResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/axle-config [put]
func docSUpsertVehicleAxleConfig() {}

// docSDeleteVehicleAxleConfig godoc
//
//	@Summary	Delete vehicle-axle-config
//	@Tags	vehicle-axle-config
//	@Security	BearerAuth
//	@Param	id	path	int	true	"asset id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/axle-config [delete]
func docSDeleteVehicleAxleConfig() {}
