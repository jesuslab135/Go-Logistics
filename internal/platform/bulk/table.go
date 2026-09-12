package bulk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Table flattens JSON objects into spreadsheet rows. Columns are the keys in
// first-seen order; nested objects become dotted columns; arrays are written
// as JSON text.
//
// The top-level "custom_fields" object is a special case: the import engine
// keys a custom field's column as "cf.<key>" (see engine.go's Flat.Columns),
// so a list response's nested "custom_fields":{"<key>":...} is flattened to
// "cf.<key>" rather than "custom_fields.<key>" — otherwise a file exported
// from any section with custom fields would come back with headers the
// importer does not recognize, and those columns would be silently dropped
// on re-upload.
type Table struct {
	columns []string
	index   map[string]int
	rows    []map[string]any
	// objectKeys records every natural (pre-remap) dotted key ever seen
	// holding a JSON object, so Records can drop a same-named bare column
	// that appeared on a different row where the value was null instead of
	// an object (see Records).
	objectKeys map[string]bool
}

// customFieldsKey is the JSON field a list response nests custom fields
// under; customFieldsPrefix is the header prefix the import engine expects
// for them instead.
const (
	customFieldsKey    = "custom_fields"
	customFieldsPrefix = "cf."
)

// Add appends one JSON object as a row.
func (t *Table) Add(raw json.RawMessage) error {
	if t.index == nil {
		t.index = map[string]int{}
	}
	row := map[string]any{}
	if err := t.flatten("", raw, row); err != nil {
		return err
	}
	t.rows = append(t.rows, row)
	return nil
}

func (t *Table) flatten(prefix string, raw json.RawMessage, row map[string]any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("export row is not a JSON object: %w", err)
	}
	if tok != json.Delim('{') {
		return errors.New("export row is not a JSON object")
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		name := keyTok.(string)
		key := prefix + name
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return err
		}
		if bytes.HasPrefix(bytes.TrimSpace(val), []byte("{")) {
			if t.objectKeys == nil {
				t.objectKeys = map[string]bool{}
			}
			t.objectKeys[key] = true
			childPrefix := key + "."
			if prefix == "" && name == customFieldsKey {
				childPrefix = customFieldsPrefix
			}
			if err := t.flatten(childPrefix, val, row); err != nil {
				return err
			}
			continue
		}
		if _, seen := t.index[key]; !seen {
			t.index[key] = len(t.columns)
			t.columns = append(t.columns, key)
		}
		row[key] = scalar(val)
	}
	return nil
}

func scalar(val json.RawMessage) any {
	trimmed := bytes.TrimSpace(val)
	switch {
	case bytes.Equal(trimmed, []byte("null")):
		return nil
	case bytes.HasPrefix(trimmed, []byte("[")):
		var compact bytes.Buffer
		if json.Compact(&compact, trimmed) == nil {
			return compact.String()
		}
		return string(trimmed)
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return string(trimmed)
	}
	if n, ok := v.(json.Number); ok {
		if i, err := n.Int64(); err == nil {
			return i
		}
		if f, err := n.Float64(); err == nil {
			return f
		}
	}
	return v
}

// Records returns the header row followed by one row per object.
//
// A column whose name was also seen holding a JSON object on some other row
// (a nested field that is null on one row and populated on another) is
// dropped here rather than kept as a bare column: which of the two forms
// showed up first would otherwise depend on row order, so two exports of the
// same section could disagree on their own header row.
func (t *Table) Records() [][]any {
	cols := make([]string, 0, len(t.columns))
	for _, c := range t.columns {
		if t.objectKeys[c] {
			continue
		}
		cols = append(cols, c)
	}

	header := make([]any, len(cols))
	for i, c := range cols {
		header[i] = c
	}
	out := [][]any{header}
	for _, row := range t.rows {
		rec := make([]any, len(cols))
		for i, c := range cols {
			rec[i] = row[c]
		}
		out = append(out, rec)
	}
	return out
}
