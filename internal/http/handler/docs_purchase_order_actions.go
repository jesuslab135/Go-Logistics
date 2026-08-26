package handler

// The workflow transitions. Each is a POST that moves the order's state, stamps
// the timestamp that transition owns and the actor from the authenticated
// caller, and records the move in the order's status log — all in one
// transaction. A move that is not legal from the current state is 409
// `invalid_transition`, naming that state.
//
// approve and reject additionally require the purchase_orders.approve action:
// deciding to commit money is the privileged act, while submitting an order for
// that decision and receiving the goods afterwards need only update.

// docPurchaseOrderSubmit godoc
//
//	@Summary		Submit a purchase order for approval
//	@Description	DRAFT or REJECTED to PENDING_APPROVAL. Stamps submitted_at and submitted_by_id.
//	@Tags			purchase-orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Purchase order id"
//	@Success		200	{object}	dto.PurchaseOrderResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-orders/{id}/submit [post]
func docPurchaseOrderSubmit() {}

// docPurchaseOrderApprove godoc
//
//	@Summary		Approve a purchase order
//	@Description	PENDING_APPROVAL to APPROVED. Requires the purchase_orders.approve permission. Stamps approved_at and approved_by_id from the caller.
//	@Tags			purchase-orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Purchase order id"
//	@Success		200	{object}	dto.PurchaseOrderResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-orders/{id}/approve [post]
func docPurchaseOrderApprove() {}

// docPurchaseOrderReject godoc
//
//	@Summary		Reject a purchase order
//	@Description	PENDING_APPROVAL to REJECTED. Requires the purchase_orders.approve permission and a reason, which is stored on the order and in its history. Rejection is not terminal — the order is revised back to DRAFT and resubmitted.
//	@Tags			purchase-orders
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int										true	"Purchase order id"
//	@Param			body	body		dto.PurchaseOrderTransitionRequest		true	"Rejection reason"
//	@Success		200		{object}	dto.PurchaseOrderResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-orders/{id}/reject [post]
func docPurchaseOrderReject() {}

// docPurchaseOrderRevise godoc
//
//	@Summary		Revise a rejected purchase order
//	@Description	REJECTED back to DRAFT so it can be edited and resubmitted. Stamps nothing: it undoes a decision rather than making one, and the rejection stays in the history.
//	@Tags			purchase-orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Purchase order id"
//	@Success		200	{object}	dto.PurchaseOrderResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-orders/{id}/revise [post]
func docPurchaseOrderRevise() {}

// docPurchaseOrderPurchase godoc
//
//	@Summary		Mark a purchase order as purchased
//	@Description	APPROVED to PURCHASED. Stamps purchased_at.
//	@Tags			purchase-orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Purchase order id"
//	@Success		200	{object}	dto.PurchaseOrderResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-orders/{id}/purchase [post]
func docPurchaseOrderPurchase() {}

// docPurchaseOrderReceivePartial godoc
//
//	@Summary		Record a partial receipt
//	@Description	PURCHASED to RECEIVED_PARTIAL. Stamps received_partial_at.
//	@Tags			purchase-orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Purchase order id"
//	@Success		200	{object}	dto.PurchaseOrderResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-orders/{id}/receive-partial [post]
func docPurchaseOrderReceivePartial() {}

// docPurchaseOrderReceiveFull godoc
//
//	@Summary		Record full receipt
//	@Description	PURCHASED or RECEIVED_PARTIAL to RECEIVED_FULL. Stamps received_full_at.
//	@Tags			purchase-orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Purchase order id"
//	@Success		200	{object}	dto.PurchaseOrderResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-orders/{id}/receive-full [post]
func docPurchaseOrderReceiveFull() {}

// docPurchaseOrderClose godoc
//
//	@Summary		Close a purchase order
//	@Description	RECEIVED_FULL to CLOSED, which is terminal. Stamps closed_at.
//	@Tags			purchase-orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Purchase order id"
//	@Success		200	{object}	dto.PurchaseOrderResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-orders/{id}/close [post]
func docPurchaseOrderClose() {}
