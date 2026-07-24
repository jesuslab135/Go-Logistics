package tire

// AxleDefinition — port of api/models/tire_model.py:46
// PositionsPerSide: 1 = single wheel, 2 = dual (inner+outer). Django validated
// the 1..4 range at the form layer only.

type AxleDefinition struct {
	ID               int64
	TemplateID       int64
	PositionIndex    int32
	Label            string
	AxleRole         string
	PositionsPerSide int32
}

func (a AxleDefinition) WheelCount() int32 {
	return a.PositionsPerSide * 2
}
