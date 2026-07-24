package issue

// Fault — port of api/models/issue_model.py:32
// AppliesToAssetTypes is a Postgres text[] (ArrayField) with a GIN index;
// empty means "applies to all types". Code is generated in Django's save()
// as FAL-NNN (:96) — that belongs in a service.

type Fault struct {
	ID                  int64
	CompanyID           int64
	Family              string
	Code                string
	Name                string
	Description         string
	AppliesToAssetTypes []string
}
