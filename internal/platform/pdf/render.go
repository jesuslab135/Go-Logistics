package pdf

import (
	"fmt"
	"net/http"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"fleet/internal/platform/reports"
)

// emptyNotice is printed instead of the tables when a period has no data.
const emptyNotice = "Sin registros en el periodo"

// overrideNote explains the asterisk on a maintenance row.
const overrideNote = "* El total de al menos un trabajo fue ajustado manualmente, por lo que puede no coincidir con refacciones más mano de obra."

var (
	headStyle  = props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Left}
	headRight  = props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Right}
	cellStyle  = props.Text{Size: 8, Align: align.Left}
	cellRight  = props.Text{Size: 8, Align: align.Right}
	titleStyle = props.Text{Size: 14, Style: fontstyle.Bold, Align: align.Left}
	subStyle   = props.Text{Size: 9, Align: align.Left, Top: 7}
	metaStyle  = props.Text{Size: 8, Align: align.Left, Top: 12}
	sectionTop = props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Left, Top: 3}
)

// column is one column of a table: its heading, its width out of 12, and
// whether its values are numbers (right-aligned).
type column struct {
	title   string
	width   int
	numeric bool
}

// imageExtension reports the format of an image maroto can embed. It embeds
// PNG and JPEG only; a WebP or GIF logo is a normal upload here, so an
// unsupported format means "no logo", not an error.
func imageExtension(data []byte) (extension.Type, bool) {
	switch http.DetectContentType(data) {
	case "image/png":
		return extension.Png, true
	case "image/jpeg":
		return extension.Jpg, true
	default:
		return "", false
	}
}

func newDocument(h reports.Header, logo []byte) (core.Maroto, error) {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.Letter).
		WithOrientation(orientation.Horizontal).
		WithLeftMargin(12).
		WithTopMargin(12).
		WithRightMargin(12).
		WithBottomMargin(12).
		WithPageNumber(props.PageNumber{
			Pattern: "Página {current} de {total}",
			Place:   props.RightBottom,
			Size:    8,
		}).
		Build()
	m := maroto.New(cfg)

	loc := h.Period.Start.Location()
	header := row.New(22)
	if ext, ok := imageExtension(logo); ok {
		header.Add(image.NewFromBytesCol(2, logo, ext, props.Rect{Center: true, Percent: 90}))
	} else {
		header.Add(col.New(2))
	}
	header.Add(col.New(10).Add(
		text.New(h.Kind.Title(), titleStyle),
		text.New(h.Company.Name, subStyle),
		text.New(fmt.Sprintf("Periodo: %s a %s    Generado: %s",
			Date(h.Period.Start, loc), Date(h.Period.LastDay(), loc), Date(h.GeneratedAt, loc)), metaStyle),
	))
	if err := m.RegisterHeader(header, row.New(1).Add(line.NewCol(12)), row.New(3)); err != nil {
		return nil, fmt.Errorf("pdf: register header: %w", err)
	}
	return m, nil
}

func section(m core.Maroto, title string) {
	m.AddRow(9, text.NewCol(12, title, sectionTop))
}

func table(m core.Maroto, columns []column, rows [][]string) {
	heads := make([]core.Col, 0, len(columns))
	for _, c := range columns {
		style := headStyle
		if c.numeric {
			style = headRight
		}
		heads = append(heads, text.NewCol(c.width, c.title, style))
	}
	m.AddRow(6, heads...)
	m.AddRow(1, line.NewCol(12))

	for _, values := range rows {
		cells := make([]core.Col, 0, len(columns))
		for i, c := range columns {
			style := cellStyle
			if c.numeric {
				style = cellRight
			}
			cells = append(cells, text.NewCol(c.width, values[i], style))
		}
		// An auto row grows with its tallest cell, so a long asset name wraps
		// instead of running over the next column.
		m.AddAutoRow(cells...)
	}
	m.AddRow(3)
}

func finish(m core.Maroto) ([]byte, error) {
	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("pdf: generate: %w", err)
	}
	return doc.GetBytes(), nil
}

func comparisonRows(label string, c reports.Comparison, format func(v reports.Comparison) [3]string) []string {
	v := format(c)
	return []string{label, v[0], v[1], v[2], Percent(c.Percent)}
}

var comparisonColumns = []column{
	{"Concepto", 4, false},
	{"Periodo", 2, true},
	{"Periodo anterior", 2, true},
	{"Diferencia", 2, true},
	{"Cambio", 2, true},
}

// FuelWeekly renders the weekly fuel report. logo may be nil.
func FuelWeekly(r reports.FuelWeeklyReport, logo []byte) ([]byte, error) {
	m, err := newDocument(r.Header, logo)
	if err != nil {
		return nil, err
	}
	if r.Empty() {
		m.AddRow(12, text.NewCol(12, emptyNotice, props.Text{Size: 11, Align: align.Center, Top: 4}))
		return finish(m)
	}
	currency := r.Company.Currency

	section(m, "Resumen")
	summary := make([][]string, 0, len(r.Volumes)+1)
	summary = append(summary, comparisonRows("Costo total", r.Cost, func(c reports.Comparison) [3]string {
		return [3]string{Money(c.Current, currency), Money(c.Previous, currency), Money(c.Difference, currency)}
	}))
	for _, v := range r.Volumes {
		label := "Volumen " + v.FuelType + " (" + Unit(v.Unit) + ")"
		summary = append(summary, comparisonRows(label, v.Change, func(c reports.Comparison) [3]string {
			return [3]string{Number(c.Current, 2), Number(c.Previous, 2), Number(c.Difference, 2)}
		}))
	}
	table(m, comparisonColumns, summary)

	totals := make([][]string, 0, len(r.Distances)+1)
	totals = append(totals, []string{"Cargas", fmt.Sprint(r.Fills)})
	for _, d := range r.Distances {
		totals = append(totals, []string{"Distancia", Number(d.Amount, 0) + " " + Unit(d.Unit)})
	}
	table(m, []column{{"Totales", 4, false}, {"", 3, true}}, totals)

	section(m, "Detalle por unidad")
	rows := make([][]string, 0, len(r.Lines))
	for _, l := range r.Lines {
		rows = append(rows, []string{
			l.AssetName,
			l.LicensePlate,
			l.FuelType,
			fmt.Sprint(l.Fills),
			Number(l.Volume, 2) + " " + Unit(l.VolumeUnit),
			Money(l.Cost, currency),
			Optional(l.AvgUnitCost, 3),
			Number(l.Distance, 0) + " " + Unit(l.MeterUnit),
			Optional(l.Efficiency, 2),
		})
	}
	table(m, []column{
		{"Unidad", 2, false},
		{"Placas", 1, false},
		{"Combustible", 1, false},
		{"Cargas", 1, true},
		{"Volumen", 2, true},
		{"Costo", 2, true},
		{"Costo unitario", 1, true},
		{"Distancia", 1, true},
		{"Rendimiento", 1, true},
	}, rows)
	return finish(m)
}

var maintenanceColumns = []column{
	{"Unidad", 3, false},
	{"Placas", 2, false},
	{"Trabajos", 1, true},
	{"Refacciones", 2, true},
	{"Mano de obra", 2, true},
	{"Total", 2, true},
}

func maintenanceRows(lines []reports.MaintenanceLine, currency string) [][]string {
	rows := make([][]string, 0, len(lines))
	for _, l := range lines {
		total := Money(l.Total, currency)
		if l.HasOverride {
			total += " *"
		}
		rows = append(rows, []string{
			l.AssetName,
			l.LicensePlate,
			fmt.Sprint(l.Jobs),
			Money(l.Parts, currency),
			Money(l.Labor, currency),
			total,
		})
	}
	return rows
}

// MaintenanceMonthly renders the monthly maintenance cost report. logo may be nil.
func MaintenanceMonthly(r reports.MaintenanceMonthlyReport, logo []byte) ([]byte, error) {
	m, err := newDocument(r.Header, logo)
	if err != nil {
		return nil, err
	}
	if r.Empty() {
		m.AddRow(12, text.NewCol(12, emptyNotice, props.Text{Size: 11, Align: align.Center, Top: 4}))
		return finish(m)
	}
	currency := r.Company.Currency

	section(m, "Resumen")
	table(m, comparisonColumns, [][]string{
		comparisonRows("Costo total", r.Cost, func(c reports.Comparison) [3]string {
			return [3]string{Money(c.Current, currency), Money(c.Previous, currency), Money(c.Difference, currency)}
		}),
	})
	table(m, []column{{"Totales", 4, false}, {"", 3, true}}, [][]string{
		{"Trabajos", fmt.Sprint(r.Jobs)},
		{"Refacciones", Money(r.Parts, currency)},
		{"Mano de obra", Money(r.Labor, currency)},
		{"Total", Money(r.Total, currency)},
	})

	section(m, fmt.Sprintf("Las %d unidades con mayor costo", len(r.Top)))
	table(m, maintenanceColumns, maintenanceRows(r.Top, currency))

	section(m, "Detalle por unidad")
	table(m, maintenanceColumns, maintenanceRows(r.Lines, currency))

	if r.HasOverride {
		m.AddAutoRow(text.NewCol(12, overrideNote, props.Text{Size: 7, Align: align.Left}))
	}
	return finish(m)
}
