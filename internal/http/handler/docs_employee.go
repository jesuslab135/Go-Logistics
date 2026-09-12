package handler

// docListEmployees godoc
//
//	@Summary		List employees
//	@Description	The members of the session's company. role_id, role_name and is_active describe each person's membership in THIS company; is_account_owner is computed from account ownership.
//	@Tags			employees
//	@Security		BearerAuth
//	@Produce		json
//	@Param			limit	query		int	false	"Page size"
//	@Param			offset	query		int	false	"Offset"
//	@Success		200		{object}	dto.EmployeePage
//	@Failure		401		{object}	dto.ErrorResponse
//	@Router			/api/v1/employees [get]
func docListEmployees() {}

// docCreateEmployee godoc
//
//	@Summary		Create employees
//	@Description	Creates the employee with one membership, in the session's company. role_id gives that membership a role and requires an administrator of the company (403 otherwise; 422 when the role belongs to another company). is_active false creates the membership suspended.
//	@Tags			employees
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateEmployeeRequest	true	"body"
//	@Success		201		{object}	dto.EmployeeResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/employees [post]
func docCreateEmployee() {}

// docGetEmployee godoc
//
//	@Summary	Get employees
//	@Tags		employees
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		int	true	"id"
//	@Success	200	{object}	dto.EmployeeResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/api/v1/employees/{id} [get]
func docGetEmployee() {}

// docUpdateEmployee godoc
//
//	@Summary		Update employees
//	@Description	Updates the profile. role_id and is_active apply to the membership in the session's company only, and an omitted field leaves it unchanged. is_active false suspends the employee here while they keep working in their other companies; reactivating also lifts a deactivation left by the old account-wide flag. Changing role_id requires an administrator of the company (403). Refused with 422: a role from another company, an administrator changing their own role (the account owner may), deactivating yourself, and deactivating the account owner. Every change is written to membership_audit.
//	@Tags			employees
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"id"
//	@Param			body	body		dto.UpdateEmployeeRequest	true	"body"
//	@Success		200		{object}	dto.EmployeeResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/employees/{id} [put]
func docUpdateEmployee() {}

// docDeleteEmployee godoc
//
//	@Summary		Delete employees
//	@Description	Removes the employee from the session's company. Someone who belongs to other companies stays in them; only when this was their last company is the person deleted, which fails with 409 foreign_key_violation when they have history (labor time, reported issues, fuel entries, inspections) — deactivate instead. The account owner and the caller themselves cannot be deleted (409). Written to membership_audit.
//	@Tags			employees
//	@Security		BearerAuth
//	@Param			id	path	int	true	"id"
//	@Success		204	"No Content"
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/api/v1/employees/{id} [delete]
func docDeleteEmployee() {}
