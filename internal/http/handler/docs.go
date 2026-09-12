package handler

// The company resource is served by a generic crud.Handler, so its OpenAPI
// operations are documented here as annotation-only stubs (swag reads the
// comments; the empty bodies are never called).

// docListCompanies godoc
//
//	@Summary	List companies (paginated)
//	@Tags		companies
//	@Security	BearerAuth
//	@Produce	json
//	@Param		limit		query		int	false	"Page size"
//	@Param		offset		query		int	false	"Offset"
//	@Success	200			{object}	dto.CompanyPage
//	@Failure	401			{object}	dto.ErrorResponse
//	@Router		/api/v1/companies [get]
func docListCompanies() {}

// docCreateCompany godoc
//
//	@Summary	Create a company
//	@Tags		companies
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		company	body		dto.CreateCompanyRequest	true	"Company"
//	@Success	201		{object}	dto.CompanyResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Failure	409		{object}	dto.ErrorResponse
//	@Router		/api/v1/companies [post]
func docCreateCompany() {}

// docGetCompany godoc
//
//	@Summary	Get a company by id
//	@Tags		companies
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		int	true	"Company id"
//	@Success	200	{object}	dto.CompanyResponse
//	@Failure	401	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/api/v1/companies/{id} [get]
func docGetCompany() {}

// docUpdateCompany godoc
//
//	@Summary	Update a company
//	@Tags		companies
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int							true	"Company id"
//	@Param		company	body		dto.UpdateCompanyRequest	true	"Company"
//	@Success	200		{object}	dto.CompanyResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/api/v1/companies/{id} [put]
func docUpdateCompany() {}

// docDeleteCompany godoc
//
//	@Summary	Delete a company
//	@Tags		companies
//	@Security	BearerAuth
//	@Param		id	path	int	true	"Company id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router		/api/v1/companies/{id} [delete]
func docDeleteCompany() {}

// docHealth godoc
//
//	@Summary	Liveness check
//	@Tags		system
//	@Produce	json
//	@Success	200	{object}	map[string]string
//	@Failure	503	{object}	map[string]string
//	@Router		/healthz [get]
func docHealth() {}
