//go:build integration

package handler

// This file exercises the onboarding chain against a real, throwaway
// Postgres database rather than mocks — the chain the whole-branch review
// named as the one thing no unit test in this repository touches:
//
//	POST /admin/accounts -> login as the owner -> POST /companies ->
//	POST /employees -> PUT /admin/employees/{id}/companies
//
// Every defect that review found would have failed this test: a missing
// account_id on CreateEmployee (the composite FK rejects the membership
// AddEmployeeCompany then tries to create), the silent no-op that shape of
// query briefly introduced (a membership the database never created being
// reported as created), and /admin/* routes wired under the company-
// membership gate (a genuine platform admin, who holds no membership by
// design, locked out of the one route this project exists to add).
//
// It is intentionally excluded from `go test ./...` by the build tag above,
// so CI is unaffected unless it is deliberately opted into. See README.md,
// "Integration tests", for how to run it.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/auth"
	"fleet/internal/db/gen"
)

// integrationAdminDSN is where to find a Postgres server to create throwaway
// databases in. Defaults to this project's own docker-compose db service
// (`docker compose up -d db`), connecting to its default "postgres"
// maintenance database — the same host and credentials README.md's local
// setup already uses, just not the "fleet" application database, so this
// test's schema churn can never touch a developer's own data.
func integrationAdminDSN() string {
	if v := os.Getenv("INTEGRATION_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable"
}

// setupThrowawayDB creates a uniquely named database, runs every migration in
// internal/db/migrations against it in order, and returns a pool connected to
// it. The database is dropped when the test finishes.
func setupThrowawayDB(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	admin, err := pgx.Connect(ctx, integrationAdminDSN())
	if err != nil {
		t.Fatalf("connect to Postgres for the throwaway database (is `docker compose up -d db` running? "+
			"override with INTEGRATION_DATABASE_URL): %v", err)
	}
	defer admin.Close(ctx)

	dbName := fmt.Sprintf("fleet_it_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{dbName}.Sanitize()); err != nil {
		t.Fatalf("create throwaway database %s: %v", dbName, err)
	}
	t.Cleanup(func() {
		cleanup, err := pgx.Connect(context.Background(), integrationAdminDSN())
		if err != nil {
			t.Logf("could not connect to drop %s, leaving it behind: %v", dbName, err)
			return
		}
		defer cleanup.Close(context.Background())
		if _, err := cleanup.Exec(context.Background(), "DROP DATABASE "+pgx.Identifier{dbName}.Sanitize()); err != nil {
			t.Logf("could not drop throwaway database %s, leaving it behind: %v", dbName, err)
		}
	})

	dsn := dsnWithDatabase(integrationAdminDSN(), dbName)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to throwaway database %s: %v", dbName, err)
	}
	t.Cleanup(pool.Close)

	runMigrations(t, ctx, pool)
	return pool
}

// dsnWithDatabase swaps the path component of a postgres:// DSN, keeping
// host, credentials and query parameters (sslmode, etc.) intact.
func dsnWithDatabase(dsn, dbName string) string {
	i := bytes.LastIndexByte([]byte(dsn), '/')
	if i < 0 {
		return dsn
	}
	base, rest := dsn[:i], dsn[i+1:]
	if q := bytes.IndexByte([]byte(rest), '?'); q >= 0 {
		return base + "/" + dbName + rest[q:]
	}
	return base + "/" + dbName
}

// runMigrations applies every *.up.sql file in internal/db/migrations, in
// filename order — the same files `docker compose up migrate` applies, read
// directly rather than through a second migration engine so there is nothing
// here that can drift from what actually ships. Each file is sent as one
// multi-statement Exec (pgx uses the simple query protocol when a query has
// no arguments, which Postgres itself executes statement-by-statement,
// including the DO $$ ... $$ blocks a couple of these files use), so no SQL
// splitting logic is needed.
func runMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve this test file's own path")
	}
	dir := filepath.Join(filepath.Dir(thisFile), "..", "..", "db", "migrations")

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations directory %s: %v", dir, err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && bytes.HasSuffix([]byte(e.Name()), []byte(".up.sql")) {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		t.Fatalf("no *.up.sql files found in %s", dir)
	}

	for _, name := range files {
		sql, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
	}
}

// newIntegrationRouter wires the real router over the throwaway pool, the
// same assembly cmd/api/main.go does. Storage is nil: nothing on the chain
// under test (accounts, companies, employees, memberships) touches it, and
// storage.Storage is an interface, so a nil value is inert until a method is
// called on it.
func newIntegrationRouter(pool *pgxpool.Pool) http.Handler {
	q := gen.New(pool)
	return NewRouter(Deps{
		Queries:     q,
		Pool:        pool,
		Tokens:      auth.NewTokenService("integration-test-secret", "fleet-integration-test", time.Hour, 24*time.Hour),
		Verifier:    NewEmployeeCredentialVerifier(q),
		Storage:     nil,
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		CORSOrigins: []string{"*"},
	})
}

// seedPlatformAdmin inserts a platform-admin employee directly (the CLI's own
// job in production — see cmd/cli platform-admin and README.md's "Create the
// first user and log in"), then logs in through the real /auth/login route so
// the rest of the test only ever talks to the API over HTTP, never the
// database, for anything the API itself is supposed to do. The employee has
// an account (so mayLogIn admits them) and deliberately no company: this is
// the exact "platform staff hold no company memberships" shape README.md's
// onboarding section documents, and the one the /admin/* routing bug
// (registerAdminRoutes under RequireCompanyMember) would have failed.
func seedPlatformAdmin(t *testing.T, ctx context.Context, pool *pgxpool.Pool, baseURL string) string {
	t.Helper()

	hash, err := auth.HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatalf("hash platform admin password: %v", err)
	}

	const email = "platform-admin@integration.test"
	var accountID int64
	if err := pool.QueryRow(ctx,
		"INSERT INTO account (name, created_at) VALUES ($1, now()) RETURNING id",
		"Integration Platform Ops",
	).Scan(&accountID); err != nil {
		t.Fatalf("seed platform admin account: %v", err)
	}

	var employeeID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO employee (
			account_id, first_name, last_name, employee_id, email, mobile_phone,
			work_phone, job_title, license_class, license_number, license_state,
			street_address, city, region, postal_code, country, password_hash,
			is_platform_admin, updated_at
		) VALUES ($1,'Platform','Admin','EMP-IT-ADMIN',$2,'','','','','','','','','','','',$3,true,now())
		RETURNING id`,
		accountID, email, hash,
	).Scan(&employeeID); err != nil {
		t.Fatalf("seed platform admin employee: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE account SET owner_employee_id = $1 WHERE id = $2", employeeID, accountID); err != nil {
		t.Fatalf("set platform admin as their own account's owner: %v", err)
	}

	var login struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, baseURL+"/auth/login", "", map[string]any{"email": email, "password": "correct-horse-battery"}, http.StatusOK, &login)
	if login.AccessToken == "" {
		t.Fatal("platform admin login returned no access token")
	}
	return login.AccessToken
}

// postJSON and putJSON issue a request with an optional bearer token, decode
// the JSON body into out (when non-nil), and fail the test if the status
// does not match wantStatus. They return the raw response body so a caller
// can additionally inspect it (e.g. to assert a field's absence).
func postJSON(t *testing.T, url, token string, body any, wantStatus int, out any) []byte {
	t.Helper()
	return doJSON(t, http.MethodPost, url, token, body, wantStatus, out)
}

func putJSON(t *testing.T, url, token string, body any, wantStatus int, out any) []byte {
	t.Helper()
	return doJSON(t, http.MethodPut, url, token, body, wantStatus, out)
}

func getJSON(t *testing.T, url, token string, wantStatus int, out any) []byte {
	t.Helper()
	return doJSON(t, http.MethodGet, url, token, nil, wantStatus, out)
}

func doJSON(t *testing.T, method, url, token string, body any, wantStatus int, out any) []byte {
	t.Helper()

	buf, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if resp.StatusCode != wantStatus {
		t.Fatalf("%s %s: status = %d, want %d, body = %s", method, url, resp.StatusCode, wantStatus, respBody)
	}
	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			t.Fatalf("decode response from %s %s: %v, body = %s", method, url, err, respBody)
		}
	}
	return respBody
}

func TestOnboardingChain(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	router := newIntegrationRouter(pool)

	srv := httptest.NewServer(router)
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)

	// 1. POST /admin/accounts: a platform admin who holds no company
	// membership provisions a client and its owner in one call.
	ownerEmail := fmt.Sprintf("owner-%d@integration.test", time.Now().UnixNano())
	var account struct {
		ID              int64  `json:"id"`
		OwnerEmployeeID int64  `json:"owner_employee_id"`
		OwnerEmail      string `json:"owner_email"`
	}
	accBody := postJSON(t, srv.URL+"/api/v1/admin/accounts", platformToken, map[string]any{
		"name":             "Integration Fleet",
		"owner_first_name": "Ivy",
		"owner_last_name":  "Owner",
		"owner_email":      ownerEmail,
		"owner_password":   "correct-horse-battery",
	}, http.StatusCreated, &account)
	if bytes.Contains(bytes.ToLower(accBody), []byte("correct-horse-battery")) {
		t.Fatal("POST /admin/accounts echoed the password back")
	}

	// 2. Login as the owner: zero companies, must still receive a token —
	// this is the deadlock the whole project exists to break.
	var login struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, srv.URL+"/auth/login", "", map[string]any{
		"email": ownerEmail, "password": "correct-horse-battery",
	}, http.StatusOK, &login)
	ownerToken := login.AccessToken

	// 3. POST /companies: the owner creates their first company.
	var company struct {
		ID int64 `json:"id"`
	}
	postJSON(t, srv.URL+"/api/v1/companies", ownerToken, map[string]any{
		"name": "Integration HQ", "tax_id": "IT-001", "address": "1 Integration Rd",
	}, http.StatusCreated, &company)

	var companyAccountID int64
	if err := pool.QueryRow(ctx, "SELECT account_id FROM company WHERE id = $1", company.ID).Scan(&companyAccountID); err != nil {
		t.Fatalf("read back created company: %v", err)
	}
	if companyAccountID != account.ID {
		t.Fatalf("company.account_id = %d, want the owner's own account %d", companyAccountID, account.ID)
	}

	// Re-login: default_company_id is only seeded once a company exists, and
	// every route after this needs a company-scoped token.
	postJSON(t, srv.URL+"/auth/login", "", map[string]any{
		"email": ownerEmail, "password": "correct-horse-battery",
	}, http.StatusOK, &login)
	ownerToken = login.AccessToken

	// 4. POST /employees: the regression this whole review wave started
	// from. Before the CreateEmployee fix this always failed — the employee
	// row got a NULL account_id, and the composite FK rejected the
	// membership insert that followed, rolling back the whole transaction.
	employeeEmail := fmt.Sprintf("employee-%d@integration.test", time.Now().UnixNano())
	var employee struct {
		ID int64 `json:"id"`
	}
	postJSON(t, srv.URL+"/api/v1/employees", ownerToken, map[string]any{
		"first_name": "Wes", "last_name": "Worker", "email": employeeEmail,
		"employee_id": "EMP-IT-1", "mobile_phone": "555", "work_phone": "555",
		"job_title": "Technician", "license_class": "N/A", "license_number": "N/A",
		"license_state": "N/A", "street_address": "1 St", "city": "City",
		"region": "R", "postal_code": "00000", "country": "MX",
	}, http.StatusCreated, &employee)

	var employeeAccountID *int64
	if err := pool.QueryRow(ctx, "SELECT account_id FROM employee WHERE id = $1", employee.ID).Scan(&employeeAccountID); err != nil {
		t.Fatalf("read back created employee: %v", err)
	}
	if employeeAccountID == nil {
		t.Fatal("employee.account_id is NULL after POST /employees — the CreateEmployee regression is back")
	}
	if *employeeAccountID != account.ID {
		t.Fatalf("employee.account_id = %d, want the owner's own account %d", *employeeAccountID, account.ID)
	}

	// 5. PUT /admin/employees/{id}/companies with a company id that does not
	// exist must be a real, surfaced error — not the silent no-op the
	// INSERT ... SELECT ... FROM company shape briefly introduced, which
	// would have reported and audited a membership the database never
	// created.
	badBody := putJSON(t, fmt.Sprintf("%s/api/v1/admin/employees/%d/companies", srv.URL, employee.ID), platformToken,
		map[string]any{"company_ids": []int64{999999999}, "default_company_id": 999999999},
		http.StatusUnprocessableEntity, nil)
	if len(badBody) == 0 {
		t.Fatal("expected an error body for a nonexistent company id")
	}

	// ... and with the company that actually exists, it must succeed and the
	// membership must actually be there afterwards.
	putJSON(t, fmt.Sprintf("%s/api/v1/admin/employees/%d/companies", srv.URL, employee.ID), platformToken,
		map[string]any{"company_ids": []int64{company.ID}, "default_company_id": company.ID},
		http.StatusOK, nil)

	var membershipAccountID int64
	err := pool.QueryRow(ctx,
		"SELECT account_id FROM employee_companies WHERE employee_id = $1 AND company_id = $2",
		employee.ID, company.ID,
	).Scan(&membershipAccountID)
	if err != nil {
		t.Fatalf("PUT /admin/employees/{id}/companies reported success but created no membership row: %v", err)
	}
	if membershipAccountID != account.ID {
		t.Fatalf("employee_companies.account_id = %d, want %d", membershipAccountID, account.ID)
	}
}

// --- Subsystem 2: a role is granted per company, not per person. ---
//
// The helpers below provision the shape every test below needs: a client
// account with a logged-in owner. They reuse seedPlatformAdmin/postJSON/
// putJSON/getJSON exactly as TestOnboardingChain does, rather than talking to
// the database directly for anything the API itself is supposed to do.

// provisionAccountOwner runs the same platform-admin provisioning
// TestOnboardingChain's steps 1-2 exercise (POST /admin/accounts, then log
// in) and hands back the owner's identity and a company-less token. The
// caller creates whatever companies the test needs from there.
func provisionAccountOwner(t *testing.T, baseURL, platformToken string) (accountID, ownerEmployeeID int64, ownerEmail, ownerToken string) {
	t.Helper()

	ownerEmail = fmt.Sprintf("owner-%d@integration.test", time.Now().UnixNano())
	var account struct {
		ID              int64 `json:"id"`
		OwnerEmployeeID int64 `json:"owner_employee_id"`
	}
	postJSON(t, baseURL+"/api/v1/admin/accounts", platformToken, map[string]any{
		"name":             "Integration Fleet",
		"owner_first_name": "Ivy",
		"owner_last_name":  "Owner",
		"owner_email":      ownerEmail,
		"owner_password":   "correct-horse-battery",
	}, http.StatusCreated, &account)

	var login struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, baseURL+"/auth/login", "", map[string]any{
		"email": ownerEmail, "password": "correct-horse-battery",
	}, http.StatusOK, &login)

	return account.ID, account.OwnerEmployeeID, ownerEmail, login.AccessToken
}

// createCompany POSTs /api/v1/companies, which requires RequireAccountOwner
// only (not company membership — see registerCompanyRoutes), so the
// company-less token provisionAccountOwner returns works directly, for as
// many companies as the caller creates.
func createCompany(t *testing.T, baseURL, token, name, taxID string) int64 {
	t.Helper()
	var company struct {
		ID int64 `json:"id"`
	}
	postJSON(t, baseURL+"/api/v1/companies", token, map[string]any{
		"name": name, "tax_id": taxID, "address": "1 Integration Rd",
	}, http.StatusCreated, &company)
	return company.ID
}

// switchCompany re-scopes a token via POST /auth/switch-company, which only
// requires Auth (see router.go) — no prior company membership on the
// presented token — so it works on the company-less token login returns
// before the caller has created or joined anything.
func switchCompany(t *testing.T, baseURL, token string, companyID int64) string {
	t.Helper()
	var pair struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, baseURL+"/auth/switch-company", token, map[string]any{"company_id": companyID}, http.StatusOK, &pair)
	return pair.AccessToken
}

// roleIDByName finds a role by name among a company's roles (GET
// /api/v1/roles, RequireAdminRole-gated). bootstrap.Company seeds exactly two
// roles per company, well inside the default page size, so no pagination is
// needed to find one by name.
func roleIDByName(t *testing.T, baseURL, token, name string) int64 {
	t.Helper()
	var page struct {
		Data []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	getJSON(t, baseURL+"/api/v1/roles", token, http.StatusOK, &page)
	for _, r := range page.Data {
		if r.Name == name {
			return r.ID
		}
	}
	t.Fatalf("role %q not found among %d roles", name, len(page.Data))
	return 0
}

// createDirectorCandidate POSTs /api/v1/employees with callerToken (which
// must be scoped to homeCompany): the route grants exactly one membership —
// the caller's own company — with no role (see employee.go's Create), which
// is the starting point every test below grants roles on top of.
func createDirectorCandidate(t *testing.T, baseURL, callerToken, emailLocalPart string) int64 {
	t.Helper()
	ts := time.Now().UnixNano()
	var employee struct {
		ID int64 `json:"id"`
	}
	postJSON(t, baseURL+"/api/v1/employees", callerToken, map[string]any{
		"first_name": "Dana", "last_name": "Director",
		"email":        fmt.Sprintf("%s-%d@integration.test", emailLocalPart, ts),
		"employee_id":  fmt.Sprintf("EMP-%s-%d", emailLocalPart, ts),
		"mobile_phone": "555", "work_phone": "555",
		"job_title": "Director", "license_class": "N/A", "license_number": "N/A",
		"license_state": "N/A", "street_address": "1 St", "city": "City",
		"region": "R", "postal_code": "00000", "country": "MX",
	}, http.StatusCreated, &employee)
	return employee.ID
}

// setPasswordAndLogIn provisions credentials via POST
// /employees/{id}/set-password (adminToken must hold an admin role in the
// employee's company) and logs in as that employee over HTTP, exactly as a
// real client would, rather than reading password_hash out of the database.
func setPasswordAndLogIn(t *testing.T, baseURL, adminToken string, employeeID int64, email string) string {
	t.Helper()
	postJSON(t, fmt.Sprintf("%s/api/v1/employees/%d/set-password", baseURL, employeeID), adminToken,
		map[string]any{"password": "correct-horse-battery"}, http.StatusNoContent, nil)

	var login struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, baseURL+"/auth/login", "", map[string]any{
		"email": email, "password": "correct-horse-battery",
	}, http.StatusOK, &login)
	return login.AccessToken
}

// The whole point of subsystem 2: one person, two companies, different
// powers. This is the assertion the whole subsystem exists to deliver — if
// this fails, report it exactly rather than adjusting it.
func TestDirectorHasDifferentPowersPerCompany(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	router := newIntegrationRouter(pool)
	srv := httptest.NewServer(router)
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)

	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerToken, "Company A", fmt.Sprintf("TAX-DIFF-A-%d", ts))
	companyB := createCompany(t, srv.URL, ownerToken, "Company B", fmt.Sprintf("TAX-DIFF-B-%d", ts))

	// bootstrap.Company grants the creating owner an admin membership in
	// every company they create, in both A and B, so the owner can list A's
	// seeded role and create a role scoped to B.
	ownerTokenA := switchCompany(t, srv.URL, ownerToken, companyA)
	roleAdminA := roleIDByName(t, srv.URL, ownerTokenA, "Administrador")

	ownerTokenB := switchCompany(t, srv.URL, ownerToken, companyB)
	var readOnlyRole struct {
		ID int64 `json:"id"`
	}
	postJSON(t, srv.URL+"/api/v1/roles", ownerTokenB, map[string]any{
		"name": "Asset Reader", "is_admin": false,
		"permissions": map[string]any{"assets": map[string]any{"read": true}},
	}, http.StatusCreated, &readOnlyRole)

	employeeID := createDirectorCandidate(t, srv.URL, ownerTokenA, "director")

	// Grant: admin in A, read-only in B, in one call.
	putJSON(t, fmt.Sprintf("%s/api/v1/account/employees/%d/companies", srv.URL, employeeID), ownerToken,
		map[string]any{"grants": []map[string]any{
			{"company_id": companyA, "role_id": roleAdminA},
			{"company_id": companyB, "role_id": readOnlyRole.ID},
		}}, http.StatusOK, nil)

	// createDirectorCandidate picked the employee's email internally; fetch it
	// back from the account listing rather than duplicating that logic here.
	var people []struct {
		EmployeeID int64  `json:"employee_id"`
		Email      string `json:"email"`
	}
	getJSON(t, srv.URL+"/api/v1/account/employees", ownerToken, http.StatusOK, &people)
	var employeeEmail string
	for _, p := range people {
		if p.EmployeeID == employeeID {
			employeeEmail = p.Email
		}
	}
	if employeeEmail == "" {
		t.Fatalf("director employee %d missing from account listing", employeeID)
	}

	directorToken := setPasswordAndLogIn(t, srv.URL, ownerTokenA, employeeID, employeeEmail)
	directorTokenA := switchCompany(t, srv.URL, directorToken, companyA)

	// switch to A: POST /api/v1/assets -> 201.
	postJSON(t, srv.URL+"/api/v1/assets", directorTokenA,
		map[string]any{"name": "Truck A", "vin_sn": fmt.Sprintf("VIN-A-%d", ts)}, http.StatusCreated, nil)

	directorTokenB := switchCompany(t, srv.URL, directorTokenA, companyB)

	// switch to B: POST /api/v1/assets -> 403.
	postJSON(t, srv.URL+"/api/v1/assets", directorTokenB,
		map[string]any{"name": "Truck B", "vin_sn": fmt.Sprintf("VIN-B-%d", ts)}, http.StatusForbidden, nil)

	// switch to B: GET /api/v1/assets -> 200.
	getJSON(t, srv.URL+"/api/v1/assets", directorTokenB, http.StatusOK, nil)
}

// An owner who strips their own role must not lose control of their company.
func TestOwnerKeepsAdminAfterLosingTheirRole(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	router := newIntegrationRouter(pool)
	srv := httptest.NewServer(router)
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, ownerID, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)

	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerToken, "Company A", fmt.Sprintf("TAX-OWN-%d", ts))
	ownerTokenA := switchCompany(t, srv.URL, ownerToken, companyA)

	// PUT /account/employees/{ownID}/companies granting their own company
	// with role_id: null.
	putJSON(t, fmt.Sprintf("%s/api/v1/account/employees/%d/companies", srv.URL, ownerID), ownerToken,
		map[string]any{"grants": []map[string]any{
			{"company_id": companyA, "role_id": nil},
		}}, http.StatusOK, nil)

	// Identity is resolved fresh from the database on every request (see
	// middleware.RequireIdentity), so the still-valid, already-issued token
	// is enough to prove the backstop: ownership, not the stripped role,
	// must still back is_admin.
	postJSON(t, srv.URL+"/api/v1/assets", ownerTokenA,
		map[string]any{"name": "Truck", "vin_sn": fmt.Sprintf("VIN-%d", ts)}, http.StatusCreated, nil)
}

// A role from another company must be refused at the API, not just the
// database.
func TestGrantingARoleFromAnotherCompanyIsRefused(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	router := newIntegrationRouter(pool)
	srv := httptest.NewServer(router)
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)

	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerToken, "Company A", fmt.Sprintf("TAX-XA-%d", ts))
	companyB := createCompany(t, srv.URL, ownerToken, "Company B", fmt.Sprintf("TAX-XB-%d", ts))

	ownerTokenA := switchCompany(t, srv.URL, ownerToken, companyA)
	roleAdminA := roleIDByName(t, srv.URL, ownerTokenA, "Administrador")

	employeeID := createDirectorCandidate(t, srv.URL, ownerTokenA, "cross")

	// PUT /account/employees/{id}/companies with company B and a role
	// belonging to company A -> 422 naming role_id.
	body := putJSON(t, fmt.Sprintf("%s/api/v1/account/employees/%d/companies", srv.URL, employeeID), ownerToken,
		map[string]any{"grants": []map[string]any{
			{"company_id": companyB, "role_id": roleAdminA},
		}}, http.StatusUnprocessableEntity, nil)

	if !bytes.Contains(body, []byte("role_id")) {
		t.Fatalf("expected the 422 body to name role_id, got: %s", body)
	}
}

// The database half of the same guarantee: a role from another company cannot
// be attached "by any means, including direct SQL" (spec, success criteria).
// Every assertion names fk_ec_role_company rather than accepting any error,
// because uq_employee_companies can fire first on an insert and make a
// wrong-company write fail for the wrong reason.
func TestWrongCompanyRoleIsRefusedByTheDatabase(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	router := newIntegrationRouter(pool)
	srv := httptest.NewServer(router)
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)

	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerToken, "Company A", fmt.Sprintf("TAX-FK-A-%d", ts))
	companyB := createCompany(t, srv.URL, ownerToken, "Company B", fmt.Sprintf("TAX-FK-B-%d", ts))

	ownerTokenA := switchCompany(t, srv.URL, ownerToken, companyA)
	roleAdminA := roleIDByName(t, srv.URL, ownerTokenA, "Administrador")
	ownerTokenB := switchCompany(t, srv.URL, ownerToken, companyB)
	roleAdminB := roleIDByName(t, srv.URL, ownerTokenB, "Administrador")

	// A member of A only, so an insert into B cannot collide with an existing
	// membership.
	employeeID := createDirectorCandidate(t, srv.URL, ownerTokenA, "fk")
	var accountID int64
	if err := pool.QueryRow(ctx, "SELECT account_id FROM employee_companies WHERE employee_id = $1 AND company_id = $2",
		employeeID, companyA).Scan(&accountID); err != nil {
		t.Fatalf("read the employee's membership in A: %v", err)
	}

	requireFKViolation := func(what string, err error) {
		t.Helper()
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			t.Fatalf("%s: want a fk_ec_role_company violation, got %v", what, err)
		}
		if pgErr.ConstraintName != "fk_ec_role_company" {
			t.Fatalf("%s: refused by %q (%s), want fk_ec_role_company", what, pgErr.ConstraintName, pgErr.Message)
		}
	}

	// Insert: a new membership in B carrying A's role.
	_, err := pool.Exec(ctx,
		"INSERT INTO employee_companies (employee_id, company_id, account_id, role_id) VALUES ($1, $2, $3, $4)",
		employeeID, companyB, accountID, roleAdminA)
	requireFKViolation("insert a membership in B with A's role", err)

	// Update: the existing membership in A re-pointed at B's role.
	_, err = pool.Exec(ctx,
		"UPDATE employee_companies SET role_id = $1 WHERE employee_id = $2 AND company_id = $3",
		roleAdminB, employeeID, companyA)
	requireFKViolation("re-point the membership in A at B's role", err)

	// Control: the identical insert with B's own role succeeds, so the refusal
	// above came from the role alone and not from anything else in the row.
	if _, err := pool.Exec(ctx,
		"INSERT INTO employee_companies (employee_id, company_id, account_id, role_id) VALUES ($1, $2, $3, $4)",
		employeeID, companyB, accountID, roleAdminB); err != nil {
		t.Fatalf("control insert with B's own role was refused, so the assertions above prove nothing: %v", err)
	}
}

// A membership with no role is a real, legitimate state — associated with
// the company, permitted nothing in it — not an oversight. Every module gate
// must refuse, including a plain read.
func TestMembershipWithNoRoleGrantsNothing(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	router := newIntegrationRouter(pool)
	srv := httptest.NewServer(router)
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)

	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerToken, "Company A", fmt.Sprintf("TAX-NR-%d", ts))
	ownerTokenA := switchCompany(t, srv.URL, ownerToken, companyA)

	employeeID := createDirectorCandidate(t, srv.URL, ownerTokenA, "norole")

	var people []struct {
		EmployeeID int64  `json:"employee_id"`
		Email      string `json:"email"`
	}
	getJSON(t, srv.URL+"/api/v1/account/employees", ownerToken, http.StatusOK, &people)
	var employeeEmail string
	for _, p := range people {
		if p.EmployeeID == employeeID {
			employeeEmail = p.Email
		}
	}
	if employeeEmail == "" {
		t.Fatalf("employee %d missing from account listing", employeeID)
	}

	// No PUT /account/employees/.../companies call: the employee keeps
	// exactly the no-role membership POST /employees granted it.
	noRoleToken := setPasswordAndLogIn(t, srv.URL, ownerTokenA, employeeID, employeeEmail)

	postJSON(t, srv.URL+"/api/v1/assets", noRoleToken,
		map[string]any{"name": "Truck", "vin_sn": fmt.Sprintf("VIN-%d", ts)}, http.StatusForbidden, nil)
	getJSON(t, srv.URL+"/api/v1/assets", noRoleToken, http.StatusForbidden, nil)
}

// The empty-grant revoke-all path on PUT /account/employees/{id}/companies.
// RevokeMembershipsNotIn is NOT (company_id = ANY($2::bigint[])); a nil
// []int64 encodes as SQL NULL, under which that predicate matches zero rows,
// so revoke-all would silently revoke nothing. account_employee.go's
// ReplaceCompanies defends with make([]int64, 0, len(req.Grants)), which
// stays a non-nil empty array when the grant list is empty — this proves
// that defence actually revokes every membership.
func TestEmptyGrantListRevokesAllMemberships(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	router := newIntegrationRouter(pool)
	srv := httptest.NewServer(router)
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)

	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerToken, "Company A", fmt.Sprintf("TAX-RV-%d", ts))
	ownerTokenA := switchCompany(t, srv.URL, ownerToken, companyA)

	employeeID := createDirectorCandidate(t, srv.URL, ownerTokenA, "revoke")

	var before int64
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM employee_companies WHERE employee_id = $1", employeeID).Scan(&before); err != nil {
		t.Fatalf("count memberships before revoke: %v", err)
	}
	if before == 0 {
		t.Fatal("employee has no membership to revoke — test setup is broken")
	}

	putJSON(t, fmt.Sprintf("%s/api/v1/account/employees/%d/companies", srv.URL, employeeID), ownerToken,
		map[string]any{"grants": []map[string]any{}}, http.StatusOK, nil)

	var after int64
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM employee_companies WHERE employee_id = $1", employeeID).Scan(&after); err != nil {
		t.Fatalf("count memberships after revoke: %v", err)
	}
	if after != 0 {
		t.Fatalf("PUT with an empty grants list left %d membership(s) behind, want 0 "+
			"(a nil []int64 encodes as SQL NULL, under which RevokeMembershipsNotIn's "+
			"NOT (company_id = ANY(NULL)) matches nothing)", after)
	}
}
