package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// specPath is the checked-in swag output. These tests read the generated file
// rather than the annotations so a stale regeneration is caught too.
func loadSpec(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "swagger.json"))
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	var spec map[string]any
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	return spec
}

func eachOperation(t *testing.T, fn func(path, method string, params []any)) {
	t.Helper()
	paths, ok := loadSpec(t)["paths"].(map[string]any)
	if !ok {
		t.Fatal("spec has no paths object")
	}
	for path, item := range paths {
		methods, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for method, op := range methods {
			body, ok := op.(map[string]any)
			if !ok {
				continue
			}
			params, _ := body["parameters"].([]any)
			fn(path, method, params)
		}
	}
}

// OpenAPI requires parameters to be unique by (name, in). Orval rejects the
// whole spec on a violation, not just the offending operation.
func TestGeneratedSpecHasNoDuplicateParams(t *testing.T) {
	eachOperation(t, func(path, method string, params []any) {
		seen := make(map[string]bool, len(params))
		for _, raw := range params {
			p, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			key, _ := p["name"].(string)
			in, _ := p["in"].(string)
			key += " in:" + in
			if seen[key] {
				t.Errorf("%s %s declares %s twice", method, path, key)
			}
			seen[key] = true
		}
	})
}

// Every collection route documents the limit/offset convention; page/page_size
// still works at runtime but the generated param types must not offer only it.
func TestGeneratedSpecUsesLimitOffset(t *testing.T) {
	eachOperation(t, func(path, method string, params []any) {
		for _, raw := range params {
			p, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if name, _ := p["name"].(string); name == "page_size" || name == "page" {
				t.Errorf("%s %s documents %q; use limit/offset", method, path, name)
			}
		}
	})
}

// definitionProps returns the properties map of a generated definition
// (e.g. "dto.MePermissionsResponse"), or fails the test if absent.
func definitionProps(t *testing.T, name string) map[string]any {
	t.Helper()
	defs, ok := loadSpec(t)["definitions"].(map[string]any)
	if !ok {
		t.Fatal("spec has no definitions object")
	}
	def, ok := defs[name].(map[string]any)
	if !ok {
		t.Fatalf("spec has no definition %q", name)
	}
	props, ok := def["properties"].(map[string]any)
	if !ok {
		t.Fatalf("definition %q has no properties", name)
	}
	return props
}

func TestMePermissionsExposesPlatformAdmin(t *testing.T) {
	props := definitionProps(t, "dto.MePermissionsResponse")
	prop, ok := props["is_platform_admin"].(map[string]any)
	if !ok {
		t.Fatal("MePermissionsResponse is missing is_platform_admin")
	}
	if prop["type"] != "boolean" {
		t.Errorf("is_platform_admin type = %v, want boolean", prop["type"])
	}
}

// custom_fields is a keyed object ({key: value}), never an array. json.RawMessage
// defaults to an int array in swaggo; every DTO must override it.
func TestCustomFieldsAreObjectsNotArrays(t *testing.T) {
	defs, ok := loadSpec(t)["definitions"].(map[string]any)
	if !ok {
		t.Fatal("spec has no definitions object")
	}
	found := 0
	for name, raw := range defs {
		def, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		props, ok := def["properties"].(map[string]any)
		if !ok {
			continue
		}
		cf, ok := props["custom_fields"].(map[string]any)
		if !ok {
			continue
		}
		found++
		if cf["type"] != "object" {
			t.Errorf("%s.custom_fields type = %v, want object", name, cf["type"])
		}
	}
	if found == 0 {
		t.Fatal("no custom_fields properties found in spec")
	}
}

func TestArchiveContractIsConcreteAndTyped(t *testing.T) {
	paths, _ := loadSpec(t)["paths"].(map[string]any)
	// No templated {resource} path may survive — that is the param the frontend patches.
	for path := range paths {
		if strings.Contains(path, "{resource}") {
			t.Errorf("spec still contains templated path %q", path)
		}
	}
	// Each concrete archive/restore op must have a schema-bearing 200.
	for _, res := range []string{"assets", "parts", "vendors", "service-tasks", "inspection-forms"} {
		for _, action := range []string{"archive", "restore"} {
			p := "/api/v1/" + res + "/{id}/" + action
			op, ok := paths[p].(map[string]any)
			if !ok {
				t.Errorf("missing path %s", p)
				continue
			}
			post, _ := op["post"].(map[string]any)
			resp, _ := post["responses"].(map[string]any)
			ok200, _ := resp["200"].(map[string]any)
			if _, hasSchema := ok200["schema"]; !hasSchema {
				t.Errorf("%s POST 200 has no schema", p)
			}
		}
	}
}

// The four computed amounts and four override fields are recomputed/mapped
// on read (see money_response_test.go); this guards against a forgotten
// swag regen leaving the generated spec stale.
func TestMoneyResponsesExposeComputedAndOverrideFields(t *testing.T) {
	want := []string{
		"discount_amount", "tax_1_amount", "tax_2_amount", "net",
		"total_override", "total_override_reason", "total_override_by_id", "total_override_at",
	}
	for _, def := range []string{"dto.PurchaseOrderResponse", "dto.WorkOrderResponse", "dto.ServiceEntryResponse"} {
		props := definitionProps(t, def)
		for _, name := range want {
			if _, ok := props[name]; !ok {
				t.Errorf("%s is missing property %q", def, name)
			}
		}
	}
}

func TestArchivableListsDocumentIncludeArchived(t *testing.T) {
	paths, _ := loadSpec(t)["paths"].(map[string]any)
	for _, res := range []string{"assets", "parts", "vendors", "service-tasks", "inspection-forms"} {
		p := "/api/v1/" + res
		op, _ := paths[p].(map[string]any)
		get, _ := op["get"].(map[string]any)
		params, _ := get["parameters"].([]any)
		has := false
		for _, raw := range params {
			if pm, ok := raw.(map[string]any); ok && pm["name"] == "include_archived" {
				has = true
			}
		}
		if !has {
			t.Errorf("%s GET does not document include_archived", p)
		}
	}
}
