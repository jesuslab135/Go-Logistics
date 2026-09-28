package pdf

import (
	"context"
	"fmt"
	"time"

	"fleet/internal/platform/reports"
)

// Document is a rendered report, ready to download or attach.
type Document struct {
	Filename string
	Data     []byte
	// Empty says the period had no data. A download still serves the document;
	// the scheduler does not email it.
	Empty bool
}

// Render reads one company's report for one period and renders it. It is the
// single path behind the scheduler, the download endpoints and the CLI, so
// what is emailed and what is downloaded cannot differ.
func Render(ctx context.Context, q reports.Querier, kind reports.Kind, c reports.Company, p reports.Period, logo []byte, now time.Time) (Document, error) {
	doc := Document{
		Filename: fmt.Sprintf("%s-%s.pdf", kind.FilePrefix(), p.StartDate().Format("2006-01-02")),
	}

	var err error
	switch kind {
	case reports.FuelWeekly:
		var r reports.FuelWeeklyReport
		if r, err = reports.FuelWeeklyFor(ctx, q, c, p, now); err != nil {
			return Document{}, err
		}
		doc.Empty = r.Empty()
		doc.Data, err = FuelWeekly(r, logo)
	case reports.MaintenanceMonthly:
		var r reports.MaintenanceMonthlyReport
		if r, err = reports.MaintenanceMonthlyFor(ctx, q, c, p, now); err != nil {
			return Document{}, err
		}
		doc.Empty = r.Empty()
		doc.Data, err = MaintenanceMonthly(r, logo)
	default:
		return Document{}, fmt.Errorf("pdf: unknown report kind %q", kind)
	}
	if err != nil {
		return Document{}, err
	}
	return doc, nil
}
