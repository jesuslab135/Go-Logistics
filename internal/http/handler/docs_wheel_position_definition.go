package handler

// docDListWheelPositionDefinitions godoc
//
//	@Summary	List wheel-position-definitions
//	@Tags	wheel-position-definitions
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.WheelPositionDefinitionPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-definitions/{id}/wheel-positions [get]
func docDListWheelPositionDefinitions() {}

// docDCreateWheelPositionDefinition godoc
//
//	@Summary	Create wheel-position-definitions
//	@Tags	wheel-position-definitions
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateWheelPositionDefinitionRequest	true	"body"
//	@Success	201	{object}	dto.WheelPositionDefinitionResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-definitions/{id}/wheel-positions [post]
func docDCreateWheelPositionDefinition() {}

// docDGetWheelPositionDefinition godoc
//
//	@Summary	Get wheel-position-definitions
//	@Tags	wheel-position-definitions
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.WheelPositionDefinitionResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-definitions/{id}/wheel-positions/{child_id} [get]
func docDGetWheelPositionDefinition() {}

// docDUpdateWheelPositionDefinition godoc
//
//	@Summary	Update wheel-position-definitions
//	@Tags	wheel-position-definitions
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateWheelPositionDefinitionRequest	true	"body"
//	@Success	200	{object}	dto.WheelPositionDefinitionResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-definitions/{id}/wheel-positions/{child_id} [put]
func docDUpdateWheelPositionDefinition() {}

// docDDeleteWheelPositionDefinition godoc
//
//	@Summary	Delete wheel-position-definitions
//	@Tags	wheel-position-definitions
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-definitions/{id}/wheel-positions/{child_id} [delete]
func docDDeleteWheelPositionDefinition() {}
