package handler

// docListAxleTemplates godoc
//
//	@Summary	List axle-templates
//	@Tags	axle-templates
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.AxleTemplatePage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-templates [get]
func docListAxleTemplates() {}

// docCreateAxleTemplate godoc
//
//	@Summary	Create axle-templates
//	@Tags	axle-templates
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateAxleTemplateRequest	true	"body"
//	@Success	201	{object}	dto.AxleTemplateResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-templates [post]
func docCreateAxleTemplate() {}

// docGetAxleTemplate godoc
//
//	@Summary	Get axle-templates
//	@Tags	axle-templates
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.AxleTemplateResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-templates/{id} [get]
func docGetAxleTemplate() {}

// docUpdateAxleTemplate godoc
//
//	@Summary	Update axle-templates
//	@Tags	axle-templates
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateAxleTemplateRequest	true	"body"
//	@Success	200	{object}	dto.AxleTemplateResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-templates/{id} [put]
func docUpdateAxleTemplate() {}

// docDeleteAxleTemplate godoc
//
//	@Summary	Delete axle-templates
//	@Tags	axle-templates
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/axle-templates/{id} [delete]
func docDeleteAxleTemplate() {}
