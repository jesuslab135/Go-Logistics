package handler

// Custom field definitions describe what may go in a resource's custom_fields
// document — the jsonb column eight resources carry, which until now accepted
// anything and validated nothing.
//
// They are admin-gated because they shape everyone's forms, and scoped to a
// company and a resource: a field meaningful on an asset is rarely meaningful
// on a purchase order, and one global set would put every field on every form.
//
// Writes to those resources validate against these definitions. Keys nobody has
// declared are preserved rather than rejected: they are the tenant's data, and
// refusing them would make records saved before a definition existed uneditable
// through a form that no longer shows the field holding them.
//
// Lists filter on them with ?custom_fields.<key>=<value> — see the individual
// registers. An undefined key there is a 422, because silently returning every
// row for a filter the caller believes is applied is worse than saying no.

// docListCustomFieldDefinitions godoc
//
//	@Summary		List custom field definitions
//	@Description	Pass ?resource= to get just one resource's fields, which is what a form asks for. Ordered by position.
//	@Tags			custom-field-definitions
//	@Security		BearerAuth
//	@Produce		json
//	@Param			resource	query	string	false	"assets, employees, issues, parts, purchase-orders, service-entries, vendors or work-orders"
//	@Param			limit		query	int		false	"Page size"
//	@Param			offset		query	int		false	"Offset"
//	@Success		200			{object}	dto.CustomFieldDefinitionPage
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Router			/api/v1/custom-field-definitions [get]
func docListCustomFieldDefinitions() {}

// docCreateCustomFieldDefinition godoc
//
//	@Summary		Declare a custom field
//	@Description	field_type is text, number, date, boolean or select; a select needs options and nothing else may have them. The key is immutable once created, because every stored document is keyed by it. A date value is a business date (YYYY-MM-DD), not a timestamp.
//	@Tags			custom-field-definitions
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body	dto.CreateCustomFieldDefinitionRequest	true	"body"
//	@Success		201		{object}	dto.CustomFieldDefinitionResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/custom-field-definitions [post]
func docCreateCustomFieldDefinition() {}

// docGetCustomFieldDefinition godoc
//
//	@Summary	Get a custom field definition
//	@Tags		custom-field-definitions
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	int	true	"id"
//	@Success	200	{object}	dto.CustomFieldDefinitionResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/api/v1/custom-field-definitions/{id} [get]
func docGetCustomFieldDefinition() {}

// docUpdateCustomFieldDefinition godoc
//
//	@Summary		Update a custom field definition
//	@Description	The key and resource cannot change: every stored document is keyed by the one and filed under the other, so changing either would orphan the values already written.
//	@Tags			custom-field-definitions
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int										true	"id"
//	@Param			body	body	dto.UpdateCustomFieldDefinitionRequest	true	"body"
//	@Success		200		{object}	dto.CustomFieldDefinitionResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/custom-field-definitions/{id} [put]
func docUpdateCustomFieldDefinition() {}

// docDeleteCustomFieldDefinition godoc
//
//	@Summary		Delete a custom field definition
//	@Description	Stops validating that key. Values already stored under it stay in their documents: they are the tenant's data, and nothing else records them.
//	@Tags			custom-field-definitions
//	@Security		BearerAuth
//	@Param			id	path	int	true	"id"
//	@Success		204	"No Content"
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/api/v1/custom-field-definitions/{id} [delete]
func docDeleteCustomFieldDefinition() {}
