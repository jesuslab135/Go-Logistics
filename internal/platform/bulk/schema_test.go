package bulk

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

type sampleNested struct {
	EngineSerial string `json:"engine_serial" binding:"omitempty,max=100"`
}

type sampleRequest struct {
	Name         string           `json:"name" binding:"required,max=100"`
	Kind         string           `json:"kind" binding:"omitempty,oneof=truck trailer"`
	Count        int32            `json:"count"`
	Cost         *decimal.Decimal `json:"cost"`
	Active       bool             `json:"active"`
	BoughtAt     *time.Time       `json:"bought_at"`
	Labels       json.RawMessage  `json:"labels"`
	Extra        json.RawMessage  `json:"extra"`
	Tags         []string         `json:"tags"`
	CustomFields json.RawMessage  `json:"custom_fields"`
	Secret       string           `json:"-"`
	Skipped      string           `json:"skipped"`
	Vehicle      *sampleNested    `json:"vehicle,omitempty"`
}

func TestSchemaOf(t *testing.T) {
	got := SchemaOf(reflect.TypeOf(sampleRequest{}), "skipped")
	want := []Column{
		{Key: "name", Kind: KindString, Required: true, Max: 100},
		{Key: "kind", Kind: KindString, OneOf: []string{"truck", "trailer"}},
		{Key: "count", Kind: KindInt},
		{Key: "cost", Kind: KindDecimal},
		{Key: "active", Kind: KindBool},
		{Key: "bought_at", Kind: KindTime},
		{Key: "labels", Kind: KindList},
		{Key: "extra", Kind: KindJSON},
		{Key: "tags", Kind: KindList},
		{Key: "vehicle.engine_serial", Kind: KindString, Max: 100},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SchemaOf:\n got %+v\nwant %+v", got, want)
	}
}

// TestSchemaOfAcceptsPointerType covers a type passed as a pointer instead of
// a value, which is how a caller working from a var of the request type
// (rather than a literal) would naturally spell reflect.TypeOf.
func TestSchemaOfAcceptsPointerType(t *testing.T) {
	got := SchemaOf(reflect.TypeOf(&sampleRequest{}), "skipped")
	want := SchemaOf(reflect.TypeOf(sampleRequest{}), "skipped")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SchemaOf(pointer type):\n got %+v\nwant %+v", got, want)
	}
}

// unexportedFieldRequest stands in for the shape every real create request
// has at least implicitly: an unexported field must never become a column,
// since a bulk row could never populate it via reflection anyway.
type unexportedFieldRequest struct {
	Name string `json:"name"`
	// secret exists only so SchemaOf has an unexported field to skip.
	secret string
}

func TestSchemaOfSkipsUnexportedFields(t *testing.T) {
	got := SchemaOf(reflect.TypeOf(unexportedFieldRequest{}))
	want := []Column{{Key: "name", Kind: KindString}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SchemaOf:\n got %+v\nwant %+v", got, want)
	}
}

// embeddedFieldRequest documents current behavior for an anonymous field: it
// is treated like any other field and needs its own JSON name to produce a
// column, matching the brief's rule ("one column per exported field with a
// JSON name"). No real create request in this codebase embeds a struct today.
type EmbeddedPart struct {
	Code string `json:"code"`
}

type embeddedFieldRequest struct {
	Name         string `json:"name"`
	EmbeddedPart `json:"part"`
}

func TestSchemaOfTreatsAnonymousFieldAsAnOrdinaryField(t *testing.T) {
	got := SchemaOf(reflect.TypeOf(embeddedFieldRequest{}))
	want := []Column{
		{Key: "name", Kind: KindString},
		{Key: "part", Kind: KindJSON},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SchemaOf:\n got %+v\nwant %+v", got, want)
	}
}
