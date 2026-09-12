package bulk

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"testing"

	"github.com/xuri/excelize/v2"

	"fleet/internal/platform/sheet"
)

// templateFake is a stand-in Importer for these tests. engine_test.go in this
// same package already declares a type named fakeImporter (with different
// fields and a *pointer* receiver), so this one is named distinctly to avoid
// a duplicate declaration.
type templateFake struct {
	cols []Column
	refs map[string]Lookup
}

func (f templateFake) Resource() string                          { return "parts" }
func (f templateFake) Columns(context.Context) ([]Column, error) { return f.cols, nil }
func (f templateFake) Refs() map[string]Lookup                   { return f.refs }
func (f templateFake) Prepare(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
	return nil, nil
}

func TestTemplateSheets(t *testing.T) {
	imp := templateFake{
		cols: []Column{
			{Key: "part_number", Kind: KindString, Required: true, Max: 100},
			{Key: "inventory_item", Kind: KindBool},
			{Key: "part_category_id", Kind: KindInt, Ref: "part_category_id", Label: "categoría"},
		},
		refs: map[string]Lookup{"part_category_id": {Label: "categoría", Load: func(context.Context) ([]Entry, error) {
			return []Entry{{ID: 2, Name: "Frenos"}, {ID: 1, Name: "Filtros"}}, nil
		}}},
	}
	sheets, err := Template(context.Background(), imp)
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets) != 3 || sheets[0].Name != "Datos" || sheets[1].Name != "Instrucciones" || sheets[2].Name != "Catálogos" {
		t.Fatalf("sheets: %+v", sheets)
	}
	header := sheets[0].Rows[0]
	if len(header) != 3 || header[0] != "part_number" || header[2] != "part_category_id" {
		t.Fatalf("header: %v", header)
	}
	if got := sheets[1].Rows[1]; got[0] != "part_number" || got[1] != "Sí" {
		t.Fatalf("instructions for part_number: %v", got)
	}
	cat := sheets[2].Rows
	if cat[0][0] != "part_category_id" || cat[1][0] != "Filtros" || cat[2][0] != "Frenos" {
		t.Fatalf("catalog (sorted names): %v", cat)
	}
	var sawRef, sawBool bool
	for _, dl := range sheets[0].DropLists {
		switch dl.Column {
		case 2:
			sawRef = dl.Source == "'Catálogos'!$A$2:$A$3" && dl.AllowOther
		case 1:
			sawBool = len(dl.Values) == 2
		}
	}
	if !sawRef || !sawBool {
		t.Fatalf("drop lists: %+v", sheets[0].DropLists)
	}
}

// TestTemplateInstructionsHeaderIsSpanish pins the exact Spanish column
// headings a user sees on the Instrucciones sheet.
func TestTemplateInstructionsHeaderIsSpanish(t *testing.T) {
	imp := templateFake{cols: []Column{{Key: "name", Kind: KindString}}}
	sheets, err := Template(context.Background(), imp)
	if err != nil {
		t.Fatal(err)
	}
	want := []any{"Columna", "Obligatoria", "Tipo", "Valores o formato", "Descripción"}
	got := sheets[1].Rows[0]
	if len(got) != len(want) {
		t.Fatalf("instructions header: %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("instructions header[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

// TestTemplateRefColumnWithNoEntriesStaysUsable covers a reference lookup
// that returns no entries (an empty catalogue, e.g. a brand-new company).
// The column must not get a broken or empty drop-down, and since it is the
// only source of catalogue data, no Catálogos sheet should be produced at all.
func TestTemplateRefColumnWithNoEntriesStaysUsable(t *testing.T) {
	imp := templateFake{
		cols: []Column{{Key: "vendor_id", Kind: KindInt, Ref: "vendor_id", Label: "proveedor"}},
		refs: map[string]Lookup{"vendor_id": {Label: "proveedor", Load: func(context.Context) ([]Entry, error) {
			return nil, nil
		}}},
	}
	sheets, err := Template(context.Background(), imp)
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets) != 2 {
		t.Fatalf("sheets = %+v; want just Datos and Instrucciones, no Catálogos for an empty lookup", sheets)
	}
	for _, dl := range sheets[0].DropLists {
		if dl.Column == 0 {
			t.Fatalf("drop list on the empty-lookup column: %+v; want none", dl)
		}
	}
}

// TestTemplateColumnWithEmptyLabelRendersEmptyDescription documents the
// deliberate choice for a Column whose Label is unset (legitimate per Task
// 3/5: Label is filled in by an importer implementation, not by SchemaOf):
// the Descripción cell is simply blank rather than falling back to the key
// or a placeholder.
func TestTemplateColumnWithEmptyLabelRendersEmptyDescription(t *testing.T) {
	imp := templateFake{cols: []Column{{Key: "part_number", Kind: KindString}}}
	sheets, err := Template(context.Background(), imp)
	if err != nil {
		t.Fatal(err)
	}
	row := sheets[1].Rows[1]
	if row[4] != "" {
		t.Fatalf("Descripción for an unlabeled column = %v, want empty", row[4])
	}
}

// TestTemplateRoundTripsThroughSheetRead is the behavior the whole feature
// depends on: a template written to an actual workbook and read back through
// sheet.Read must produce a header row whose cells exactly match the
// column keys the import engine matches against (Run builds its byKey map
// from Column.Key). A header sheet.Read can't recognise is a silent dead end
// for the user filling in the template.
func TestTemplateRoundTripsThroughSheetRead(t *testing.T) {
	imp := templateFake{
		cols: []Column{
			{Key: "part_number", Kind: KindString, Required: true},
			{Key: "inventory_item", Kind: KindBool},
			{Key: "part_category_id", Kind: KindInt, Ref: "part_category_id", Label: "categoría"},
		},
		refs: map[string]Lookup{"part_category_id": {Label: "categoría", Load: func(context.Context) ([]Entry, error) {
			return []Entry{{ID: 1, Name: "Filtros"}}, nil
		}}},
	}
	sheets, err := Template(context.Background(), imp)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := sheet.WriteXLSX(&buf, sheets); err != nil {
		t.Fatal(err)
	}

	rows, err := sheet.Read(&buf, "plantilla.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("Read returned no rows")
	}
	wantHeader := []string{"part_number", "inventory_item", "part_category_id"}
	if len(rows[0]) != len(wantHeader) {
		t.Fatalf("header = %v, want %v", rows[0], wantHeader)
	}
	for i, want := range wantHeader {
		if rows[0][i] != want {
			t.Fatalf("header[%d] = %q, want %q", i, rows[0][i], want)
		}
	}
}

// sqrefLastRow extracts the trailing row number of a data validation Sqref
// such as "C2:C5001".
var sqrefLastRow = regexp.MustCompile(`\d+$`)

// TestTemplateDropListRangeCoversDataRows proves the drop-down actually
// applies to the rows a user would fill in, not just the header row: the
// validation's range must extend deep into the sheet, not stop at row 2.
func TestTemplateDropListRangeCoversDataRows(t *testing.T) {
	imp := templateFake{
		cols: []Column{{Key: "part_category_id", Kind: KindInt, Ref: "part_category_id", Label: "categoría"}},
		refs: map[string]Lookup{"part_category_id": {Label: "categoría", Load: func(context.Context) ([]Entry, error) {
			return []Entry{{ID: 1, Name: "Filtros"}, {ID: 2, Name: "Frenos"}}, nil
		}}},
	}
	sheets, err := Template(context.Background(), imp)
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets[0].DropLists) != 1 {
		t.Fatalf("drop lists = %+v, want exactly 1", sheets[0].DropLists)
	}

	var buf bytes.Buffer
	if err := sheet.WriteXLSX(&buf, sheets); err != nil {
		t.Fatal(err)
	}

	// sheet.Read only returns cell values, so open the same bytes with
	// excelize directly to inspect the validation's Sqref (the range it
	// actually applies to).
	f, err := excelize.OpenReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dvs, err := f.GetDataValidations(sheet.DataSheet)
	if err != nil {
		t.Fatal(err)
	}
	if len(dvs) != 1 {
		t.Fatalf("data validations = %+v, want exactly 1", dvs)
	}
	last := sqrefLastRow.FindString(dvs[0].Sqref)
	n, err := strconv.Atoi(last)
	if err != nil {
		t.Fatalf("could not parse trailing row from Sqref %q: %v", dvs[0].Sqref, err)
	}
	if n < 1000 {
		t.Fatalf("drop-down range ends at row %d, want it to reach deep into the data rows (>= 1000)", n)
	}
}
