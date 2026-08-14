package handler

// docListFuelTypes godoc
//
//	@Summary	List fuel-types
//	@Tags	fuel-types
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.FuelTypePage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-types [get]
func docListFuelTypes() {}

// docCreateFuelType godoc
//
//	@Summary	Create fuel-types
//	@Tags	fuel-types
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateFuelTypeRequest	true	"body"
//	@Success	201	{object}	dto.FuelTypeResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-types [post]
func docCreateFuelType() {}

// docGetFuelType godoc
//
//	@Summary	Get fuel-types
//	@Tags	fuel-types
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.FuelTypeResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-types/{id} [get]
func docGetFuelType() {}

// docUpdateFuelType godoc
//
//	@Summary	Update fuel-types
//	@Tags	fuel-types
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateFuelTypeRequest	true	"body"
//	@Success	200	{object}	dto.FuelTypeResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-types/{id} [put]
func docUpdateFuelType() {}

// docDeleteFuelType godoc
//
//	@Summary	Delete fuel-types
//	@Tags	fuel-types
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-types/{id} [delete]
func docDeleteFuelType() {}
