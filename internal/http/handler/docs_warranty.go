package handler

// docListWarranties godoc
//
//	@Summary	List warranties
//	@Tags	warranties
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.WarrantyPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/warranties [get]
func docListWarranties() {}

// docCreateWarranty godoc
//
//	@Summary	Create warranties
//	@Tags	warranties
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateWarrantyRequest	true	"body"
//	@Success	201	{object}	dto.WarrantyResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/warranties [post]
func docCreateWarranty() {}

// docGetWarranty godoc
//
//	@Summary	Get warranties
//	@Tags	warranties
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.WarrantyResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/warranties/{id} [get]
func docGetWarranty() {}

// docUpdateWarranty godoc
//
//	@Summary	Update warranties
//	@Tags	warranties
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateWarrantyRequest	true	"body"
//	@Success	200	{object}	dto.WarrantyResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/warranties/{id} [put]
func docUpdateWarranty() {}

// docDeleteWarranty godoc
//
//	@Summary	Delete warranties
//	@Tags	warranties
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/warranties/{id} [delete]
func docDeleteWarranty() {}
