package handler

// docListFaults godoc
//
//	@Summary	List faults
//	@Tags	faults
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.FaultPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/faults [get]
func docListFaults() {}

// docCreateFault godoc
//
//	@Summary	Create faults
//	@Tags	faults
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateFaultRequest	true	"body"
//	@Success	201	{object}	dto.FaultResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/faults [post]
func docCreateFault() {}

// docGetFault godoc
//
//	@Summary	Get faults
//	@Tags	faults
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.FaultResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/faults/{id} [get]
func docGetFault() {}

// docUpdateFault godoc
//
//	@Summary	Update faults
//	@Tags	faults
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateFaultRequest	true	"body"
//	@Success	200	{object}	dto.FaultResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/faults/{id} [put]
func docUpdateFault() {}

// docDeleteFault godoc
//
//	@Summary	Delete faults
//	@Tags	faults
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/faults/{id} [delete]
func docDeleteFault() {}
