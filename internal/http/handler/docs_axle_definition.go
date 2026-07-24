package handler

// docNListAxleDefinitions godoc
//
//	@Summary	List axle-definitions
//	@Tags	axle-definitions
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.AxleDefinitionPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-templates/{id}/definitions [get]
func docNListAxleDefinitions() {}

// docNCreateAxleDefinition godoc
//
//	@Summary	Create axle-definitions
//	@Tags	axle-definitions
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateAxleDefinitionRequest	true	"body"
//	@Success	201	{object}	dto.AxleDefinitionResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-templates/{id}/definitions [post]
func docNCreateAxleDefinition() {}

// docNGetAxleDefinition godoc
//
//	@Summary	Get axle-definitions
//	@Tags	axle-definitions
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.AxleDefinitionResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-templates/{id}/definitions/{child_id} [get]
func docNGetAxleDefinition() {}

// docNUpdateAxleDefinition godoc
//
//	@Summary	Update axle-definitions
//	@Tags	axle-definitions
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateAxleDefinitionRequest	true	"body"
//	@Success	200	{object}	dto.AxleDefinitionResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-templates/{id}/definitions/{child_id} [put]
func docNUpdateAxleDefinition() {}

// docNDeleteAxleDefinition godoc
//
//	@Summary	Delete axle-definitions
//	@Tags	axle-definitions
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-templates/{id}/definitions/{child_id} [delete]
func docNDeleteAxleDefinition() {}
