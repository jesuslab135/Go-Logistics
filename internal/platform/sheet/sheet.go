// Package sheet reads uploaded spreadsheets into rows of strings and writes
// workbooks and CSV files. It knows nothing about what the rows mean.
package sheet

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ErrUnsupported is returned for a file that is neither .xlsx nor .csv.
var ErrUnsupported = errors.New("unsupported file type: upload an .xlsx or .csv file")

// DataSheet is the sheet Read prefers, and the one templates put data in.
const DataSheet = "Datos"

// maxDataRows bounds the range drop-down lists apply to. It is deliberately
// one more than IMPORT_MAX_ROWS' default of 5000 data rows (the header is row
// 1); see bulkEnvInt in internal/http/handler/bulk.go, which names this
// constant in turn. Raise that environment variable without raising this and
// every row past 5001 silently loses its drop-downs.
const maxDataRows = 5001

// Read returns every row of the Datos sheet (or the first sheet) of an .xlsx,
// or of a .csv, with each cell trimmed. XLSX cells are read raw, so a date
// comes back as its Excel serial number rather than in whatever display format
// the cell happens to have.
//
// A returned row may be shorter than the header row: excelize (and CSV) trim
// each row to its own last non-empty cell, so trailing blank cells are simply
// absent rather than empty strings. A fully blank interior row can come back
// as a nil slice. Callers must range over a row rather than index it by
// header column position.
func Read(r io.Reader, filename string) ([][]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch {
	case ext == ".xlsx" || bytes.HasPrefix(data, []byte("PK\x03\x04")):
		return readXLSX(data)
	case ext == ".csv":
		return readCSV(data)
	default:
		return nil, ErrUnsupported
	}
}

func readXLSX(data []byte) ([][]string, error) {
	// OpenReader can return a non-nil *File alongside a non-nil error (for
	// example when the zip and shared strings parse but a later part such as
	// styles or the calc chain fails on a corrupt upload). Register the
	// close before checking the error so any temp files excelize already
	// spooled to disk are always cleaned up.
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("read xlsx: %w", err)
	}

	name := f.GetSheetName(0)
	if idx, err := f.GetSheetIndex(DataSheet); err == nil && idx >= 0 {
		name = DataSheet
	}
	rows, err := f.GetRows(name, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("read sheet %q: %w", name, err)
	}
	return trimAll(rows), nil
}

func readCSV(data []byte) ([][]string, error) {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	rd := csv.NewReader(bytes.NewReader(data))
	rd.FieldsPerRecord = -1
	if header, _, _ := bytes.Cut(data, []byte("\n")); bytes.Count(header, []byte(";")) > bytes.Count(header, []byte(",")) {
		rd.Comma = ';'
	}
	rows, err := rd.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	return trimAll(rows), nil
}

func trimAll(rows [][]string) [][]string {
	for _, row := range rows {
		for i := range row {
			row[i] = strings.TrimSpace(row[i])
		}
	}
	return rows
}

// DropList puts an in-cell drop-down on one column of a sheet, from row 2 down.
type DropList struct {
	Column int      // zero-based column
	Values []string // inline values; keep short (Excel caps them at 255 characters)
	Source string   // or a range on another sheet, e.g. "'Catálogos'!$A$2:$A$40"
	// AllowOther lets a user type a value outside the list (for example an id
	// where the list shows names); Excel then only shows a notice.
	AllowOther bool
}

// Sheet is one worksheet to write.
type Sheet struct {
	Name         string
	Rows         [][]any
	DropLists    []DropList
	Widths       map[int]float64
	FreezeHeader bool
}

// WriteXLSX writes the sheets, in order, as one workbook. Row 1 of each sheet is bold.
func WriteXLSX(w io.Writer, sheets []Sheet) error {
	f := excelize.NewFile()
	defer f.Close()

	bold, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return err
	}
	for i, s := range sheets {
		if i == 0 {
			if err := f.SetSheetName("Sheet1", s.Name); err != nil {
				return err
			}
		} else if _, err := f.NewSheet(s.Name); err != nil {
			return err
		}
		for r, row := range s.Rows {
			cell, err := excelize.CoordinatesToCellName(1, r+1)
			if err != nil {
				return err
			}
			if err := f.SetSheetRow(s.Name, cell, &row); err != nil {
				return err
			}
		}
		if len(s.Rows) > 0 {
			if err := f.SetRowStyle(s.Name, 1, 1, bold); err != nil {
				return err
			}
		}
		for col, width := range s.Widths {
			name := ColumnName(col)
			if err := f.SetColWidth(s.Name, name, name, width); err != nil {
				return err
			}
		}
		if s.FreezeHeader {
			if err := f.SetPanes(s.Name, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
				return err
			}
		}
		for _, dl := range s.DropLists {
			if err := addDropList(f, s.Name, dl); err != nil {
				return err
			}
		}
	}
	f.SetActiveSheet(0)
	return f.Write(w)
}

// WriteXLSXStream writes ONE sheet through excelize's StreamWriter, which
// serialises each row as it is handed over instead of building the whole
// worksheet as an object graph first. WriteXLSX holds roughly a kilobyte per
// cell until it serialises, so a 20000-row wide export is on the order of
// 800k cells and hundreds of megabytes - enough to have the API container
// killed rather than merely answer slowly.
//
// This is additive, not a replacement: the StreamWriter writes a single sheet
// and cannot add data validation, which the three-sheet templates and their
// drop-downs both need. An export is one sheet with no validation, so only the
// export path uses this and templates keep WriteXLSX unchanged.
func WriteXLSXStream(w io.Writer, name string, rows [][]any) error {
	f := excelize.NewFile()
	defer f.Close()

	if err := f.SetSheetName("Sheet1", name); err != nil {
		return err
	}
	bold, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return err
	}
	sw, err := f.NewStreamWriter(name)
	if err != nil {
		return err
	}
	if err := sw.SetPanes(&excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
		return err
	}

	for r, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, r+1)
		if err != nil {
			return err
		}
		values := row
		if r == 0 {
			// Row 1 is bold, matching WriteXLSX. The StreamWriter takes the
			// style per cell rather than per row.
			values = make([]any, len(row))
			for i, v := range row {
				values[i] = excelize.Cell{StyleID: bold, Value: v}
			}
		}
		if err := sw.SetRow(cell, values); err != nil {
			return err
		}
	}

	if err := sw.Flush(); err != nil {
		return err
	}
	return f.Write(w)
}

func addDropList(f *excelize.File, sheetName string, dl DropList) error {
	dv := excelize.NewDataValidation(true)
	col := ColumnName(dl.Column)
	dv.Sqref = fmt.Sprintf("%s2:%s%d", col, col, maxDataRows)
	if dl.Source != "" {
		dv.SetSqrefDropList(dl.Source)
	} else if err := dv.SetDropList(dl.Values); err != nil {
		return err
	}
	if dl.AllowOther {
		dv.SetError(excelize.DataValidationErrorStyleInformation, "", "")
	}
	return f.AddDataValidation(sheetName, dv)
}

// WriteCSV writes rows as UTF-8 CSV with a byte-order mark, so Excel shows
// accents correctly. A nil value is an empty cell, and a cell a spreadsheet
// would otherwise execute is defused - see csvCell.
func WriteCSV(w io.Writer, rows [][]any) error {
	if _, err := io.WriteString(w, "\xef\xbb\xbf"); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	for _, row := range rows {
		rec := make([]string, len(row))
		for i, v := range row {
			if v != nil {
				rec[i] = csvCell(v)
			}
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// csvCell renders one value, defusing CSV formula injection. Excel and
// LibreOffice execute a cell whose text starts with "=", "+", "-", "@", a tab
// or a carriage return, so an exported vendor name or description beginning
// with one of those runs as a formula on whatever machine opens the file -
// reachable by anyone who can type a name into this API. OWASP's guidance is
// to prefix the cell with a single quote, which those programs strip back off
// when they display it. The .xlsx path needs none of this: it writes inline
// strings, which are never evaluated.
//
// A value that is simply a number is left alone. "-1234.50" starts with "-"
// but a spreadsheet reads it as the negative number it is and never as a
// formula, and quoting it would turn every negative amount in every export
// into text.
func csvCell(v any) string {
	s := fmt.Sprint(v)
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
	default:
		return s
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return s
	}
	return "'" + s
}

// ColumnName is the spreadsheet letter of a zero-based column: 0 → "A".
func ColumnName(i int) string {
	name, _ := excelize.ColumnNumberToName(i + 1)
	return name
}
