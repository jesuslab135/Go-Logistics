package handler

// docNListTireMountLogs godoc
//
//	@Summary	List tire-mount-logs
//	@Tags	tire-mount-logs
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.TireMountLogPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/mount-logs [get]
func docNListTireMountLogs() {}

// docNCreateTireMountLog godoc
//
//	@Summary	Create tire-mount-logs
//	@Tags	tire-mount-logs
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateTireMountLogRequest	true	"body"
//	@Success	201	{object}	dto.TireMountLogResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/mount-logs [post]
func docNCreateTireMountLog() {}

// docNGetTireMountLog godoc
//
//	@Summary	Get tire-mount-logs
//	@Tags	tire-mount-logs
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.TireMountLogResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/mount-logs/{child_id} [get]
func docNGetTireMountLog() {}

// docNUpdateTireMountLog godoc
//
//	@Summary	Update tire-mount-logs
//	@Tags	tire-mount-logs
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateTireMountLogRequest	true	"body"
//	@Success	200	{object}	dto.TireMountLogResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/mount-logs/{child_id} [put]
func docNUpdateTireMountLog() {}

// docNDeleteTireMountLog godoc
//
//	@Summary	Delete tire-mount-logs
//	@Tags	tire-mount-logs
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/mount-logs/{child_id} [delete]
func docNDeleteTireMountLog() {}
