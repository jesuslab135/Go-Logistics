package handler

// docNListWorkOrderStatusLogs godoc
//
//	@Summary	List work-order-status-logs
//	@Tags	work-order-status-logs
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.WorkOrderStatusLogPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/status-logs [get]
func docNListWorkOrderStatusLogs() {}

// docNCreateWorkOrderStatusLog godoc
//
//	@Summary	Create work-order-status-logs
//	@Tags	work-order-status-logs
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateWorkOrderStatusLogRequest	true	"body"
//	@Success	201	{object}	dto.WorkOrderStatusLogResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/status-logs [post]
func docNCreateWorkOrderStatusLog() {}

// docNGetWorkOrderStatusLog godoc
//
//	@Summary	Get work-order-status-logs
//	@Tags	work-order-status-logs
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.WorkOrderStatusLogResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/status-logs/{child_id} [get]
func docNGetWorkOrderStatusLog() {}

// docNUpdateWorkOrderStatusLog godoc
//
//	@Summary	Update work-order-status-logs
//	@Tags	work-order-status-logs
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateWorkOrderStatusLogRequest	true	"body"
//	@Success	200	{object}	dto.WorkOrderStatusLogResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/status-logs/{child_id} [put]
func docNUpdateWorkOrderStatusLog() {}

// docNDeleteWorkOrderStatusLog godoc
//
//	@Summary	Delete work-order-status-logs
//	@Tags	work-order-status-logs
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/status-logs/{child_id} [delete]
func docNDeleteWorkOrderStatusLog() {}
