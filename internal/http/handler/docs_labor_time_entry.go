package handler

// docDListLaborTimeEntries godoc
//
//	@Summary	List labor-time-entries
//	@Tags	labor-time-entries
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.LaborTimeEntryPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-sub-line-items/{id}/labor-entries [get]
func docDListLaborTimeEntries() {}

// docDCreateLaborTimeEntry godoc
//
//	@Summary	Create labor-time-entries
//	@Tags	labor-time-entries
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateLaborTimeEntryRequest	true	"body"
//	@Success	201	{object}	dto.LaborTimeEntryResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-sub-line-items/{id}/labor-entries [post]
func docDCreateLaborTimeEntry() {}

// docDGetLaborTimeEntry godoc
//
//	@Summary	Get labor-time-entries
//	@Tags	labor-time-entries
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.LaborTimeEntryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id} [get]
func docDGetLaborTimeEntry() {}

// docDUpdateLaborTimeEntry godoc
//
//	@Summary	Update labor-time-entries
//	@Tags	labor-time-entries
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateLaborTimeEntryRequest	true	"body"
//	@Success	200	{object}	dto.LaborTimeEntryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id} [put]
func docDUpdateLaborTimeEntry() {}

// docDDeleteLaborTimeEntry godoc
//
//	@Summary	Delete labor-time-entries
//	@Tags	labor-time-entries
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id} [delete]
func docDDeleteLaborTimeEntry() {}
