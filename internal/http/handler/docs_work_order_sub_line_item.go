package handler

// docDListWorkOrderSubLineItems godoc
//
//	@Summary	List work-order-sub-line-items
//	@Tags	work-order-sub-line-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.WorkOrderSubLineItemPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-line-items/{id}/sub-line-items [get]
func docDListWorkOrderSubLineItems() {}

// docDCreateWorkOrderSubLineItem godoc
//
//	@Summary	Create work-order-sub-line-items
//	@Tags	work-order-sub-line-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateWorkOrderSubLineItemRequest	true	"body"
//	@Success	201	{object}	dto.WorkOrderSubLineItemResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-line-items/{id}/sub-line-items [post]
func docDCreateWorkOrderSubLineItem() {}

// docDGetWorkOrderSubLineItem godoc
//
//	@Summary	Get work-order-sub-line-items
//	@Tags	work-order-sub-line-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.WorkOrderSubLineItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-line-items/{id}/sub-line-items/{child_id} [get]
func docDGetWorkOrderSubLineItem() {}

// docDUpdateWorkOrderSubLineItem godoc
//
//	@Summary	Update work-order-sub-line-items
//	@Tags	work-order-sub-line-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateWorkOrderSubLineItemRequest	true	"body"
//	@Success	200	{object}	dto.WorkOrderSubLineItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-line-items/{id}/sub-line-items/{child_id} [put]
func docDUpdateWorkOrderSubLineItem() {}

// docDDeleteWorkOrderSubLineItem godoc
//
//	@Summary	Delete work-order-sub-line-items
//	@Tags	work-order-sub-line-items
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-line-items/{id}/sub-line-items/{child_id} [delete]
func docDDeleteWorkOrderSubLineItem() {}
