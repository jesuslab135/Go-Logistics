package reports

import (
	"sort"

	"github.com/shopspring/decimal"
)

// FuelRow is one asset and fuel type, aggregated over a period by the
// database. EffDistance and EffVolume cover only the entries whose efficiency
// the server could derive (full tank to full tank).
type FuelRow struct {
	AssetID      int64
	AssetName    string
	LicensePlate string
	MeterUnit    string
	VolumeUnit   string
	FuelType     string
	Fills        int64
	Volume       decimal.Decimal
	Cost         decimal.Decimal
	Distance     decimal.Decimal
	EffDistance  decimal.Decimal
	EffVolume    decimal.Decimal
}

type FuelLine struct {
	AssetName    string
	LicensePlate string
	FuelType     string
	Fills        int64
	Volume       decimal.Decimal
	VolumeUnit   string
	Cost         decimal.Decimal
	AvgUnitCost  *decimal.Decimal
	Distance     decimal.Decimal
	MeterUnit    string
	Efficiency   *decimal.Decimal
}

// VolumeTotal is the volume of one fuel type in one unit, set against the
// same fuel and unit a week earlier. Volumes in different units are never
// added together.
type VolumeTotal struct {
	FuelType string
	Unit     string
	Change   Comparison
}

type FuelWeeklyReport struct {
	Header
	Lines     []FuelLine
	Fills     int64
	TotalCost decimal.Decimal
	Volumes   []VolumeTotal
	Distances []UnitTotal
	Cost      Comparison
}

func (r FuelWeeklyReport) Empty() bool { return len(r.Lines) == 0 }

// volumeKey groups a volume total by fuel type and unit. Litres of diesel and
// litres of DEF are never summed together, and neither are litres and
// gallons of the same fuel.
type volumeKey struct{ fuelType, unit string }

// BuildFuelWeekly turns the aggregated rows of a week, and of the week before
// it, into the report.
func BuildFuelWeekly(h Header, current, previous []FuelRow) FuelWeeklyReport {
	r := FuelWeeklyReport{Header: h, Lines: make([]FuelLine, 0, len(current))}

	curVolumes := map[volumeKey]decimal.Decimal{}
	prevVolumes := map[volumeKey]decimal.Decimal{}
	// An asset that burns two fuels (diesel and DEF) has two rows covering the
	// same road, so its distance is the longest of them, not their sum.
	distances := map[int64]decimal.Decimal{}
	units := map[int64]string{}

	for _, row := range current {
		line := FuelLine{
			AssetName:    row.AssetName,
			LicensePlate: row.LicensePlate,
			FuelType:     row.FuelType,
			Fills:        row.Fills,
			Volume:       row.Volume,
			VolumeUnit:   row.VolumeUnit,
			Cost:         row.Cost,
			Distance:     row.Distance,
			MeterUnit:    row.MeterUnit,
		}
		if row.Volume.Sign() > 0 {
			avg := row.Cost.Div(row.Volume).Round(3)
			line.AvgUnitCost = &avg
		}
		if row.EffVolume.Sign() > 0 && row.EffDistance.Sign() > 0 {
			eff := row.EffDistance.Div(row.EffVolume).Round(2)
			line.Efficiency = &eff
		}
		r.Lines = append(r.Lines, line)

		r.Fills += row.Fills
		r.TotalCost = r.TotalCost.Add(row.Cost)
		key := volumeKey{row.FuelType, row.VolumeUnit}
		curVolumes[key] = curVolumes[key].Add(row.Volume)
		if d, seen := distances[row.AssetID]; !seen || row.Distance.GreaterThan(d) {
			distances[row.AssetID] = row.Distance
		}
		units[row.AssetID] = row.MeterUnit
	}

	prevCost := decimal.Zero
	for _, row := range previous {
		prevCost = prevCost.Add(row.Cost)
		key := volumeKey{row.FuelType, row.VolumeUnit}
		prevVolumes[key] = prevVolumes[key].Add(row.Volume)
	}

	keys := make(map[volumeKey]bool, len(curVolumes)+len(prevVolumes))
	for k := range curVolumes {
		keys[k] = true
	}
	for k := range prevVolumes {
		keys[k] = true
	}
	for k := range keys {
		r.Volumes = append(r.Volumes, VolumeTotal{
			FuelType: k.fuelType,
			Unit:     k.unit,
			Change:   Compare(curVolumes[k], prevVolumes[k]),
		})
	}
	sort.Slice(r.Volumes, func(i, j int) bool {
		if r.Volumes[i].FuelType != r.Volumes[j].FuelType {
			return r.Volumes[i].FuelType < r.Volumes[j].FuelType
		}
		return r.Volumes[i].Unit < r.Volumes[j].Unit
	})

	byUnit := map[string]decimal.Decimal{}
	for asset, d := range distances {
		byUnit[units[asset]] = byUnit[units[asset]].Add(d)
	}
	for unit, amount := range byUnit {
		r.Distances = append(r.Distances, UnitTotal{Unit: unit, Amount: amount})
	}
	sort.Slice(r.Distances, func(i, j int) bool { return r.Distances[i].Unit < r.Distances[j].Unit })

	r.Cost = Compare(r.TotalCost, prevCost)
	return r
}
