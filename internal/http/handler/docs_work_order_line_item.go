package handler

// docNListWorkOrderLineItems godoc
//
//	@Summary	List work-order-line-items
//	@Tags	work-order-line-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.WorkOrderLineItemPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/line-items [get]
func docNListWorkOrderLineItems() {}

// docNCreateWorkOrderLineItem godoc
//
//	@Summary	Create work-order-line-items
//	@Tags	work-order-line-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateWorkOrderLineItemRequest	true	"body"
//	@Success	201	{object}	dto.WorkOrderLineItemResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/line-items [post]
func docNCreateWorkOrderLineItem() {}

// docNGetWorkOrderLineItem godoc
//
//	@Summary	Get work-order-line-items
//	@Tags	work-order-line-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.WorkOrderLineItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/line-items/{child_id} [get]
func docNGetWorkOrderLineItem() {}

// docNUpdateWorkOrderLineItem godoc
//
//	@Summary	Update work-order-line-items
//	@Tags	work-order-line-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateWorkOrderLineItemRequest	true	"body"
//	@Success	200	{object}	dto.WorkOrderLineItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/line-items/{child_id} [put]
func docNUpdateWorkOrderLineItem() {}

// docNDeleteWorkOrderLineItem godoc
//
//	@Summary	Delete work-order-line-items
//	@Tags	work-order-line-items
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/line-items/{child_id} [delete]
func docNDeleteWorkOrderLineItem() {}
