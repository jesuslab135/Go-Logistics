package filter

import "testing"

func TestWhereRawNumbersItsOwnPlaceholders(t *testing.T) {
	w := NewWhere(1).
		Add("a.company_id", Eq, int64(3)).
		Raw("(a.name ILIKE ? OR a.vin_sn ILIKE ?)", "%x%", "%x%")

	const want = "a.company_id = $1 AND (a.name ILIKE $2 OR a.vin_sn ILIKE $3)"
	if got := w.SQL(); got != want {
		t.Errorf("SQL() = %q, want %q", got, want)
	}
	if got := len(w.Args()); got != 3 {
		t.Errorf("len(Args()) = %d, want 3", got)
	}
	if got := w.Next(); got != 4 {
		t.Errorf("Next() = %d, want 4", got)
	}
}

func TestWhereRawContinuesFromExistingOffset(t *testing.T) {
	w := NewWhere(5).Raw("x = ?", 1)

	if got := w.SQL(); got != "x = $5" {
		t.Errorf("SQL() = %q, want %q", got, "x = $5")
	}
}

func TestWhereEmptyRendersNothing(t *testing.T) {
	w := NewWhere(1)

	if got := w.SQL(); got != "" {
		t.Errorf("SQL() = %q, want empty", got)
	}
	if got := w.Next(); got != 1 {
		t.Errorf("Next() = %d, want 1", got)
	}
}
