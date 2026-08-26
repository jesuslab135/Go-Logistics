package handler

// The transition history is read-only: rows are appended by the transition that
// caused them. The write verbs answer 405 `route_retired` and are not
// documented as operations.

// docNListPurchaseOrderStatusLogs godoc
//
//	@Summary		List purchase-order transitions
//	@Description	Append-only transition history, newest first. Each row records the state moved from and to, who moved it, and the reason where one was required. Written inside the transaction that moved the order, so it cannot disagree with the order.
//	@Tags			purchase-orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id		path	int	true	"Purchase order id"
//	@Param			limit	query	int	false	"Page size"
//	@Param			offset	query	int	false	"Offset"
//	@Success		200		{object}	dto.PurchaseOrderStatusLogPage
//	@Failure		401		{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-orders/{id}/status-logs [get]
func docNListPurchaseOrderStatusLogs() {}

// docNGetPurchaseOrderStatusLog godoc
//
//	@Summary	Get one purchase-order transition
//	@Tags		purchase-orders
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id			path	int	true	"Purchase order id"
//	@Param		child_id	path	int	true	"Transition id"
//	@Success	200			{object}	dto.PurchaseOrderStatusLogResponse
//	@Failure	404			{object}	dto.ErrorResponse
//	@Router		/api/v1/purchase-orders/{id}/status-logs/{child_id} [get]
func docNGetPurchaseOrderStatusLog() {}
