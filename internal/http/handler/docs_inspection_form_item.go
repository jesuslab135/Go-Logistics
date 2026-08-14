package handler

// docNListInspectionFormItems godoc
//
//	@Summary	List inspection-form-items
//	@Tags	inspection-form-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.InspectionFormItemPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-forms/{id}/items [get]
func docNListInspectionFormItems() {}

// docNCreateInspectionFormItem godoc
//
//	@Summary	Create inspection-form-items
//	@Tags	inspection-form-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateInspectionFormItemRequest	true	"body"
//	@Success	201	{object}	dto.InspectionFormItemResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-forms/{id}/items [post]
func docNCreateInspectionFormItem() {}

// docNGetInspectionFormItem godoc
//
//	@Summary	Get inspection-form-items
//	@Tags	inspection-form-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.InspectionFormItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-forms/{id}/items/{child_id} [get]
func docNGetInspectionFormItem() {}

// docNUpdateInspectionFormItem godoc
//
//	@Summary	Update inspection-form-items
//	@Tags	inspection-form-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateInspectionFormItemRequest	true	"body"
//	@Success	200	{object}	dto.InspectionFormItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-forms/{id}/items/{child_id} [put]
func docNUpdateInspectionFormItem() {}

// docNDeleteInspectionFormItem godoc
//
//	@Summary	Delete inspection-form-items
//	@Tags	inspection-form-items
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-forms/{id}/items/{child_id} [delete]
func docNDeleteInspectionFormItem() {}
