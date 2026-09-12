package sheet

import (
	"bytes"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
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

func TestColumnName(t *testing.T) {
	for i, want := range map[int]string{0: "A", 25: "Z", 26: "AA"} {
		if got := ColumnName(i); got != want {
			t.Errorf("ColumnName(%d) = %q, want %q", i, got, want)
		}
	}
}
