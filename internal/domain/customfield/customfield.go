// Package customfield validates a custom_fields document against the
// definitions a company has declared for a resource.
//
// The jsonb columns were a free-for-all: eight resources carried one, nothing
// described what belonged in it, and no value was ever checked. The alternative
// to this — a raw JSON editor in front of a business user — is the outcome the
// original audit advised against, and it is not really an alternative: it moves
// the problem to whoever has to read the result.
package customfield

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// The field types a definition can declare.
//
// Multi-select is deliberately absent. It changes the stored shape from a
// scalar to an array, which doubles the filtering work and makes every consumer
// handle two shapes per key; it can be added when something actually needs it.
const (
	TypeText    = "text"
	TypeNumber  = "number"
	TypeDate    = "date"
	TypeBoolean = "boolean"
	TypeSelect  = "select"
)

// Types lists the vocabulary, for validation and error messages.
var Types = []string{TypeText, TypeNumber, TypeDate, TypeBoolean, TypeSelect}

// ValidType reports whether a definition may declare this type.
func ValidType(t string) bool {
	return slices.Contains(Types, t)
}

// Definition is one declared field.
type Definition struct {
	Key      string
	Label    string
	Type     string
	Required bool
	// Options are the permitted values for a select, and are ignored otherwise.
	Options []string
}

// Validate checks a document against the definitions for its resource and
// returns one message per offending key, in the shape the API's 422 body uses.
//
// Keys with no definition are preserved, not rejected. A company that has been
// writing undeclared keys keeps them: they are that tenant's data, and refusing
// them on the next write would make previously-saved records uneditable through
// a form that no longer shows the fields holding them. Declaring a field is how
// you start validating it, not a wall that goes up around everything else.
func Validate(raw json.RawMessage, defs []Definition) map[string]string {
	if len(defs) == 0 && len(raw) == 0 {
		return nil
	}

	doc := map[string]json.RawMessage{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &doc); err != nil {
			return map[string]string{"custom_fields": "must be a JSON object"}
		}
	}

	errs := map[string]string{}
	for _, def := range defs {
		value, present := doc[def.Key]
		if !present || isNull(value) {
			if def.Required {
				errs[field(def.Key)] = fmt.Sprintf("%s is required", def.Label)
			}
			continue
		}
		if msg := checkValue(def, value); msg != "" {
			errs[field(def.Key)] = msg
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// field names the offending key the way the rest of the API names a field, so a
// client can attach the message to the input that produced it.
func field(key string) string { return "custom_fields." + key }

func isNull(v json.RawMessage) bool {
	return len(v) == 0 || string(v) == "null"
}

func checkValue(def Definition, value json.RawMessage) string {
	switch def.Type {
	case TypeText:
		var s string
		if json.Unmarshal(value, &s) != nil {
			return "must be text"
		}
	case TypeNumber:
		var f float64
		if json.Unmarshal(value, &f) != nil {
			return "must be a number"
		}
	case TypeBoolean:
		var b bool
		if json.Unmarshal(value, &b) != nil {
			return "must be true or false"
		}
	case TypeDate:
		var s string
		if json.Unmarshal(value, &s) != nil {
			return "must be a date as a string"
		}
		// Date only, not a timestamp: these are business dates — a service due
		// date, a certification expiry — and storing a time nobody entered
		// invites timezone bugs in whatever renders it back.
		if _, err := time.Parse(time.DateOnly, s); err != nil {
			return "must be a date in YYYY-MM-DD form"
		}
	case TypeSelect:
		var s string
		if json.Unmarshal(value, &s) != nil {
			return "must be one of the permitted values"
		}
		if !slices.Contains(def.Options, s) {
			return "must be one of: " + strings.Join(def.Options, ", ")
		}
	default:
		// A type this build does not know cannot be checked, so it is left
		// alone rather than rejected: refusing data because the server is older
		// than the definition would be the wrong way round.
		return ""
	}
	return ""
}

// FilterValue turns a query-string value into the JSON scalar a containment
// query needs, so custom_fields.cost_centre=A12 matches the string "A12" and
// custom_fields.axles=3 matches the number 3 rather than the string "3".
func FilterValue(def Definition, raw string) (json.RawMessage, error) {
	switch def.Type {
	case TypeNumber:
		if _, err := strconv.ParseFloat(raw, 64); err != nil {
			return nil, fmt.Errorf("must be a number")
		}
		return json.RawMessage(raw), nil
	case TypeBoolean:
		if raw != "true" && raw != "false" {
			return nil, fmt.Errorf("must be true or false")
		}
		return json.RawMessage(raw), nil
	default:
		quoted, err := json.Marshal(raw)
		if err != nil {
			return nil, err
		}
		return quoted, nil
	}
}
