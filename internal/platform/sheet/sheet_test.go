package sheet

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestXLSXRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	err := WriteXLSX(&buf, []Sheet{{Name: "Datos", Rows: [][]any{
		{"name", "unit_cost"},
		{"  Llantera Sur ", 12.5},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := Read(&buf, "plantilla.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"name", "unit_cost"}, {"Llantera Sur", "12.5"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReadPrefersTheDatosSheet(t *testing.T) {
	var buf bytes.Buffer
	err := WriteXLSX(&buf, []Sheet{
		{Name: "Instrucciones", Rows: [][]any{{"léeme"}}},
		{Name: "Datos", Rows: [][]any{{"name"}, {"Taller"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := Read(&buf, "x.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	if got[0][0] != "name" || got[1][0] != "Taller" {
		t.Fatalf("read the wrong sheet: %q", got)
	}
}

// Spanish-locale Excel saves CSV with ';' and a UTF-8 byte-order mark.
func TestReadCSVHandlesSemicolonsAndBOM(t *testing.T) {
	got, err := Read(strings.NewReader("\xef\xbb\xbfname;city\nTaller Norte;Tijuana\n"), "datos.csv")
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"name", "city"}, {"Taller Norte", "Tijuana"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReadRejectsOtherFiles(t *testing.T) {
	if _, err := Read(strings.NewReader("hello"), "notes.txt"); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("got %v, want ErrUnsupported", err)
	}
}

// Dates are read as Excel serials, not in whatever display format the cell has.
func TestXLSXDatesReadAsSerials(t *testing.T) {
	var buf bytes.Buffer
	day := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	if err := WriteXLSX(&buf, []Sheet{{Name: "Datos", Rows: [][]any{{"date"}, {day}}}}); err != nil {
		t.Fatal(err)
	}
	got, err := Read(&buf, "x.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	serial, err := strconv.ParseFloat(got[1][0], 64)
	if err != nil {
		t.Fatalf("date cell read as %q, want an Excel serial number", got[1][0])
	}
	back := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC).Add(time.Duration(serial * 24 * float64(time.Hour)))
	if !back.Equal(day) {
		t.Fatalf("serial %v is %v, want %v", serial, back, day)
	}
}

func TestWriteCSVHasBOMAndEmptyNils(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, [][]any{{"name", "notes"}, {"Taller", nil}}); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "\xef\xbb\xbfname,notes\nTaller,\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// The template generator (a later task) builds its drop-downs directly on
// WriteXLSX + DropList; this proves the validation excelize writes actually
// lands on the intended column and range, for both an inline value list and
// a range Source, and that AllowOther relaxes the error style rather than
// being silently ignored.
func TestWriteXLSXDataValidations(t *testing.T) {
	var buf bytes.Buffer
	err := WriteXLSX(&buf, []Sheet{{
		Name: "Datos",
		Rows: [][]any{
			{"name", "category", "vendor_id"},
			{"Llantera Sur", "Neumáticos", "42"},
		},
		DropLists: []DropList{
			{Column: 1, Values: []string{"Neumáticos", "Frenos", "Aceite"}},
			{Column: 2, Source: "'Catálogos'!$A$2:$A$40", AllowOther: true},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}

	f, err := excelize.OpenReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	dvs, err := f.GetDataValidations("Datos")
	if err != nil {
		t.Fatal(err)
	}
	if len(dvs) != 2 {
		t.Fatalf("got %d data validations, want 2", len(dvs))
	}

	inline := dvs[0]
	if want := fmt.Sprintf("B2:B%d", maxDataRows); inline.Sqref != want {
		t.Errorf("inline drop list Sqref = %q, want %q", inline.Sqref, want)
	}
	if inline.Type != "list" {
		t.Errorf("inline drop list Type = %q, want %q", inline.Type, "list")
	}
	if want := `"Neumáticos,Frenos,Aceite"`; inline.Formula1 != want {
		t.Errorf("inline drop list Formula1 = %q, want %q", inline.Formula1, want)
	}
	if inline.ErrorStyle != nil {
		t.Errorf("inline drop list ErrorStyle = %q, want unset (no AllowOther)", *inline.ErrorStyle)
	}

	ranged := dvs[1]
	if want := fmt.Sprintf("C2:C%d", maxDataRows); ranged.Sqref != want {
		t.Errorf("range drop list Sqref = %q, want %q", ranged.Sqref, want)
	}
	if ranged.Type != "list" {
		t.Errorf("range drop list Type = %q, want %q", ranged.Type, "list")
	}
	if want := "'Catálogos'!$A$2:$A$40"; ranged.Formula1 != want {
		t.Errorf("range drop list Formula1 = %q, want the source range %q", ranged.Formula1, want)
	}
	// AllowOther: a source-range list still gets a drop-down, but the error
	// alert is downgraded to "information" so an out-of-list value (e.g. an
	// id the list shows by name) is merely flagged, not rejected.
	if ranged.ErrorStyle == nil || *ranged.ErrorStyle != "information" {
		t.Errorf("AllowOther drop list ErrorStyle = %v, want \"information\"", ranged.ErrorStyle)
	}
	if !ranged.ShowErrorMessage {
		t.Error("AllowOther drop list should still show an (informational) error message")
	}
}

func TestColumnName(t *testing.T) {
	for i, want := range map[int]string{0: "A", 25: "Z", 26: "AA"} {
		if got := ColumnName(i); got != want {
			t.Errorf("ColumnName(%d) = %q, want %q", i, got, want)
		}
	}
}
