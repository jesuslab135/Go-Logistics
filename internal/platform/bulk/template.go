package bulk

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"fleet/internal/platform/sheet"
)

// inlineLimit is how many characters of permitted values fit an inline
// drop-down; Excel caps the list at 255.
const inlineLimit = 250

// catalogSheetName is the sheet listing reference names and over-long
// permitted-value lists; used both for the sheet itself and for the range
// formulas the Datos drop-downs point at, so the two can never drift apart.
const catalogSheetName = "Catálogos"

// Template builds the workbook for one section: Datos to fill in,
// Instrucciones describing every column, and Catálogos listing the names
// reference columns accept.
func Template(ctx context.Context, imp Importer) ([]sheet.Sheet, error) {
	cols, err := imp.Columns(ctx)
	if err != nil {
		return nil, err
	}
	refs := imp.Refs()

	datos := sheet.Sheet{Name: sheet.DataSheet, FreezeHeader: true, Widths: map[int]float64{}}
	instr := sheet.Sheet{
		Name:   "Instrucciones",
		Rows:   [][]any{{"Columna", "Obligatoria", "Tipo", "Valores o formato", "Descripción"}},
		Widths: map[int]float64{0: 30, 1: 12, 2: 18, 3: 48, 4: 48},
	}
	var catHeader []any
	var catValues [][]string

	addCatalog := func(key string, values []string) string {
		ci := len(catHeader)
		catHeader = append(catHeader, key)
		catValues = append(catValues, values)
		col := sheet.ColumnName(ci)
		return fmt.Sprintf("'%s'!$%s$2:$%s$%d", catalogSheetName, col, col, len(values)+1)
	}

	header := make([]any, len(cols))
	for i, col := range cols {
		header[i] = col.Key
		datos.Widths[i] = 22
		instr.Rows = append(instr.Rows, []any{col.Key, yesNo(col.Required), kindLabel(col), hint(col), col.Label})

		switch {
		case col.Ref != "":
			entries, err := refs[col.Ref].Load(ctx)
			if err != nil {
				return nil, err
			}
			names := uniqueSortedNames(entries)
			if len(names) > 0 {
				datos.DropLists = append(datos.DropLists, sheet.DropList{Column: i, Source: addCatalog(col.Key, names), AllowOther: true})
			}
		case col.Kind == KindBool || col.CustomType == "boolean":
			datos.DropLists = append(datos.DropLists, sheet.DropList{Column: i, Values: []string{"sí", "no"}, AllowOther: true})
		case len(col.OneOf) > 0:
			if len(strings.Join(col.OneOf, ",")) <= inlineLimit {
				datos.DropLists = append(datos.DropLists, sheet.DropList{Column: i, Values: col.OneOf})
			} else {
				datos.DropLists = append(datos.DropLists, sheet.DropList{Column: i, Source: addCatalog(col.Key, col.OneOf)})
			}
		}
	}
	datos.Rows = [][]any{header}

	sheets := []sheet.Sheet{datos, instr}
	if len(catHeader) > 0 {
		sheets = append(sheets, catalogSheet(catHeader, catValues))
	}
	return sheets, nil
}

func catalogSheet(header []any, values [][]string) sheet.Sheet {
	longest := 0
	for _, v := range values {
		longest = max(longest, len(v))
	}
	rows := [][]any{header}
	for r := 0; r < longest; r++ {
		row := make([]any, len(values))
		for c, v := range values {
			if r < len(v) {
				row[c] = v[r]
			}
		}
		rows = append(rows, row)
	}
	widths := map[int]float64{}
	for c := range header {
		widths[c] = 28
	}
	return sheet.Sheet{Name: catalogSheetName, Rows: rows, Widths: widths, FreezeHeader: true}
}

func uniqueSortedNames(entries []Entry) []string {
	seen := map[string]bool{}
	var names []string
	for _, e := range entries {
		if e.Name != "" && !seen[e.Name] {
			seen[e.Name] = true
			names = append(names, e.Name)
		}
	}
	sort.Strings(names)
	return names
}

func yesNo(b bool) string {
	if b {
		return "Sí"
	}
	return "No"
}

func kindLabel(col Column) string {
	switch {
	case col.Ref != "":
		return "referencia"
	case col.CustomType == "date":
		return "fecha"
	case col.CustomType == "number":
		return "número"
	case col.CustomType == "boolean":
		return "sí / no"
	}
	switch col.Kind {
	case KindInt:
		return "número entero"
	case KindDecimal, KindFloat:
		return "número"
	case KindBool:
		return "sí / no"
	case KindTime:
		return "fecha"
	case KindJSON:
		return "JSON"
	case KindList:
		return "lista"
	}
	return "texto"
}

func hint(col Column) string {
	switch {
	case col.Ref != "" && col.Label == "":
		// Label is a manual annotation an importer author adds alongside Ref
		// (SchemaOf never sets either), so a Ref without a Label is
		// reachable. Fall back to a gap-free phrasing rather than "Nombre
		// de  (ver..." with the missing name leaving a double space, and
		// rather than interpolating col.Key, whose snake_case API field name
		// would read badly inside Spanish prose.
		return "Nombre (ver hoja Catálogos) o su id"
	case col.Ref != "":
		return "Nombre de " + col.Label + " (ver hoja Catálogos) o su id"
	case len(col.OneOf) > 0:
		return strings.Join(col.OneOf, ", ")
	case col.Kind == KindTime || col.CustomType == "date":
		return "AAAA-MM-DD o DD/MM/AAAA"
	case col.Kind == KindBool || col.CustomType == "boolean":
		return "sí o no"
	case col.Kind == KindList:
		return "valores separados por comas"
	case col.Kind == KindDecimal:
		return "número; se aceptan $ y comas de miles"
	case col.Kind == KindString && col.Max > 0:
		return fmt.Sprintf("máximo %d caracteres", col.Max)
	}
	return ""
}
