package handler

// M2M join tables are served by a generic crud.LinkHandler, so their OpenAPI
// operations are annotation-only stubs. Each list returns the linked entities,
// not the join rows; POST is idempotent and answers 201 either way; DELETE
// identifies the link by the far side's id.

// docListIssueAssignees godoc
//
//	@Summary	List employees assigned to an issue
//	@Tags	issues
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"issue id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.EmployeePage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/issues/{id}/assigned-to [get]
func docListIssueAssignees() {}

// docAddIssueAssignee godoc
//
//	@Summary	Assign an employee to an issue
//	@Tags	issues
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"issue id"
//	@Param	body	body	dto.LinkEmployeeRequest	true	"body"
//	@Success	201	{object}	dto.EmployeeResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/issues/{id}/assigned-to [post]
func docAddIssueAssignee() {}

// docRemoveIssueAssignee godoc
//
//	@Summary	Unassign an employee from an issue
//	@Tags	issues
//	@Security	BearerAuth
//	@Param	id	path	int	true	"issue id"
//	@Param	employee_id	path	int	true	"employee id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/issues/{id}/assigned-to/{employee_id} [delete]
func docRemoveIssueAssignee() {}

// docListIssueWatchers godoc
//
//	@Summary	List employees watching an issue
//	@Tags	issues
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"issue id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.EmployeePage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/issues/{id}/watchers [get]
func docListIssueWatchers() {}

// docAddIssueWatcher godoc
//
//	@Summary	Add a watcher to an issue
//	@Tags	issues
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"issue id"
//	@Param	body	body	dto.LinkEmployeeRequest	true	"body"
//	@Success	201	{object}	dto.EmployeeResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/issues/{id}/watchers [post]
func docAddIssueWatcher() {}

// docRemoveIssueWatcher godoc
//
//	@Summary	Remove a watcher from an issue
//	@Tags	issues
//	@Security	BearerAuth
//	@Param	id	path	int	true	"issue id"
//	@Param	employee_id	path	int	true	"employee id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/issues/{id}/watchers/{employee_id} [delete]
func docRemoveIssueWatcher() {}

// docListWorkOrderIssues godoc
//
//	@Summary	List issues linked to a work order
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"work order id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.IssuePage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/issues [get]
func docListWorkOrderIssues() {}

// docAddWorkOrderIssue godoc
//
//	@Summary	Link an issue to a work order
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"work order id"
//	@Param	body	body	dto.LinkIssueRequest	true	"body"
//	@Success	201	{object}	dto.IssueResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/issues [post]
func docAddWorkOrderIssue() {}

// docRemoveWorkOrderIssue godoc
//
//	@Summary	Unlink an issue from a work order
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Param	id	path	int	true	"work order id"
//	@Param	issue_id	path	int	true	"issue id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/issues/{issue_id} [delete]
func docRemoveWorkOrderIssue() {}

// docListWorkOrderFaults godoc
//
//	@Summary	List faults linked to a work order
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"work order id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.FaultPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/faults [get]
func docListWorkOrderFaults() {}

// docAddWorkOrderFault godoc
//
//	@Summary	Link a fault to a work order
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"work order id"
//	@Param	body	body	dto.LinkFaultRequest	true	"body"
//	@Success	201	{object}	dto.FaultResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/faults [post]
func docAddWorkOrderFault() {}

// docRemoveWorkOrderFault godoc
//
//	@Summary	Unlink a fault from a work order
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Param	id	path	int	true	"work order id"
//	@Param	fault_id	path	int	true	"fault id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-orders/{id}/faults/{fault_id} [delete]
func docRemoveWorkOrderFault() {}

// docListServiceEntryLineItemIssues godoc
//
//	@Summary	List issues linked to a service entry line item
//	@Tags	service-entries
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"service entry line item id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.IssuePage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entry-line-items/{id}/issues [get]
func docListServiceEntryLineItemIssues() {}

// docAddServiceEntryLineItemIssue godoc
//
//	@Summary	Link an issue to a service entry line item
//	@Tags	service-entries
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"service entry line item id"
//	@Param	body	body	dto.LinkIssueRequest	true	"body"
//	@Success	201	{object}	dto.IssueResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entry-line-items/{id}/issues [post]
func docAddServiceEntryLineItemIssue() {}

// docRemoveServiceEntryLineItemIssue godoc
//
//	@Summary	Unlink an issue from a service entry line item
//	@Tags	service-entries
//	@Security	BearerAuth
//	@Param	id	path	int	true	"service entry line item id"
//	@Param	issue_id	path	int	true	"issue id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-entry-line-items/{id}/issues/{issue_id} [delete]
func docRemoveServiceEntryLineItemIssue() {}

// docListWorkOrderLineItemIssues godoc
//
//	@Summary	List issues linked to a work order line item
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"work order line item id"
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.IssuePage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-line-items/{id}/issues [get]
func docListWorkOrderLineItemIssues() {}

// docAddWorkOrderLineItemIssue godoc
//
//	@Summary	Link an issue to a work order line item
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"work order line item id"
//	@Param	body	body	dto.LinkIssueRequest	true	"body"
//	@Success	201	{object}	dto.IssueResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-line-items/{id}/issues [post]
func docAddWorkOrderLineItemIssue() {}

// docRemoveWorkOrderLineItemIssue godoc
//
//	@Summary	Unlink an issue from a work order line item
//	@Tags	work-orders
//	@Security	BearerAuth
//	@Param	id	path	int	true	"work order line item id"
//	@Param	issue_id	path	int	true	"issue id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/work-order-line-items/{id}/issues/{issue_id} [delete]
func docRemoveWorkOrderLineItemIssue() {}
