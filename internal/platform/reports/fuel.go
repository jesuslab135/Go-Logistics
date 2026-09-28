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

type FuelWeeklyReport struct {
	Header
	Lines     []FuelLine
	Fills     int64
	TotalCost decimal.Decimal
	Volumes   []UnitTotal
	Distances []UnitTotal
	Cost      Comparison
	Volume    Comparison
}

func (r FuelWeeklyReport) Empty() bool { return len(r.Lines) == 0 }

// BuildFuelWeekly turns the aggregated rows of a week, and of the week before
// it, into the report.
func BuildFuelWeekly(h Header, current, previous []FuelRow) FuelWeeklyReport {
	r := FuelWeeklyReport{Header: h, Lines: make([]FuelLine, 0, len(current))}

	volumes := map[[2]string]decimal.Decimal{}
	// An asset that burns two fuels (diesel and DEF) has two rows covering the
	// same road, so its distance is the longest of them, not their sum.
	distances := map[int64]decimal.Decimal{}
	units := map[int64]string{}
	rawVolume := decimal.Zero

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
		rawVolume = rawVolume.Add(row.Volume)
		key := [2]string{row.FuelType, row.VolumeUnit}
		volumes[key] = volumes[key].Add(row.Volume)
		if d, seen := distances[row.AssetID]; !seen || row.Distance.GreaterThan(d) {
			distances[row.AssetID] = row.Distance
		}
		units[row.AssetID] = row.MeterUnit
	}

	for key, amount := range volumes {
		r.Volumes = append(r.Volumes, UnitTotal{Label: key[0], Unit: key[1], Amount: amount})
	}
	sort.Slice(r.Volumes, func(i, j int) bool {
		if r.Volumes[i].Label != r.Volumes[j].Label {
			return r.Volumes[i].Label < r.Volumes[j].Label
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

	prevCost, prevVolume := decimal.Zero, decimal.Zero
	for _, row := range previous {
		prevCost = prevCost.Add(row.Cost)
		prevVolume = prevVolume.Add(row.Volume)
	}
	r.Cost = Compare(r.TotalCost, prevCost)
	r.Volume = Compare(rawVolume, prevVolume)
	return r
}
