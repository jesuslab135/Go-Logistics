package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
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
