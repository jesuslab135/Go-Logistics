package handler

// docListParts godoc
//
//	@Summary	List parts
//	@Tags	parts
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Param	include_archived	query	bool	false	"Include archived records"
//	@Success	200	{object}	dto.PartPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/parts [get]
func docListParts() {}

// docCreatePart godoc
//
//	@Summary	Create parts
//	@Tags	parts
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreatePartRequest	true	"body"
//	@Success	201	{object}	dto.PartResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/parts [post]
func docCreatePart() {}

// docGetPart godoc
//
//	@Summary	Get parts
//	@Tags	parts
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.PartResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/parts/{id} [get]
func docGetPart() {}

// docUpdatePart godoc
//
//	@Summary	Update parts
//	@Tags	parts
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdatePartRequest	true	"body"
//	@Success	200	{object}	dto.PartResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/parts/{id} [put]
func docUpdatePart() {}

// docDeletePart godoc
//
//	@Summary	Delete parts
//	@Tags	parts
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/parts/{id} [delete]
func docDeletePart() {}
