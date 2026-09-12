package dto

// ImportReport documents the body of a bulk import (201, or 200 for a dry
// run) and the details of its 422 import_failed error. It mirrors
// bulk.Report, which swag does not read; TestImportReportDocMatchesEngine
// keeps the two in step.
type ImportReport struct {
	Resource string           `json:"resource"`
	DryRun   bool             `json:"dry_run"`
	Rows     int              `json:"rows"`
	Created  int              `json:"created"`
	Errors   []ImportRowError `json:"errors"`
}

// ImportRowError is one problem, located by spreadsheet row (the header is row 1).
type ImportRowError struct {
	Row     int    `json:"row"`
	Column  string `json:"column,omitempty"`
	Message string `json:"message"`
}
