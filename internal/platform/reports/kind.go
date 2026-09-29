// Package reports computes the scheduled reports. It knows the data and the
// arithmetic; it does not know PDF, email or HTTP.
package reports

// Kind names a report. The values are stored in report_recipient.report_kind
// and report_run.report_kind, so they are part of the schema.
type Kind string

const (
	FuelWeekly         Kind = "fuel_weekly"
	MaintenanceMonthly Kind = "maintenance_monthly"
)

// Kinds lists every report the scheduler produces, in the order it runs them.
var Kinds = []Kind{FuelWeekly, MaintenanceMonthly}

func ParseKind(s string) (Kind, bool) {
	for _, k := range Kinds {
		if string(k) == s {
			return k, true
		}
	}
	return "", false
}

// Title is the heading printed on the PDF and used as the email subject.
func (k Kind) Title() string {
	if k == MaintenanceMonthly {
		return "Reporte mensual de costos de mantenimiento"
	}
	return "Reporte semanal de combustible"
}

// FilePrefix starts the attachment's file name.
func (k Kind) FilePrefix() string {
	if k == MaintenanceMonthly {
		return "mantenimiento-mensual"
	}
	return "combustible-semanal"
}
