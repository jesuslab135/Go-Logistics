package handler

// docCreatePurchaseOrder godoc
//
//	@Summary	Create purchase-orders
//	@Tags	purchase-orders
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreatePurchaseOrderRequest	true	"body"
//	@Success	201	{object}	dto.PurchaseOrderResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/purchase-orders [post]
func docCreatePurchaseOrder() {}

// docGetPurchaseOrder godoc
//
//	@Summary	Get purchase-orders
//	@Tags	purchase-orders
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.PurchaseOrderResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/purchase-orders/{id} [get]
func docGetPurchaseOrder() {}

// docUpdatePurchaseOrder godoc
//
//	@Summary	Update purchase-orders
//	@Tags	purchase-orders
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdatePurchaseOrderRequest	true	"body"
//	@Success	200	{object}	dto.PurchaseOrderResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/purchase-orders/{id} [put]
func docUpdatePurchaseOrder() {}

// docDeletePurchaseOrder godoc
//
//	@Summary	Delete purchase-orders
//	@Tags	purchase-orders
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/purchase-orders/{id} [delete]
func docDeletePurchaseOrder() {}
