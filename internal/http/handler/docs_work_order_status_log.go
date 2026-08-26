package handler

// The status log is read-only. POST, PUT and DELETE still route — they answer
// 405 `route_retired` naming the replacement rather than 404 — but they are not
// documented as operations, because they are not ones a client should call.

// docNListWorkOrderStatusLogs godoc
//
//	@Summary		List work-order-status-logs
//	@Description	Append-only status history, newest first. Rows are written by the operation that changes the work order's status, inside the same transaction, so this history cannot disagree with the work order. actor_type is "employee" or "system"; actor_employee_id is null for a system transition and for rows written before actors were recorded.
//	@Tags			work-order-status-logs
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id		path	int	true	"parent id"
//	@Param			limit	query	int	false	"Page size"
//	@Param			offset	query	int	false	"Offset"
//	@Success		200		{object}	dto.WorkOrderStatusLogPage
//	@Failure		401		{object}	dto.ErrorResponse
//	@Router			/api/v1/work-orders/{id}/status-logs [get]
func docNListWorkOrderStatusLogs() {}

// docNGetWorkOrderStatusLog godoc
//
//	@Summary	Get work-order-status-logs
//	@Tags		work-order-status-logs
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id			path	int	true	"parent id"
//	@Param		child_id	path	int	true	"id"
//	@Success	200			{object}	dto.WorkOrderStatusLogResponse
//	@Failure	404			{object}	dto.ErrorResponse
//	@Router		/api/v1/work-orders/{id}/status-logs/{child_id} [get]
func docNGetWorkOrderStatusLog() {}
