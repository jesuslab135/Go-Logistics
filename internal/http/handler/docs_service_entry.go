package handler

// docListServiceEntries godoc
//
//	@Summary	List service-entries
//	@Tags	service-entries
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.ServiceEntryPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entries [get]
func docListServiceEntries() {}

// docCreateServiceEntry godoc
//
//	@Summary	Create service-entries
//	@Tags	service-entries
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateServiceEntryRequest	true	"body"
//	@Success	201	{object}	dto.ServiceEntryResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entries [post]
func docCreateServiceEntry() {}

// docGetServiceEntry godoc
//
//	@Summary	Get service-entries
//	@Tags	service-entries
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.ServiceEntryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entries/{id} [get]
func docGetServiceEntry() {}

// docUpdateServiceEntry godoc
//
//	@Summary	Update service-entries
//	@Tags	service-entries
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateServiceEntryRequest	true	"body"
//	@Success	200	{object}	dto.ServiceEntryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entries/{id} [put]
func docUpdateServiceEntry() {}

// docDeleteServiceEntry godoc
//
//	@Summary	Delete service-entries
//	@Tags	service-entries
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entries/{id} [delete]
func docDeleteServiceEntry() {}
