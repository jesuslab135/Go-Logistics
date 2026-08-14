package handler

// docCreateWorkOrder godoc
//
//	@Summary	Create work-orders
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateWorkOrderRequest	true	"body"
//	@Success	201	{object}	dto.WorkOrderResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders [post]
func docCreateWorkOrder() {}

// docGetWorkOrder godoc
//
//	@Summary	Get work-orders
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.WorkOrderResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id} [get]
func docGetWorkOrder() {}

// docUpdateWorkOrder godoc
//
//	@Summary	Update work-orders
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateWorkOrderRequest	true	"body"
//	@Success	200	{object}	dto.WorkOrderResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id} [put]
func docUpdateWorkOrder() {}

// docDeleteWorkOrder godoc
//
//	@Summary	Delete work-orders
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id} [delete]
func docDeleteWorkOrder() {}
