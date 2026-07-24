package tire

// WheelPositionDefinition — port of api/models/tire_model.py:102
// Code follows the TMC convention (LF, RF, LRI, LRO). Slot 1 = inner, 2 = outer.

type WheelPositionDefinition struct {
	ID     int64
	AxleID int64
	Code   string
	Side   string
	Slot   int32
}
