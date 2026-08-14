package handler

// docListWorkOrderStatuses godoc
//
//	@Summary	List work-order-statuses
//	@Tags	work-order-statuses
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.WorkOrderStatusPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-statuses [get]
func docListWorkOrderStatuses() {}

// docCreateWorkOrderStatus godoc
//
//	@Summary	Create work-order-statuses
//	@Tags	work-order-statuses
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateWorkOrderStatusRequest	true	"body"
//	@Success	201	{object}	dto.WorkOrderStatusResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-statuses [post]
func docCreateWorkOrderStatus() {}

// docGetWorkOrderStatus godoc
//
//	@Summary	Get work-order-statuses
//	@Tags	work-order-statuses
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.WorkOrderStatusResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-statuses/{id} [get]
func docGetWorkOrderStatus() {}

// docUpdateWorkOrderStatus godoc
//
//	@Summary	Update work-order-statuses
//	@Tags	work-order-statuses
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateWorkOrderStatusRequest	true	"body"
//	@Success	200	{object}	dto.WorkOrderStatusResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-statuses/{id} [put]
func docUpdateWorkOrderStatus() {}

// docDeleteWorkOrderStatus godoc
//
//	@Summary	Delete work-order-statuses
//	@Tags	work-order-statuses
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-statuses/{id} [delete]
func docDeleteWorkOrderStatus() {}
