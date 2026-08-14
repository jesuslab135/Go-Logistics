package handler

// docListGroups godoc
//
//	@Summary	List groups
//	@Tags	groups
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.GroupPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/groups [get]
func docListGroups() {}

// docCreateGroup godoc
//
//	@Summary	Create groups
//	@Tags	groups
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateGroupRequest	true	"body"
//	@Success	201	{object}	dto.GroupResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/groups [post]
func docCreateGroup() {}

// docGetGroup godoc
//
//	@Summary	Get groups
//	@Tags	groups
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.GroupResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/groups/{id} [get]
func docGetGroup() {}

// docUpdateGroup godoc
//
//	@Summary	Update groups
//	@Tags	groups
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateGroupRequest	true	"body"
//	@Success	200	{object}	dto.GroupResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/groups/{id} [put]
func docUpdateGroup() {}

// docDeleteGroup godoc
//
//	@Summary	Delete groups
//	@Tags	groups
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/groups/{id} [delete]
func docDeleteGroup() {}
