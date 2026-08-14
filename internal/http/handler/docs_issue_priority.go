package handler

// docListIssuePriorities godoc
//
//	@Summary	List issue-priorities
//	@Tags	issue-priorities
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.IssuePriorityPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/issue-priorities [get]
func docListIssuePriorities() {}

// docCreateIssuePriority godoc
//
//	@Summary	Create issue-priorities
//	@Tags	issue-priorities
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateIssuePriorityRequest	true	"body"
//	@Success	201	{object}	dto.IssuePriorityResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/issue-priorities [post]
func docCreateIssuePriority() {}

// docGetIssuePriority godoc
//
//	@Summary	Get issue-priorities
//	@Tags	issue-priorities
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.IssuePriorityResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/issue-priorities/{id} [get]
func docGetIssuePriority() {}

// docUpdateIssuePriority godoc
//
//	@Summary	Update issue-priorities
//	@Tags	issue-priorities
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateIssuePriorityRequest	true	"body"
//	@Success	200	{object}	dto.IssuePriorityResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/issue-priorities/{id} [put]
func docUpdateIssuePriority() {}

// docDeleteIssuePriority godoc
//
//	@Summary	Delete issue-priorities
//	@Tags	issue-priorities
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/issue-priorities/{id} [delete]
func docDeleteIssuePriority() {}
