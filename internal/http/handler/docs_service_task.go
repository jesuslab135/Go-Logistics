package handler

// docListServiceTasks godoc
//
//	@Summary	List service-tasks
//	@Tags	service-tasks
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.ServiceTaskPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-tasks [get]
func docListServiceTasks() {}

// docCreateServiceTask godoc
//
//	@Summary	Create service-tasks
//	@Tags	service-tasks
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateServiceTaskRequest	true	"body"
//	@Success	201	{object}	dto.ServiceTaskResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-tasks [post]
func docCreateServiceTask() {}

// docGetServiceTask godoc
//
//	@Summary	Get service-tasks
//	@Tags	service-tasks
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.ServiceTaskResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-tasks/{id} [get]
func docGetServiceTask() {}

// docUpdateServiceTask godoc
//
//	@Summary	Update service-tasks
//	@Tags	service-tasks
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateServiceTaskRequest	true	"body"
//	@Success	200	{object}	dto.ServiceTaskResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-tasks/{id} [put]
func docUpdateServiceTask() {}

// docDeleteServiceTask godoc
//
//	@Summary	Delete service-tasks
//	@Tags	service-tasks
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-tasks/{id} [delete]
func docDeleteServiceTask() {}
