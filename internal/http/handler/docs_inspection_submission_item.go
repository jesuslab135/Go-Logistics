package handler

// docNListInspectionSubmissionItems godoc
//
//	@Summary	List inspection-submission-items
//	@Tags	inspection-submission-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.InspectionSubmissionItemPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-submissions/{id}/items [get]
func docNListInspectionSubmissionItems() {}

// docNCreateInspectionSubmissionItem godoc
//
//	@Summary	Create inspection-submission-items
//	@Tags	inspection-submission-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateInspectionSubmissionItemRequest	true	"body"
//	@Success	201	{object}	dto.InspectionSubmissionItemResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-submissions/{id}/items [post]
func docNCreateInspectionSubmissionItem() {}

// docNGetInspectionSubmissionItem godoc
//
//	@Summary	Get inspection-submission-items
//	@Tags	inspection-submission-items
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.InspectionSubmissionItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-submissions/{id}/items/{child_id} [get]
func docNGetInspectionSubmissionItem() {}

// docNUpdateInspectionSubmissionItem godoc
//
//	@Summary	Update inspection-submission-items
//	@Tags	inspection-submission-items
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateInspectionSubmissionItemRequest	true	"body"
//	@Success	200	{object}	dto.InspectionSubmissionItemResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-submissions/{id}/items/{child_id} [put]
func docNUpdateInspectionSubmissionItem() {}

// docNDeleteInspectionSubmissionItem godoc
//
//	@Summary	Delete inspection-submission-items
//	@Tags	inspection-submission-items
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-submissions/{id}/items/{child_id} [delete]
func docNDeleteInspectionSubmissionItem() {}
