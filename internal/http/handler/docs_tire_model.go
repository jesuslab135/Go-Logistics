package handler

// docListTireModels godoc
//
//	@Summary	List tire-models
//	@Tags	tire-models
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.TireModelPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tire-models [get]
func docListTireModels() {}

// docCreateTireModel godoc
//
//	@Summary	Create tire-models
//	@Tags	tire-models
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateTireModelRequest	true	"body"
//	@Success	201	{object}	dto.TireModelResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/tire-models [post]
func docCreateTireModel() {}

// docGetTireModel godoc
//
//	@Summary	Get tire-models
//	@Tags	tire-models
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.TireModelResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tire-models/{id} [get]
func docGetTireModel() {}

// docUpdateTireModel godoc
//
//	@Summary	Update tire-models
//	@Tags	tire-models
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateTireModelRequest	true	"body"
//	@Success	200	{object}	dto.TireModelResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tire-models/{id} [put]
func docUpdateTireModel() {}

// docDeleteTireModel godoc
//
//	@Summary	Delete tire-models
//	@Tags	tire-models
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tire-models/{id} [delete]
func docDeleteTireModel() {}
