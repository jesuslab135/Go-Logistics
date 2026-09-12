package dto

import (
	"bytes"
	"encoding/json"
)

// OptionalInt64 tells a JSON field that was omitted apart from one sent as
// null. A plain *int64 cannot: both decode to nil. The difference matters where
// null carries a meaning of its own. For an employee's role_id, null means
// "clear the role" and an omitted field means "leave it alone", so a client
// that never sends role_id can never strip anyone's role by accident.
type OptionalInt64 struct {
	// Set is true when the field was present in the body, null included.
	Set   bool
	Value *int64
}

// UnmarshalJSON only runs when the key is present, which is what makes Set
// reliable.
func (o *OptionalInt64) UnmarshalJSON(b []byte) error {
	o.Set = true
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		o.Value = nil
		return nil
	}
	var v int64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	o.Value = &v
	return nil
}
