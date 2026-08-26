package handler

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/db/gen"
	"fleet/internal/domain/customfield"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/filter"
	"fleet/internal/platform/paginate"
)

// customFieldResources are the resources that carry a custom_fields column.
// A definition naming anything else would describe a form nothing renders, so
// it is refused rather than stored as a fact about nothing.
var customFieldResources = []string{
	"assets", "employees", "issues", "parts",
	"purchase-orders", "service-entries", "vendors", "work-orders",
}

type CustomFieldDefinitionStore struct{ q *gen.Queries }

func NewCustomFieldDefinitionStore(q *gen.Queries) *CustomFieldDefinitionStore {
	return &CustomFieldDefinitionStore{q: q}
}

func (s *CustomFieldDefinitionStore) List(ctx context.Context, p paginate.Params) ([]dto.CustomFieldDefinitionResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	// The resource filter is what a form uses: it asks for its own fields, not
	// for every field the company has declared anywhere.
	resource := resourceFilter(ctx)

	rows, err := s.q.ListCustomFieldDefinitions(ctx, gen.ListCustomFieldDefinitionsParams{
		CompanyID: company, Resource: resource, Lim: int32(p.Limit), Off: int32(p.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountCustomFieldDefinitions(ctx, gen.CountCustomFieldDefinitionsParams{
		CompanyID: company, Resource: resource,
	})
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.CustomFieldDefinitionResponse, len(rows))
	for i, r := range rows {
		out[i] = toCustomFieldDefinitionResponse(r)
	}
	return out, total, nil
}

func (s *CustomFieldDefinitionStore) Get(ctx context.Context, id int64) (dto.CustomFieldDefinitionResponse, error) {
	r, err := s.q.GetCustomFieldDefinition(ctx, gen.GetCustomFieldDefinitionParams{
		ID: id, CompanyID: middleware.CompanyFromContext(ctx),
	})
	if err != nil {
		return dto.CustomFieldDefinitionResponse{}, err
	}
	return toCustomFieldDefinitionResponse(r), nil
}

func (s *CustomFieldDefinitionStore) Create(ctx context.Context, in dto.CreateCustomFieldDefinitionRequest) (dto.CustomFieldDefinitionResponse, error) {
	if err := validateDefinition(in.Resource, in.Key, in.FieldType, in.Options); err != nil {
		return dto.CustomFieldDefinitionResponse{}, err
	}

	now := time.Now().UTC()
	r, err := s.q.CreateCustomFieldDefinition(ctx, gen.CreateCustomFieldDefinitionParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Resource:  in.Resource,
		Key:       in.Key,
		Label:     in.Label,
		FieldType: in.FieldType,
		Required:  in.Required,
		Options:   optionsJSON(in.Options),
		Position:  in.Position,
		CreatedAt: now,
	})
	if err != nil {
		return dto.CustomFieldDefinitionResponse{}, err
	}
	return toCustomFieldDefinitionResponse(r), nil
}

// Update cannot change the key or the resource: every stored document is keyed
// by the one and filed under the other, so changing either would orphan the
// values already written.
func (s *CustomFieldDefinitionStore) Update(ctx context.Context, id int64, in dto.UpdateCustomFieldDefinitionRequest) (dto.CustomFieldDefinitionResponse, error) {
	if err := validateDefinition("", "", in.FieldType, in.Options); err != nil {
		return dto.CustomFieldDefinitionResponse{}, err
	}

	r, err := s.q.UpdateCustomFieldDefinition(ctx, gen.UpdateCustomFieldDefinitionParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Label:     in.Label,
		FieldType: in.FieldType,
		Required:  in.Required,
		Options:   optionsJSON(in.Options),
		Position:  in.Position,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return dto.CustomFieldDefinitionResponse{}, err
	}
	return toCustomFieldDefinitionResponse(r), nil
}

// Delete stops validating a key. The values already stored under it stay in
// their documents: they are the tenant's data, and nothing else records them.
func (s *CustomFieldDefinitionStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteCustomFieldDefinition(ctx, gen.DeleteCustomFieldDefinitionParams{
		ID: id, CompanyID: middleware.CompanyFromContext(ctx),
	})
}

// validateDefinition checks what the database cannot. An empty resource or key
// means the caller is updating, where neither may change.
func validateDefinition(resource, key, fieldType string, options []string) error {
	details := map[string]string{}

	if resource != "" && !slices.Contains(customFieldResources, resource) {
		details["resource"] = "must be one of: " + strings.Join(customFieldResources, ", ")
	}
	if key != "" && !validCustomFieldKey(key) {
		details["key"] = "must be lowercase letters, digits and underscores, starting with a letter"
	}
	if !customfield.ValidType(fieldType) {
		details["field_type"] = "must be one of: " + strings.Join(customfield.Types, ", ")
	}
	// A select with nothing to select from is a field nobody can fill in.
	if fieldType == customfield.TypeSelect && len(options) == 0 {
		details["options"] = "a select field needs at least one option"
	}
	if fieldType != customfield.TypeSelect && len(options) > 0 {
		details["options"] = "only a select field has options"
	}

	if len(details) > 0 {
		return apierr.Validation(details)
	}
	return nil
}

// validCustomFieldKey constrains a key to what can safely be a JSON key, a
// query parameter and a form field name at once — which is what it has to be.
func validCustomFieldKey(key string) bool {
	if key == "" || key[0] < 'a' || key[0] > 'z' {
		return false
	}
	for _, r := range key {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
		default:
			return false
		}
	}
	return true
}

func optionsJSON(options []string) []byte {
	if options == nil {
		options = []string{}
	}
	raw, err := json.Marshal(options)
	if err != nil {
		return []byte("[]")
	}
	return raw
}

func toCustomFieldDefinitionResponse(r gen.CustomFieldDefinition) dto.CustomFieldDefinitionResponse {
	options := []string{}
	_ = json.Unmarshal(r.Options, &options)

	return dto.CustomFieldDefinitionResponse{
		ID:        r.ID,
		CompanyID: r.CompanyID,
		Resource:  r.Resource,
		Key:       r.Key,
		Label:     r.Label,
		FieldType: r.FieldType,
		Required:  r.Required,
		Options:   options,
		Position:  r.Position,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// definitionsFor loads a company's definitions for one resource, in the shape
// the domain validator takes.
func definitionsFor(ctx context.Context, q *gen.Queries, resource string) ([]customfield.Definition, error) {
	rows, err := q.DefinitionsForResource(ctx, gen.DefinitionsForResourceParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Resource:  resource,
	})
	if err != nil {
		return nil, err
	}

	out := make([]customfield.Definition, len(rows))
	for i, r := range rows {
		options := []string{}
		_ = json.Unmarshal(r.Options, &options)
		out[i] = customfield.Definition{
			Key:      r.Key,
			Label:    r.Label,
			Type:     r.FieldType,
			Required: r.Required,
			Options:  options,
		}
	}
	return out, nil
}

// validateCustomFields is what every write to a custom_fields-carrying resource
// calls. A company with no definitions for that resource is the common case and
// costs one indexed lookup.
func validateCustomFields(ctx context.Context, q *gen.Queries, resource string, raw json.RawMessage) error {
	defs, err := definitionsFor(ctx, q, resource)
	if err != nil {
		return err
	}
	if details := customfield.Validate(raw, defs); details != nil {
		return apierr.Validation(details)
	}
	return nil
}

// resourceFilter reads ?resource= from the request, carried on the context
// because the generic CRUD handler owns the List call.
func resourceFilter(ctx context.Context) *string {
	v, _ := ctx.Value(resourceFilterKey{}).(string)
	if v == "" {
		return nil
	}
	return &v
}

type resourceFilterKey struct{}

// WithResourceFilter reads ?resource= for the definitions list.
func WithResourceFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if v := c.Query("resource"); v != "" {
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), resourceFilterKey{}, v))
		}
		c.Next()
	}
}

// applyCustomFieldFilters adds a jsonb containment predicate for every
// ?custom_fields.<key>= parameter on the request.
//
// Containment (@>) rather than a key extraction, because it is what the GIN
// index on the column can answer; ->> would be a sequential scan. The value is
// converted through the key's own definition, so custom_fields.axles=3 matches
// the number 3 rather than the string "3" — a mismatch there returns nothing at
// all, which reads as "no results" instead of "wrong query".
//
// An unknown key is refused rather than ignored: silently returning every row
// for a filter the caller believes is applied is worse than saying no.
func applyCustomFieldFilters(c *gin.Context, where *filter.Where, column, resource string, q *gen.Queries) error {
	const prefix = "custom_fields."

	var keys []string
	for name := range c.Request.URL.Query() {
		if strings.HasPrefix(name, prefix) {
			keys = append(keys, name)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	slices.Sort(keys) // stable predicate order, so the same request is the same SQL

	defs, err := definitionsFor(c.Request.Context(), q, resource)
	if err != nil {
		return err
	}
	byKey := make(map[string]customfield.Definition, len(defs))
	for _, d := range defs {
		byKey[d.Key] = d
	}

	details := map[string]string{}
	for _, name := range keys {
		key := strings.TrimPrefix(name, prefix)
		def, ok := byKey[key]
		if !ok {
			details[name] = "no custom field named " + key + " is defined for " + resource
			continue
		}
		value, err := customfield.FilterValue(def, c.Query(name))
		if err != nil {
			details[name] = err.Error()
			continue
		}
		where.Raw(column+" @> ?", string(mustObject(key, value)))
	}

	if len(details) > 0 {
		return apierr.Validation(details)
	}
	return nil
}

// mustObject wraps a scalar as the single-key document containment compares
// against: {"axles": 3}.
func mustObject(key string, value json.RawMessage) json.RawMessage {
	doc, err := json.Marshal(map[string]json.RawMessage{key: value})
	if err != nil {
		return json.RawMessage("{}")
	}
	return doc
}
