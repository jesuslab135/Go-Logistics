package handler

// docBulkImport godoc
//
//	@Summary		Bulk import from a spreadsheet
//	@Description	Available on every importable section (vendors, parts, assets, employees, work-orders…; see README "Bulk import and export"). Upload the section's template, filled in, as .xlsx or .csv. All-or-nothing: if any row fails, nothing is imported and the 422 lists every problem by row and column. dry_run=true runs the whole import and rolls it back. Reference columns accept the referenced record's name or id. Requires the section's create permission.
//	@Tags			bulk
//	@Security		BearerAuth
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			section	path		string	true	"Section path, e.g. vendors"
//	@Param			file	formData	file	true	".xlsx or .csv"
//	@Param			dry_run	query		bool	false	"Validate and roll back"
//	@Success		201		{object}	dto.ImportReport
//	@Success		200		{object}	dto.ImportReport	"dry run"
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		413		{object}	dto.ErrorResponse
//	@Failure		415		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse	"import_failed; details is an ImportReport"
//	@Failure		429		{object}	dto.ErrorResponse	"too many imports already running"
//	@Router			/api/v1/{section}/import [post]
func docBulkImport() {}

// docBulkTemplate godoc
//
//	@Summary		Download a section's Excel template
//	@Description	Sheets: Datos (the header row to fill in), Instrucciones (every column explained) and Catálogos (the names reference columns accept, from the caller's company). Custom fields appear as cf.<key> columns. Requires the section's read permission.
//	@Tags			bulk
//	@Security		BearerAuth
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param			section	path	string	true	"Section path, e.g. vendors"
//	@Success		200		{file}	file
//	@Failure		403		{object}	dto.ErrorResponse
//	@Router			/api/v1/{section}/import/template [get]
func docBulkTemplate() {}

// docBulkExport godoc
//
//	@Summary		Export a list to Excel or CSV
//	@Description	Every row of the list, with the list's own filters (pass them as on the list) and permissions. Up to EXPORT_MAX_ROWS rows; X-Export-Truncated is set when the list is longer. A file longer than IMPORT_MAX_ROWS must be split before it can be imported again.
//	@Tags			bulk
//	@Security		BearerAuth
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Produce		text/csv
//	@Param			section	path	string	true	"List path, e.g. work-orders"
//	@Param			format	query	string	false	"xlsx (default) or csv"
//	@Success		200		{file}	file
//	@Failure		400		{object}	dto.ErrorResponse	"format is neither xlsx nor csv"
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse	"too many exports already running"
//	@Router			/api/v1/{section}/export [get]
func docBulkExport() {}
