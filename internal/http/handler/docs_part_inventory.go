package handler

// docNListPartInventories godoc
//
//	@Summary	List part-inventory
//	@Tags	part-inventory
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.PartInventoryPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/parts/{id}/inventory [get]
func docNListPartInventories() {}

// docNCreatePartInventory godoc
//
//	@Summary	Create part-inventory
//	@Tags	part-inventory
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreatePartInventoryRequest	true	"body"
//	@Success	201	{object}	dto.PartInventoryResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/parts/{id}/inventory [post]
func docNCreatePartInventory() {}

// docNGetPartInventory godoc
//
//	@Summary	Get part-inventory
//	@Tags	part-inventory
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.PartInventoryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/parts/{id}/inventory/{child_id} [get]
func docNGetPartInventory() {}

// docNUpdatePartInventory godoc
//
//	@Summary	Update part-inventory
//	@Tags	part-inventory
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdatePartInventoryRequest	true	"body"
//	@Success	200	{object}	dto.PartInventoryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/parts/{id}/inventory/{child_id} [put]
func docNUpdatePartInventory() {}

// docNDeletePartInventory godoc
//
//	@Summary	Delete part-inventory
//	@Tags	part-inventory
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/parts/{id}/inventory/{child_id} [delete]
func docNDeletePartInventory() {}
