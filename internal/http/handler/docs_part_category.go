package handler

// docListPartCategories godoc
//
//	@Summary	List part-categories
//	@Tags	part-categories
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.PartCategoryPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-categories [get]
func docListPartCategories() {}

// docCreatePartCategory godoc
//
//	@Summary	Create part-categories
//	@Tags	part-categories
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreatePartCategoryRequest	true	"body"
//	@Success	201	{object}	dto.PartCategoryResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-categories [post]
func docCreatePartCategory() {}

// docGetPartCategory godoc
//
//	@Summary	Get part-categories
//	@Tags	part-categories
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.PartCategoryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-categories/{id} [get]
func docGetPartCategory() {}

// docUpdatePartCategory godoc
//
//	@Summary	Update part-categories
//	@Tags	part-categories
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdatePartCategoryRequest	true	"body"
//	@Success	200	{object}	dto.PartCategoryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-categories/{id} [put]
func docUpdatePartCategory() {}

// docDeletePartCategory godoc
//
//	@Summary	Delete part-categories
//	@Tags	part-categories
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/part-categories/{id} [delete]
func docDeletePartCategory() {}
