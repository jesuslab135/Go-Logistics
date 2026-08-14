package handler

// docListMeasurementUnits godoc
//
//	@Summary	List measurement-units
//	@Tags	measurement-units
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.MeasurementUnitPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/measurement-units [get]
func docListMeasurementUnits() {}

// docCreateMeasurementUnit godoc
//
//	@Summary	Create measurement-units
//	@Tags	measurement-units
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateMeasurementUnitRequest	true	"body"
//	@Success	201	{object}	dto.MeasurementUnitResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/measurement-units [post]
func docCreateMeasurementUnit() {}

// docGetMeasurementUnit godoc
//
//	@Summary	Get measurement-units
//	@Tags	measurement-units
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.MeasurementUnitResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/measurement-units/{id} [get]
func docGetMeasurementUnit() {}

// docUpdateMeasurementUnit godoc
//
//	@Summary	Update measurement-units
//	@Tags	measurement-units
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateMeasurementUnitRequest	true	"body"
//	@Success	200	{object}	dto.MeasurementUnitResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/measurement-units/{id} [put]
func docUpdateMeasurementUnit() {}

// docDeleteMeasurementUnit godoc
//
//	@Summary	Delete measurement-units
//	@Tags	measurement-units
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/measurement-units/{id} [delete]
func docDeleteMeasurementUnit() {}
