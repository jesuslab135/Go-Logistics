package handler

// docListEmployees godoc
//
//	@Summary	List employees
//	@Tags	employees
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.EmployeePage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/employees [get]
func docListEmployees() {}

// docCreateEmployee godoc
//
//	@Summary	Create employees
//	@Tags	employees
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateEmployeeRequest	true	"body"
//	@Success	201	{object}	dto.EmployeeResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/employees [post]
func docCreateEmployee() {}

// docGetEmployee godoc
//
//	@Summary	Get employees
//	@Tags	employees
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.EmployeeResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/employees/{id} [get]
func docGetEmployee() {}

// docUpdateEmployee godoc
//
//	@Summary	Update employees
//	@Tags	employees
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateEmployeeRequest	true	"body"
//	@Success	200	{object}	dto.EmployeeResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/employees/{id} [put]
func docUpdateEmployee() {}

// docDeleteEmployee godoc
//
//	@Summary	Delete employees
//	@Tags	employees
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/employees/{id} [delete]
func docDeleteEmployee() {}
