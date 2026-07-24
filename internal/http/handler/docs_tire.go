package handler

// docListTires godoc
//
//	@Summary	List tires
//	@Tags	tires
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.TirePage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires [get]
func docListTires() {}

// docCreateTire godoc
//
//	@Summary	Create tires
//	@Tags	tires
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateTireRequest	true	"body"
//	@Success	201	{object}	dto.TireResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires [post]
func docCreateTire() {}

// docGetTire godoc
//
//	@Summary	Get tires
//	@Tags	tires
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.TireResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id} [get]
func docGetTire() {}

// docUpdateTire godoc
//
//	@Summary	Update tires
//	@Tags	tires
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateTireRequest	true	"body"
//	@Success	200	{object}	dto.TireResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id} [put]
func docUpdateTire() {}

// docDeleteTire godoc
//
//	@Summary	Delete tires
//	@Tags	tires
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id} [delete]
func docDeleteTire() {}
