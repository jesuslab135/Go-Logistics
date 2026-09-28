package reports

import (
	"testing"

	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func assertDec(t *testing.T, name string, got decimal.Decimal, want string) {
	t.Helper()
	if !got.Equal(dec(want)) {
		t.Fatalf("%s = %s, want %s", name, got, want)
	}
}

func TestCompare(t *testing.T) {
	c := Compare(dec("150"), dec("120"))
	assertDec(t, "difference", c.Difference, "30")
	if c.Percent == nil {
		t.Fatal("percent is nil")
	}
	assertDec(t, "percent", *c.Percent, "25")

	down := Compare(dec("90"), dec("120"))
	assertDec(t, "difference", down.Difference, "-30")
	assertDec(t, "percent", *down.Percent, "-25")

	// No division by zero, and no invented percentage, when there is nothing
	// to compare against.
	fresh := Compare(dec("90"), decimal.Zero)
	if fresh.Percent != nil {
		t.Fatalf("percent = %s, want nil when the previous period is zero", fresh.Percent)
	}
	assertDec(t, "difference", fresh.Difference, "90")
}

func TestBuildFuelWeekly(t *testing.T) {
	current := []FuelRow{
		{AssetID: 1, AssetName: "Unidad 7", MeterUnit: "km", VolumeUnit: "liters", FuelType: "diesel",
			Fills: 3, Volume: dec("300"), Cost: dec("7500"), Distance: dec("900"), EffDistance: dec("600"), EffVolume: dec("200")},
		{AssetID: 1, AssetName: "Unidad 7", MeterUnit: "km", VolumeUnit: "liters", FuelType: "def",
			Fills: 1, Volume: dec("20"), Cost: dec("400"), Distance: dec("850")},
		{AssetID: 2, AssetName: "Unidad 9", MeterUnit: "mi", VolumeUnit: "gallons", FuelType: "diesel",
			Fills: 1, Volume: dec("50"), Cost: dec("2100"), Distance: decimal.Zero},
	}
	previous := []FuelRow{{AssetID: 1, Volume: dec("400"), Cost: dec("8000")}}

	r := BuildFuelWeekly(Header{Kind: FuelWeekly}, current, previous)

	if r.Empty() || len(r.Lines) != 3 {
		t.Fatalf("lines = %d, want 3", len(r.Lines))
	}
	if r.Fills != 5 {
		t.Fatalf("fills = %d, want 5", r.Fills)
	}
	assertDec(t, "total cost", r.TotalCost, "10000")

	first := r.Lines[0]
	assertDec(t, "average unit cost", *first.AvgUnitCost, "25")
	// 600 over 200, not 900 over 300: only full-tank intervals count.
	assertDec(t, "efficiency", *first.Efficiency, "3")
	if r.Lines[1].Efficiency != nil {
		t.Fatal("a row with no derivable interval must have no efficiency")
	}

	if len(r.Volumes) != 3 {
		t.Fatalf("volume totals = %+v, want one per fuel type and unit", r.Volumes)
	}
	if got := r.Volumes[0]; got.Label != "def" || got.Unit != "liters" || !got.Amount.Equal(dec("20")) {
		t.Fatalf("first volume total = %+v", got)
	}

	if len(r.Distances) != 2 {
		t.Fatalf("distance totals = %+v, want km and mi", r.Distances)
	}
	// Unidad 7 drove 900 km once, not 900 + 850.
	if got := r.Distances[0]; got.Unit != "km" || !got.Amount.Equal(dec("900")) {
		t.Fatalf("km total = %+v, want 900", got)
	}

	assertDec(t, "cost difference", r.Cost.Difference, "2000")
	assertDec(t, "cost percent", *r.Cost.Percent, "25")
	assertDec(t, "volume previous", r.Volume.Previous, "400")
}

func TestBuildFuelWeeklyEmpty(t *testing.T) {
	r := BuildFuelWeekly(Header{}, nil, nil)
	if !r.Empty() {
		t.Fatal("no rows must be an empty report")
	}
	if r.Cost.Percent != nil {
		t.Fatal("an empty report has no percentage")
	}
	assertDec(t, "total cost", r.TotalCost, "0")
}

// A fill of zero volume (a correction row) must not divide by zero.
func TestBuildFuelWeeklyZeroVolume(t *testing.T) {
	r := BuildFuelWeekly(Header{}, []FuelRow{{AssetID: 1, AssetName: "U", Fills: 1, Cost: dec("10")}}, nil)
	if r.Lines[0].AvgUnitCost != nil || r.Lines[0].Efficiency != nil {
		t.Fatalf("line = %+v, want no ratios", r.Lines[0])
	}
}

func TestBuildMaintenanceMonthly(t *testing.T) {
	var current []MaintenanceRow
	for i := range 7 {
		current = append(current, MaintenanceRow{
			AssetID: int64(i + 1), AssetName: "U", Jobs: 2,
			Parts: dec("100"), Labor: dec("50"), Total: dec("150"),
		})
	}
	// An overridden job: the total is what somebody typed, not parts + labor.
	current[0].Total = dec("1000")
	current[0].HasOverride = true
	previous := []MaintenanceRow{{Total: dec("950")}}

	r := BuildMaintenanceMonthly(Header{Kind: MaintenanceMonthly}, current, previous)

	if len(r.Lines) != 7 || len(r.Top) != 5 {
		t.Fatalf("lines = %d, top = %d, want 7 and 5", len(r.Lines), len(r.Top))
	}
	if r.Jobs != 14 {
		t.Fatalf("jobs = %d, want 14", r.Jobs)
	}
	assertDec(t, "parts", r.Parts, "700")
	assertDec(t, "labor", r.Labor, "350")
	assertDec(t, "total", r.Total, "1900")
	if !r.HasOverride {
		t.Fatal("the report must say that a total was overridden")
	}
	assertDec(t, "percent", *r.Cost.Percent, "100")
}

func TestBuildMaintenanceMonthlyFewerThanTop(t *testing.T) {
	r := BuildMaintenanceMonthly(Header{}, []MaintenanceRow{{AssetName: "U", Total: dec("1")}}, nil)
	if len(r.Top) != 1 {
		t.Fatalf("top = %d, want 1", len(r.Top))
	}
	if empty := BuildMaintenanceMonthly(Header{}, nil, nil); !empty.Empty() || len(empty.Top) != 0 {
		t.Fatal("no rows must be an empty report with an empty top")
	}
}
