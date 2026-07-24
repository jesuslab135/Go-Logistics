package handler

// docNListServiceTaskParts godoc
//
//	@Summary	List service-task-parts
//	@Tags	service-task-parts
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.ServiceTaskPartPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-tasks/{id}/parts [get]
func docNListServiceTaskParts() {}

// docNCreateServiceTaskPart godoc
//
//	@Summary	Create service-task-parts
//	@Tags	service-task-parts
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateServiceTaskPartRequest	true	"body"
//	@Success	201	{object}	dto.ServiceTaskPartResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-tasks/{id}/parts [post]
func docNCreateServiceTaskPart() {}

// docNGetServiceTaskPart godoc
//
//	@Summary	Get service-task-parts
//	@Tags	service-task-parts
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.ServiceTaskPartResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-tasks/{id}/parts/{child_id} [get]
func docNGetServiceTaskPart() {}

// docNUpdateServiceTaskPart godoc
//
//	@Summary	Update service-task-parts
//	@Tags	service-task-parts
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateServiceTaskPartRequest	true	"body"
//	@Success	200	{object}	dto.ServiceTaskPartResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-tasks/{id}/parts/{child_id} [put]
func docNUpdateServiceTaskPart() {}

// docNDeleteServiceTaskPart godoc
//
//	@Summary	Delete service-task-parts
//	@Tags	service-task-parts
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-tasks/{id}/parts/{child_id} [delete]
func docNDeleteServiceTaskPart() {}
