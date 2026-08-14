package handler

// docCreateIssue godoc
//
//	@Summary	Create issues
//	@Tags	issues
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateIssueRequest	true	"body"
//	@Success	201	{object}	dto.IssueResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/issues [post]
func docCreateIssue() {}

// docGetIssue godoc
//
//	@Summary	Get issues
//	@Tags	issues
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.IssueResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/issues/{id} [get]
func docGetIssue() {}

// docUpdateIssue godoc
//
//	@Summary	Update issues
//	@Tags	issues
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateIssueRequest	true	"body"
//	@Success	200	{object}	dto.IssueResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/issues/{id} [put]
func docUpdateIssue() {}

// docDeleteIssue godoc
//
//	@Summary	Delete issues
//	@Tags	issues
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/issues/{id} [delete]
func docDeleteIssue() {}
