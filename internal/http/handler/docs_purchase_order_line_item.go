package handler

// docNListPurchaseOrderLineItems godoc
//
//	@Summary	List purchase-order-line-items
//	@Tags	purchase-order-line-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.PurchaseOrderLineItemPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/purchase-orders/{id}/line-items [get]
func docNListPurchaseOrderLineItems() {}

// docNCreatePurchaseOrderLineItem godoc
//
//	@Summary	Create purchase-order-line-items
//	@Tags	purchase-order-line-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreatePurchaseOrderLineItemRequest	true	"body"
//	@Success	201	{object}	dto.PurchaseOrderLineItemResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/purchase-orders/{id}/line-items [post]
func docNCreatePurchaseOrderLineItem() {}

// docNGetPurchaseOrderLineItem godoc
//
//	@Summary	Get purchase-order-line-items
//	@Tags	purchase-order-line-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.PurchaseOrderLineItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/purchase-orders/{id}/line-items/{child_id} [get]
func docNGetPurchaseOrderLineItem() {}

// docNUpdatePurchaseOrderLineItem godoc
//
//	@Summary	Update purchase-order-line-items
//	@Tags	purchase-order-line-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdatePurchaseOrderLineItemRequest	true	"body"
//	@Success	200	{object}	dto.PurchaseOrderLineItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/purchase-orders/{id}/line-items/{child_id} [put]
func docNUpdatePurchaseOrderLineItem() {}

// docNDeletePurchaseOrderLineItem godoc
//
//	@Summary	Delete purchase-order-line-items
//	@Tags	purchase-order-line-items
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/purchase-orders/{id}/line-items/{child_id} [delete]
func docNDeletePurchaseOrderLineItem() {}
