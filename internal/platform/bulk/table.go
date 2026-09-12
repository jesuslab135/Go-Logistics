package bulk

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Table flattens JSON objects into spreadsheet rows. Columns are the keys in
// first-seen order; nested objects become dotted columns; arrays are written
// as JSON text.
type Table struct {
	columns []string
	index   map[string]int
	rows    []map[string]any
}

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
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return fmt.Errorf("export row is not a JSON object")
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key := prefix + keyTok.(string)
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return err
		}
		if bytes.HasPrefix(bytes.TrimSpace(val), []byte("{")) {
			if err := t.flatten(key+".", val, row); err != nil {
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
func (t *Table) Records() [][]any {
	header := make([]any, len(t.columns))
	for i, c := range t.columns {
		header[i] = c
	}
	out := [][]any{header}
	for _, row := range t.rows {
		rec := make([]any, len(t.columns))
		for i, c := range t.columns {
			rec[i] = row[c]
		}
		out = append(out, rec)
	}
	return out
}
