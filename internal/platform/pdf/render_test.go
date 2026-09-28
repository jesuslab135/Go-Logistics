package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"

	"fleet/internal/platform/reports"
)

func header(t *testing.T, kind reports.Kind) reports.Header {
	t.Helper()
	mx, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 28, 6, 0, 0, 0, mx)
	return reports.Header{
		Kind:        kind,
		Company:     reports.Company{ID: 1, Name: "Transportes Durán S.A. de C.V.", Currency: "MXN", Timezone: mx.String()},
		Period:      reports.LatestDue(kind, now, mx, 6),
		GeneratedAt: now,
	}
}

func testLogo(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for x := range 40 {
		for y := range 20 {
			img.Set(x, y, color.RGBA{R: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func assertPDF(t *testing.T, out []byte, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Fatalf("output does not start with %%PDF-: %q", out[:min(20, len(out))])
	}
}

// pageCount counts page objects. maroto writes each as "/Type /Page" followed
// by a line break; "/Type /Pages" is the page tree and must not be counted.
func pageCount(out []byte) int {
	return bytes.Count(out, []byte("/Type /Page\n"))
}

func fuelRows(n int) []reports.FuelRow {
	rows := make([]reports.FuelRow, 0, n)
	for i := range n {
		rows = append(rows, reports.FuelRow{
			AssetID: int64(i + 1), AssetName: fmt.Sprintf("Unidad %03d", i+1), LicensePlate: "ABC-123",
			MeterUnit: "km", VolumeUnit: "liters", FuelType: "diesel", Fills: 3,
			Volume: dec("300"), Cost: dec("7500"), Distance: dec("900"),
			EffDistance: dec("600"), EffVolume: dec("200"),
		})
	}
	return rows
}

func TestFuelWeekly(t *testing.T) {
	r := reports.BuildFuelWeekly(header(t, reports.FuelWeekly), fuelRows(3), fuelRows(2))
	out, err := FuelWeekly(r, testLogo(t))
	assertPDF(t, out, err)
}

func TestFuelWeeklyEmpty(t *testing.T) {
	r := reports.BuildFuelWeekly(header(t, reports.FuelWeekly), nil, nil)
	out, err := FuelWeekly(r, nil)
	assertPDF(t, out, err)
	if pageCount(out) != 1 {
		t.Fatalf("pages = %d, want 1 for an empty report", pageCount(out))
	}
}

// TestFuelWeeklyRendersEachFuelAndUnit exercises the per-(fuel type, unit)
// loop over r.Volumes in the "Resumen" table with more than one entry: three
// combinations in the current week (diesel/liters, def/liters,
// diesel/gallons) plus a fourth, gasoline/liters, that appears only in the
// previous week. That gives one entry with a zero previous (nil Percent —
// there is no percentage change from nothing) and one with a zero current
// (a non-nil Percent, since its previous is non-zero).
func TestFuelWeeklyRendersEachFuelAndUnit(t *testing.T) {
	current := []reports.FuelRow{
		{AssetID: 1, AssetName: "Unidad 001", LicensePlate: "ABC-123", MeterUnit: "km", VolumeUnit: "liters", FuelType: "diesel", Fills: 2,
			Volume: dec("200"), Cost: dec("5000"), Distance: dec("600")},
		{AssetID: 2, AssetName: "Unidad 002", LicensePlate: "ABC-124", MeterUnit: "km", VolumeUnit: "liters", FuelType: "def", Fills: 1,
			Volume: dec("50"), Cost: dec("900"), Distance: dec("300")},
		{AssetID: 3, AssetName: "Unidad 003", LicensePlate: "ABC-125", MeterUnit: "km", VolumeUnit: "gallons", FuelType: "diesel", Fills: 1,
			Volume: dec("80"), Cost: dec("3000"), Distance: dec("500")},
	}
	previous := []reports.FuelRow{
		{AssetID: 4, AssetName: "Unidad 004", LicensePlate: "XYZ-999", MeterUnit: "km", VolumeUnit: "liters", FuelType: "gasoline", Fills: 2,
			Volume: dec("120"), Cost: dec("2400"), Distance: dec("400")},
	}
	r := reports.BuildFuelWeekly(header(t, reports.FuelWeekly), current, previous)

	want := []struct{ fuelType, unit string }{
		{"def", "liters"},
		{"diesel", "gallons"},
		{"diesel", "liters"},
		{"gasoline", "liters"},
	}
	if len(r.Volumes) != len(want) {
		t.Fatalf("Volumes = %d entries, want %d: %+v", len(r.Volumes), len(want), r.Volumes)
	}
	for i, w := range want {
		if r.Volumes[i].FuelType != w.fuelType || r.Volumes[i].Unit != w.unit {
			t.Errorf("Volumes[%d] = %s/%s, want %s/%s", i, r.Volumes[i].FuelType, r.Volumes[i].Unit, w.fuelType, w.unit)
		}
	}
	// gasoline/liters exists only in the previous week: zero current, and a
	// Percent since its previous is non-zero.
	gasoline := r.Volumes[3]
	if !gasoline.Change.Current.IsZero() {
		t.Errorf("gasoline current = %s, want 0", gasoline.Change.Current)
	}
	if gasoline.Change.Percent == nil {
		t.Error("gasoline Percent = nil, want a change computed against a non-zero previous")
	}
	// def/liters, diesel/gallons and diesel/liters exist only in the current
	// week: zero previous, and no Percent.
	for _, v := range r.Volumes[:3] {
		if !v.Change.Previous.IsZero() {
			t.Errorf("%s/%s previous = %s, want 0", v.FuelType, v.Unit, v.Change.Previous)
		}
		if v.Change.Percent != nil {
			t.Errorf("%s/%s Percent = %v, want nil", v.FuelType, v.Unit, *v.Change.Percent)
		}
	}

	out, err := FuelWeekly(r, nil)
	assertPDF(t, out, err)

	// maroto does not Flate-compress content streams in this configuration
	// (Step 8's sample PDF could be grepped for plain Spanish text directly,
	// and a throwaway probe confirmed it for this very report), so checking
	// the rendered row labels as raw bytes is reliable. PDF string literals
	// escape "(" and ")" as "\(" and "\)".
	for _, label := range []string{
		`Volumen def \(L\)`,
		`Volumen diesel \(gal\)`,
		`Volumen diesel \(L\)`,
		`Volumen gasoline \(L\)`,
	} {
		if !bytes.Contains(out, []byte(label)) {
			t.Errorf("output does not contain label %q", label)
		}
	}
}

// A fleet of 500 units must flow onto more pages, not fail and not be cut off.
func TestFuelWeeklyLargeFleetPaginates(t *testing.T) {
	r := reports.BuildFuelWeekly(header(t, reports.FuelWeekly), fuelRows(500), nil)
	out, err := FuelWeekly(r, testLogo(t))
	assertPDF(t, out, err)
	if pageCount(out) < 5 {
		t.Fatalf("pages = %d, want several for 500 rows", pageCount(out))
	}
}

// Asset names are free text up to 100 characters.
func TestLongNamesAndAccentsRender(t *testing.T) {
	rows := fuelRows(2)
	rows[0].AssetName = strings.Repeat("Tractocamión Kenworth ", 5)[:100]
	rows[1].AssetName = "Ñandú — «grúa» 100% añejo"
	r := reports.BuildFuelWeekly(header(t, reports.FuelWeekly), rows, nil)
	out, err := FuelWeekly(r, nil)
	assertPDF(t, out, err)
}

// A logo maroto cannot embed is dropped; the report still renders.
func TestUnsupportedOrBrokenLogoIsIgnored(t *testing.T) {
	r := reports.BuildFuelWeekly(header(t, reports.FuelWeekly), fuelRows(1), nil)
	logos := map[string][]byte{
		"webp":      []byte("RIFF\x00\x00\x00\x00WEBPVP8 "),
		"gif":       []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00;"),
		"not image": []byte("<html>404</html>"),
		"empty":     {},
	}
	for name, logo := range logos {
		t.Run(name, func(t *testing.T) {
			out, err := FuelWeekly(r, logo)
			assertPDF(t, out, err)
		})
	}
}

func TestMaintenanceMonthly(t *testing.T) {
	var rows []reports.MaintenanceRow
	for i := range 8 {
		rows = append(rows, reports.MaintenanceRow{
			AssetID: int64(i + 1), AssetName: fmt.Sprintf("Unidad %d", i+1), LicensePlate: "XYZ-1",
			Jobs: 2, Parts: dec("1000"), Labor: dec("500"), Total: dec("1500"),
		})
	}
	rows[0].HasOverride = true
	rows[0].Total = dec("9000")

	r := reports.BuildMaintenanceMonthly(header(t, reports.MaintenanceMonthly), rows, nil)
	out, err := MaintenanceMonthly(r, testLogo(t))
	assertPDF(t, out, err)

	empty := reports.BuildMaintenanceMonthly(header(t, reports.MaintenanceMonthly), nil, nil)
	out, err = MaintenanceMonthly(empty, nil)
	assertPDF(t, out, err)
}
