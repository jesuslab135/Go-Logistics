package handler

// docListPartLocations godoc
//
//	@Summary	List part-locations
//	@Tags	part-locations
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.PartLocationPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-locations [get]
func docListPartLocations() {}

// docCreatePartLocation godoc
//
//	@Summary	Create part-locations
//	@Tags	part-locations
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreatePartLocationRequest	true	"body"
//	@Success	201	{object}	dto.PartLocationResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-locations [post]
func docCreatePartLocation() {}

// docGetPartLocation godoc
//
//	@Summary	Get part-locations
//	@Tags	part-locations
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.PartLocationResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-locations/{id} [get]
func docGetPartLocation() {}

// docUpdatePartLocation godoc
//
//	@Summary	Update part-locations
//	@Tags	part-locations
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdatePartLocationRequest	true	"body"
//	@Success	200	{object}	dto.PartLocationResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-locations/{id} [put]
func docUpdatePartLocation() {}

// docDeletePartLocation godoc
//
//	@Summary	Delete part-locations
//	@Tags	part-locations
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-locations/{id} [delete]
func docDeletePartLocation() {}
