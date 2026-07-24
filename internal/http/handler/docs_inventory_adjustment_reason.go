package handler

// docListInventoryAdjustmentReasons godoc
//
//	@Summary	List inventory-adjustment-reasons
//	@Tags	inventory-adjustment-reasons
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.InventoryAdjustmentReasonPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inventory-adjustment-reasons [get]
func docListInventoryAdjustmentReasons() {}

// docCreateInventoryAdjustmentReason godoc
//
//	@Summary	Create inventory-adjustment-reasons
//	@Tags	inventory-adjustment-reasons
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateInventoryAdjustmentReasonRequest	true	"body"
//	@Success	201	{object}	dto.InventoryAdjustmentReasonResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/inventory-adjustment-reasons [post]
func docCreateInventoryAdjustmentReason() {}

// docGetInventoryAdjustmentReason godoc
//
//	@Summary	Get inventory-adjustment-reasons
//	@Tags	inventory-adjustment-reasons
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.InventoryAdjustmentReasonResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inventory-adjustment-reasons/{id} [get]
func docGetInventoryAdjustmentReason() {}

// docUpdateInventoryAdjustmentReason godoc
//
//	@Summary	Update inventory-adjustment-reasons
//	@Tags	inventory-adjustment-reasons
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateInventoryAdjustmentReasonRequest	true	"body"
//	@Success	200	{object}	dto.InventoryAdjustmentReasonResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inventory-adjustment-reasons/{id} [put]
func docUpdateInventoryAdjustmentReason() {}

// docDeleteInventoryAdjustmentReason godoc
//
//	@Summary	Delete inventory-adjustment-reasons
//	@Tags	inventory-adjustment-reasons
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inventory-adjustment-reasons/{id} [delete]
func docDeleteInventoryAdjustmentReason() {}
