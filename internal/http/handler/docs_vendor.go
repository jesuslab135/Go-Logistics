package handler

// docListVendors godoc
//
//	@Summary	List vendors
//	@Tags	vendors
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Param	include_archived	query	bool	false	"Include archived records"
//	@Success	200	{object}	dto.VendorPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/vendors [get]
func docListVendors() {}

// docCreateVendor godoc
//
//	@Summary	Create vendors
//	@Tags	vendors
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateVendorRequest	true	"body"
//	@Success	201	{object}	dto.VendorResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/vendors [post]
func docCreateVendor() {}

// docGetVendor godoc
//
//	@Summary	Get vendors
//	@Tags	vendors
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.VendorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/vendors/{id} [get]
func docGetVendor() {}

// docUpdateVendor godoc
//
//	@Summary	Update vendors
//	@Tags	vendors
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateVendorRequest	true	"body"
//	@Success	200	{object}	dto.VendorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/vendors/{id} [put]
func docUpdateVendor() {}

// docDeleteVendor godoc
//
//	@Summary	Delete vendors
//	@Tags	vendors
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/vendors/{id} [delete]
func docDeleteVendor() {}
