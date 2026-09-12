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

// maxCatalogEntries caps how many names one reference column contributes to
// the Catálogos sheet. Every reference column otherwise loads the whole list
// it points at, on a GET any reader may call: /work-orders/import/template
// alone pulls every asset and every employee in the company. That is unbounded
// work per request and the cheapest way to make the database do it. Reference
// drop-downs are AllowOther, so any record the list leaves out is still
// reachable by typing its id - noteCatalogCapped says so on Instrucciones
// whenever a list is actually cut.
const maxCatalogEntries = 1000

// The Notas block at the foot of the Instrucciones sheet.
const (
	// noteDecimals states the rule CellValue applies to KindDecimal cells. It
	// is the one convention a user cannot infer: the comma is the decimal
	// separator in Spanish and the thousands separator in English, and the
	// same endpoint takes files written both ways.
	noteDecimals = "Números: el último . o , de la celda es el separador decimal. " +
		"\"1.234,56\" y \"1,234.56\" son ambos 1234.56; \"1,5\" es 1.5 y \"1,234\" es 1234. " +
		"Se aceptan $ y espacios. \"1.234\" se rechaza por ambiguo: escriba 1234 o 1.234,00."
	noteCatalogCapped = "Algunas listas de la hoja Catálogos muestran solo los primeros 1000 registros. " +
		"Si el registro que busca no aparece, escriba su id en la celda."
)

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

	// Lookups shared by more than one column are loaded once and share one
	// catalogue column: work-orders points both assigned_to_id and
	// issued_by_id at the employee list, and assets points two columns at
	// vendors and two more at trailer classifications. Keyed by the lookup's
	// label, which names the underlying list one-to-one (see bulkLookups); a
	// lookup with no label falls back to its own key, which is unique by
	// construction. An empty catalogue caches an empty source, so a second
	// column sharing it skips the drop-down too.
	sources := map[string]string{}
	capped := false

	header := make([]any, len(cols))
	for i, col := range cols {
		header[i] = col.Key
		datos.Widths[i] = 22
		instr.Rows = append(instr.Rows, []any{col.Key, yesNo(col.Required), kindLabel(col), hint(col), col.Label})

		switch {
		case col.Ref != "":
			lk := refs[col.Ref]
			id := lk.Label
			if id == "" {
				id = col.Ref
			}
			source, loaded := sources[id]
			if !loaded {
				entries, err := lk.Load(ctx)
				if err != nil {
					return nil, err
				}
				names := uniqueSortedNames(entries)
				if len(names) > maxCatalogEntries {
					names = names[:maxCatalogEntries]
					capped = true
				}
				if len(names) > 0 {
					source = addCatalog(col.Key, names)
				}
				sources[id] = source
			}
			if source != "" {
				datos.DropLists = append(datos.DropLists, sheet.DropList{Column: i, Source: source, AllowOther: true})
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

	// Notes go in column A, which overflows across the empty cells beside it,
	// so the whole sentence is readable without merging anything.
	instr.Rows = append(instr.Rows, []any{}, []any{"Notas"}, []any{noteDecimals})
	if capped {
		instr.Rows = append(instr.Rows, []any{noteCatalogCapped})
	}

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
		// The old wording ("se aceptan $ y comas de miles") told Spanish users
		// their decimal comma was a thousands separator, which is exactly the
		// misreading the parser used to make. See noteDecimals.
		return "número; el último . o , es el decimal (1.234,56 = 1,234.56 = 1234.56); ver Notas"
	case col.Kind == KindString && col.Max > 0:
		return fmt.Sprintf("máximo %d caracteres", col.Max)
	}
	return ""
}
