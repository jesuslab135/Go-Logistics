package handler

// docSGetVehicle godoc
//
//	@Summary	Get vehicle
//	@Tags	vehicle
//	@Security	BearerAuth
//	@Param	id	path	int	true	"asset id"
//	@Produce	json
//	@Success	200	{object}	dto.VehicleResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/vehicle [get]
func docSGetVehicle() {}

// docSUpsertVehicle godoc
//
//	@Summary	Create or update vehicle
//	@Tags	vehicle
//	@Security	BearerAuth
//	@Param	id	path	int	true	"asset id"
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.UpsertVehicleRequest	true	"body"
//	@Success	200	{object}	dto.VehicleResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/vehicle [put]
func docSUpsertVehicle() {}

// docSDeleteVehicle godoc
//
//	@Summary	Delete vehicle
//	@Tags	vehicle
//	@Security	BearerAuth
//	@Param	id	path	int	true	"asset id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/vehicle [delete]
func docSDeleteVehicle() {}
