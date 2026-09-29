//go:build integration

package handler

// The report endpoints over HTTP: validation, tenant scoping, permissions,
// and that a download is a PDF and leaves no run behind.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestRecipientEndpoints(t *testing.T) {
	f := newReportFixture(t)
	base := f.srv + "/api/v1/reports/recipients"

	var created struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	}
	postJSON(t, base, f.token, map[string]any{"report_kind": "fuel_weekly", "email": "  Flota@Cliente.mx "}, http.StatusCreated, &created)
	if created.Email != "Flota@Cliente.mx" {
		t.Fatalf("email = %q, want it trimmed", created.Email)
	}

	// The same mailbox in another letter case is the same mailbox.
	postJSON(t, base, f.token, map[string]any{"report_kind": "fuel_weekly", "email": "flota@cliente.mx"}, http.StatusConflict, nil)
	// The same address for the other report is a different subscription.
	postJSON(t, base, f.token, map[string]any{"report_kind": "maintenance_monthly", "email": "flota@cliente.mx"}, http.StatusCreated, nil)

	for name, body := range map[string]map[string]any{
		"unknown kind":     {"report_kind": "tires_daily", "email": "a@b.mx"},
		"not an address":   {"report_kind": "fuel_weekly", "email": "flota"},
		"display name":     {"report_kind": "fuel_weekly", "email": "Flota <a@b.mx>"},
		"two addresses":    {"report_kind": "fuel_weekly", "email": "a@b.mx, c@d.mx"},
		"header injection": {"report_kind": "fuel_weekly", "email": "a@b.mx\r\nBcc: espia@otro.mx"},
		"missing email":    {"report_kind": "fuel_weekly"},
	} {
		t.Run(name, func(t *testing.T) {
			postJSON(t, base, f.token, body, http.StatusUnprocessableEntity, nil)
		})
	}

	var list struct {
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	getJSON(t, base, f.token, http.StatusOK, &list)
	if len(list.Data) != 2 {
		t.Fatalf("recipients = %d, want 2", len(list.Data))
	}
	getJSON(t, base+"?report_kind=fuel_weekly", f.token, http.StatusOK, &list)
	if len(list.Data) != 1 {
		t.Fatalf("fuel recipients = %d, want 1", len(list.Data))
	}
	getJSON(t, base+"?report_kind=nope", f.token, http.StatusBadRequest, nil)

	// Another company of the same owner can neither see nor delete it.
	other := createCompany(t, f.srv, f.owner, "Otra Empresa", fmt.Sprintf("TAX-OTRA-%d", time.Now().UnixNano()))
	otherToken := switchCompany(t, f.srv, f.owner, other)
	getJSON(t, base, otherToken, http.StatusOK, &list)
	if len(list.Data) != 0 {
		t.Fatalf("another company sees %d recipients, want 0", len(list.Data))
	}
	doJSON(t, http.MethodDelete, fmt.Sprintf("%s/%d", base, created.ID), otherToken, nil, http.StatusNotFound, nil)

	doJSON(t, http.MethodDelete, fmt.Sprintf("%s/%d", base, created.ID), f.token, nil, http.StatusNoContent, nil)
	doJSON(t, http.MethodDelete, fmt.Sprintf("%s/%d", base, created.ID), f.token, nil, http.StatusNotFound, nil)
}

func (f reportFixture) getPDF(t *testing.T, path, token string, want int) (*http.Response, []byte) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, f.srv+"/api/v1"+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != want {
		t.Fatalf("GET %s: status = %d, want %d: %s", path, resp.StatusCode, want, body)
	}
	return resp, body
}

func TestDownloadEndpoints(t *testing.T) {
	f := newReportFixture(t)

	for _, path := range []string{
		"/reports/fuel/weekly",
		"/reports/fuel/weekly?week=2026-09-23",
		"/reports/maintenance/monthly",
		"/reports/maintenance/monthly?month=2026-08",
	} {
		resp, body := f.getPDF(t, path, f.token, http.StatusOK)
		if got := resp.Header.Get("Content-Type"); got != "application/pdf" {
			t.Fatalf("%s: Content-Type = %q", path, got)
		}
		if !bytes.HasPrefix(body, []byte("%PDF-")) {
			t.Fatalf("%s: the body is not a PDF", path)
		}
	}

	resp, _ := f.getPDF(t, "/reports/fuel/weekly?week=2026-09-23", f.token, http.StatusOK)
	if got, want := resp.Header.Get("Content-Disposition"), `attachment; filename="combustible-semanal-2026-09-21.pdf"`; got != want {
		t.Fatalf("Content-Disposition = %q, want %q", got, want)
	}

	for _, path := range []string{
		"/reports/fuel/weekly?week=23/09/2026",
		"/reports/fuel/weekly?week=2026-13-40",
		"/reports/maintenance/monthly?month=agosto",
	} {
		f.getPDF(t, path, f.token, http.StatusBadRequest)
	}

	// A download is not a run: it must not mark the period as handled.
	var n int
	if err := f.pool.QueryRow(context.Background(), `SELECT count(*) FROM report_run`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("report_run has %d rows after downloads, want 0", n)
	}
}

// A role that does not grant the module is refused on every route.
func TestReportsNeedTheModulePermission(t *testing.T) {
	f := newReportFixture(t)
	limited := f.limitedToken(t)

	getJSON(t, f.srv+"/api/v1/reports/recipients", limited, http.StatusForbidden, nil)
	getJSON(t, f.srv+"/api/v1/reports/runs", limited, http.StatusForbidden, nil)
	postJSON(t, f.srv+"/api/v1/reports/recipients", limited,
		map[string]any{"report_kind": "fuel_weekly", "email": "a@b.mx"}, http.StatusForbidden, nil)
	f.getPDF(t, "/reports/fuel/weekly", limited, http.StatusForbidden)
}
