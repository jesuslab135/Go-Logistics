package handler

// docListInspectionForms godoc
//
//	@Summary	List inspection-forms
//	@Tags	inspection-forms
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.InspectionFormPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-forms [get]
func docListInspectionForms() {}

// docCreateInspectionForm godoc
//
//	@Summary	Create inspection-forms
//	@Tags	inspection-forms
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateInspectionFormRequest	true	"body"
//	@Success	201	{object}	dto.InspectionFormResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-forms [post]
func docCreateInspectionForm() {}

// docGetInspectionForm godoc
//
//	@Summary	Get inspection-forms
//	@Tags	inspection-forms
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.InspectionFormResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-forms/{id} [get]
func docGetInspectionForm() {}

// docUpdateInspectionForm godoc
//
//	@Summary	Update inspection-forms
//	@Tags	inspection-forms
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateInspectionFormRequest	true	"body"
//	@Success	200	{object}	dto.InspectionFormResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-forms/{id} [put]
func docUpdateInspectionForm() {}

// docDeleteInspectionForm godoc
//
//	@Summary	Delete inspection-forms
//	@Tags	inspection-forms
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-forms/{id} [delete]
func docDeleteInspectionForm() {}
