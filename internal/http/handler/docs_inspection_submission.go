package handler

// docListInspectionSubmissions godoc
//
//	@Summary	List inspection-submissions
//	@Tags	inspection-submissions
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.InspectionSubmissionPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-submissions [get]
func docListInspectionSubmissions() {}

// docCreateInspectionSubmission godoc
//
//	@Summary	Create inspection-submissions
//	@Tags	inspection-submissions
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateInspectionSubmissionRequest	true	"body"
//	@Success	201	{object}	dto.InspectionSubmissionResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-submissions [post]
func docCreateInspectionSubmission() {}

// docGetInspectionSubmission godoc
//
//	@Summary	Get inspection-submissions
//	@Tags	inspection-submissions
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.InspectionSubmissionResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-submissions/{id} [get]
func docGetInspectionSubmission() {}

// docUpdateInspectionSubmission godoc
//
//	@Summary	Update inspection-submissions
//	@Tags	inspection-submissions
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateInspectionSubmissionRequest	true	"body"
//	@Success	200	{object}	dto.InspectionSubmissionResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-submissions/{id} [put]
func docUpdateInspectionSubmission() {}

// docDeleteInspectionSubmission godoc
//
//	@Summary	Delete inspection-submissions
//	@Tags	inspection-submissions
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inspection-submissions/{id} [delete]
func docDeleteInspectionSubmission() {}
