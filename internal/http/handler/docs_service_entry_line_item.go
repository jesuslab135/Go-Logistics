package handler

// docNListServiceEntryLineItems godoc
//
//	@Summary	List service-entry-line-items
//	@Tags	service-entry-line-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.ServiceEntryLineItemPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entries/{id}/line-items [get]
func docNListServiceEntryLineItems() {}

// docNCreateServiceEntryLineItem godoc
//
//	@Summary	Create service-entry-line-items
//	@Tags	service-entry-line-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateServiceEntryLineItemRequest	true	"body"
//	@Success	201	{object}	dto.ServiceEntryLineItemResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entries/{id}/line-items [post]
func docNCreateServiceEntryLineItem() {}

// docNGetServiceEntryLineItem godoc
//
//	@Summary	Get service-entry-line-items
//	@Tags	service-entry-line-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.ServiceEntryLineItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entries/{id}/line-items/{child_id} [get]
func docNGetServiceEntryLineItem() {}

// docNUpdateServiceEntryLineItem godoc
//
//	@Summary	Update service-entry-line-items
//	@Tags	service-entry-line-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateServiceEntryLineItemRequest	true	"body"
//	@Success	200	{object}	dto.ServiceEntryLineItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entries/{id}/line-items/{child_id} [put]
func docNUpdateServiceEntryLineItem() {}

// docNDeleteServiceEntryLineItem godoc
//
//	@Summary	Delete service-entry-line-items
//	@Tags	service-entry-line-items
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entries/{id}/line-items/{child_id} [delete]
func docNDeleteServiceEntryLineItem() {}
