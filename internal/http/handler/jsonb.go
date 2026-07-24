package handler

import "encoding/json"

// jsonbOrDefault returns the raw JSON bytes for a NOT NULL jsonb column,
// substituting the column default (e.g. "[]" or "{}") when the client omits it.
func jsonbOrDefault(raw json.RawMessage, def string) []byte {
	if len(raw) == 0 {
		return []byte(def)
	}
	return []byte(raw)
}
