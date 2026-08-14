package handler

// docNListTireInstallations godoc
//
//	@Summary	List tire-installations
//	@Tags	tire-installations
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.TireInstallationPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/installations [get]
func docNListTireInstallations() {}

// docNCreateTireInstallation godoc
//
//	@Summary	Create tire-installations
//	@Tags	tire-installations
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateTireInstallationRequest	true	"body"
//	@Success	201	{object}	dto.TireInstallationResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/installations [post]
func docNCreateTireInstallation() {}

// docNGetTireInstallation godoc
//
//	@Summary	Get tire-installations
//	@Tags	tire-installations
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.TireInstallationResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/installations/{child_id} [get]
func docNGetTireInstallation() {}

// docNUpdateTireInstallation godoc
//
//	@Summary	Update tire-installations
//	@Tags	tire-installations
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateTireInstallationRequest	true	"body"
//	@Success	200	{object}	dto.TireInstallationResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/installations/{child_id} [put]
func docNUpdateTireInstallation() {}

// docNDeleteTireInstallation godoc
//
//	@Summary	Delete tire-installations
//	@Tags	tire-installations
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/installations/{child_id} [delete]
func docNDeleteTireInstallation() {}
