package handler

// docListRoles godoc
//
//	@Summary	List roles
//	@Tags	roles
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.RolePage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/roles [get]
func docListRoles() {}

// docCreateRole godoc
//
//	@Summary	Create roles
//	@Tags	roles
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateRoleRequest	true	"body"
//	@Success	201	{object}	dto.RoleResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/roles [post]
func docCreateRole() {}

// docGetRole godoc
//
//	@Summary	Get roles
//	@Tags	roles
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.RoleResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/roles/{id} [get]
func docGetRole() {}

// docUpdateRole godoc
//
//	@Summary	Update roles
//	@Tags	roles
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateRoleRequest	true	"body"
//	@Success	200	{object}	dto.RoleResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/roles/{id} [put]
func docUpdateRole() {}

// docDeleteRole godoc
//
//	@Summary	Delete roles
//	@Tags	roles
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/roles/{id} [delete]
func docDeleteRole() {}
