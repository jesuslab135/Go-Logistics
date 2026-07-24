package handler

// docListPartManufacturers godoc
//
//	@Summary	List part-manufacturers
//	@Tags	part-manufacturers
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.PartManufacturerPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-manufacturers [get]
func docListPartManufacturers() {}

// docCreatePartManufacturer godoc
//
//	@Summary	Create part-manufacturers
//	@Tags	part-manufacturers
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreatePartManufacturerRequest	true	"body"
//	@Success	201	{object}	dto.PartManufacturerResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-manufacturers [post]
func docCreatePartManufacturer() {}

// docGetPartManufacturer godoc
//
//	@Summary	Get part-manufacturers
//	@Tags	part-manufacturers
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.PartManufacturerResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-manufacturers/{id} [get]
func docGetPartManufacturer() {}

// docUpdatePartManufacturer godoc
//
//	@Summary	Update part-manufacturers
//	@Tags	part-manufacturers
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdatePartManufacturerRequest	true	"body"
//	@Success	200	{object}	dto.PartManufacturerResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-manufacturers/{id} [put]
func docUpdatePartManufacturer() {}

// docDeletePartManufacturer godoc
//
//	@Summary	Delete part-manufacturers
//	@Tags	part-manufacturers
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-manufacturers/{id} [delete]
func docDeletePartManufacturer() {}
