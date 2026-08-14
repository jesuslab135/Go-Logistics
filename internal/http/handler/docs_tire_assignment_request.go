package handler

// docListTireAssignmentRequests godoc
//
//	@Summary	List tire-assignment-requests
//	@Tags	tire-assignment-requests
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.TireAssignmentRequestPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tire-assignment-requests [get]
func docListTireAssignmentRequests() {}

// docCreateTireAssignmentRequest godoc
//
//	@Summary	Create tire-assignment-requests
//	@Tags	tire-assignment-requests
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateTireAssignmentRequestRequest	true	"body"
//	@Success	201	{object}	dto.TireAssignmentRequestResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/tire-assignment-requests [post]
func docCreateTireAssignmentRequest() {}

// docGetTireAssignmentRequest godoc
//
//	@Summary	Get tire-assignment-requests
//	@Tags	tire-assignment-requests
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.TireAssignmentRequestResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tire-assignment-requests/{id} [get]
func docGetTireAssignmentRequest() {}

// docUpdateTireAssignmentRequest godoc
//
//	@Summary	Update tire-assignment-requests
//	@Tags	tire-assignment-requests
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateTireAssignmentRequestRequest	true	"body"
//	@Success	200	{object}	dto.TireAssignmentRequestResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tire-assignment-requests/{id} [put]
func docUpdateTireAssignmentRequest() {}

// docDeleteTireAssignmentRequest godoc
//
//	@Summary	Delete tire-assignment-requests
//	@Tags	tire-assignment-requests
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tire-assignment-requests/{id} [delete]
func docDeleteTireAssignmentRequest() {}
