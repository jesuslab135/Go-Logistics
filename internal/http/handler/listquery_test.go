package handler

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"fleet/internal/db/gen"
	"fleet/internal/platform/filter"
	"fleet/internal/platform/paginate"
)

func TestSnakeCase(t *testing.T) {
	tests := []struct{ field, want string }{
		{"ID", "id"},
		{"CompanyID", "company_id"},
		{"PositionCode", "position_code"},
		{"TreadDepth32nds", "tread_depth_32nds"},
		{"CurrentTreadDepth32nds", "current_tread_depth_32nds"},
		{"Classification2", "classification_2"},
		{"Psi", "psi"},
		{"PerformedByID", "performed_by_id"},
		{"IsPrimary", "is_primary"},
		{"UpdatedAt", "updated_at"},
	}

	for _, tt := range tests {
		if got := snakeCase(tt.field); got != tt.want {
			t.Errorf("snakeCase(%q) = %q, want %q", tt.field, got, tt.want)
		}
	}
}

// The projection is derived from the sqlc row struct, so this asserts the
// derivation still reproduces the real table. If sqlc's naming ever diverges
// from snakeCase, this fails here rather than at runtime with a mis-scanned row.
func TestTableColumnsMatchSchema(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
		want []string
	}{
		{
			name: "tire",
			typ:  reflect.TypeFor[gen.Tire](),
			want: []string{
				"id", "company_id", "tire_identification_number", "tire_model_id", "status",
				"current_tread_depth_32nds", "current_psi", "total_miles", "current_vehicle_id",
				"current_position_code", "purchase_date", "purchase_cost", "vendor_id", "created_at",
			},
		},
		{
			name: "tire_mount_log",
			typ:  reflect.TypeFor[gen.TireMountLog](),
			want: []string{
				"id", "tire_id", "vehicle_id", "position_code", "event_type", "event_date",
				"odometer", "tread_depth_32nds", "psi", "performed_by_id", "reason",
			},
		},
		{
			name: "revoked_token",
			typ:  reflect.TypeFor[gen.RevokedToken](),
			want: []string{"jti", "employee_id", "expires_at"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tableColumns(tt.typ); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tableColumns =\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func TestBuildNamesColumnsExplicitly(t *testing.T) {
	spec := newListSpec[gen.RevokedToken]("revoked_token", "rt").
		filter(filter.NewWhere(1).Add("rt.employee_id", filter.Eq, int64(3))).
		orderBy("rt.expires_at DESC, rt.jti DESC")

	listSQL, listArgs, countSQL, countArgs := spec.build(paginate.Params{Limit: 25, Offset: 0})

	const wantList = "SELECT rt.jti, rt.employee_id, rt.expires_at FROM revoked_token rt " +
		"WHERE rt.employee_id = $1 ORDER BY rt.expires_at DESC, rt.jti DESC LIMIT $2 OFFSET $3"
	if listSQL != wantList {
		t.Errorf("list SQL =\n%q\nwant\n%q", listSQL, wantList)
	}
	if !reflect.DeepEqual(listArgs, []any{int64(3), 25, 0}) {
		t.Errorf("list args = %#v", listArgs)
	}

	// The count carries the same WHERE but neither ORDER BY nor LIMIT.
	const wantCount = "SELECT count(*) FROM revoked_token rt WHERE rt.employee_id = $1"
	if countSQL != wantCount {
		t.Errorf("count SQL =\n%q\nwant\n%q", countSQL, wantCount)
	}
	if !reflect.DeepEqual(countArgs, []any{int64(3)}) {
		t.Errorf("count args = %#v", countArgs)
	}
}

func TestBuildWithoutConditions(t *testing.T) {
	spec := newListSpec[gen.RevokedToken]("revoked_token", "rt")

	listSQL, listArgs, countSQL, _ := spec.build(paginate.Params{Limit: 10, Offset: 20})

	if !strings.HasSuffix(listSQL, "FROM revoked_token rt LIMIT $1 OFFSET $2") {
		t.Errorf("list SQL = %q", listSQL)
	}
	if !reflect.DeepEqual(listArgs, []any{10, 20}) {
		t.Errorf("list args = %#v", listArgs)
	}
	if countSQL != "SELECT count(*) FROM revoked_token rt" {
		t.Errorf("count SQL = %q", countSQL)
	}
}

type testWrapperRow struct {
	gen.RevokedToken
	EmployeeName *string
}

func TestBuildWrapperRowAppendsExtras(t *testing.T) {
	spec := newListSpec[testWrapperRow]("revoked_token", "rt").
		join("JOIN employee e ON e.id = rt.employee_id").
		selecting("e.first_name")

	listSQL, _, countSQL, _ := spec.build(paginate.Params{Limit: 5, Offset: 0})

	const wantSelect = "SELECT rt.jti, rt.employee_id, rt.expires_at, e.first_name " +
		"FROM revoked_token rt JOIN employee e ON e.id = rt.employee_id"
	if !strings.HasPrefix(listSQL, wantSelect) {
		t.Errorf("list SQL =\n%q\nwant prefix\n%q", listSQL, wantSelect)
	}
	if !strings.Contains(countSQL, "JOIN employee e ON e.id = rt.employee_id") {
		t.Errorf("count SQL lost the join: %q", countSQL)
	}
}

func TestBuildCountDistinct(t *testing.T) {
	spec := newListSpec[gen.Tire]("tire", "ti").distinct()

	_, _, countSQL, _ := spec.build(paginate.Params{Limit: 5, Offset: 0})

	if !strings.HasPrefix(countSQL, "SELECT count(DISTINCT ti.id)") {
		t.Errorf("count SQL = %q", countSQL)
	}
}

// A projection that does not line up with the row struct would scan columns
// into the wrong fields, so it must fail at build time.
func TestBuildRejectsMismatchedProjection(t *testing.T) {
	tests := []struct {
		name string
		run  func()
	}{
		{
			name: "wrapper field without an expression",
			run: func() {
				newListSpec[testWrapperRow]("revoked_token", "rt").build(paginate.Params{})
			},
		},
		{
			name: "extra expression without a field",
			run: func() {
				newListSpec[gen.RevokedToken]("revoked_token", "rt").selecting("e.first_name").build(paginate.Params{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected a panic")
				}
			}()
			tt.run()
		})
	}
}

func TestSpecRejectsUnsafeFragments(t *testing.T) {
	tests := []struct {
		name string
		run  func()
	}{
		{"non-identifier table", func() { newListSpec[gen.Tire]("tire; DROP TABLE tire", "ti") }},
		{"non-identifier alias", func() { newListSpec[gen.Tire]("tire", "ti OR 1=1") }},
		{"terminator in join", func() { newListSpec[gen.Tire]("tire", "ti").join("JOIN x ON true; DELETE FROM tire") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected a panic")
				}
			}()
			tt.run()
		})
	}
}

func testContext(t *testing.T, query string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/?"+query, nil)
	return c
}

func TestOrderClauseAppendsTieBreaker(t *testing.T) {
	allowed := map[string]string{"name": "a.name", "updated_at": "a.updated_at"}
	fallback := []filter.Sort{{Column: "a.name"}}

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"default", "", "a.name ASC, a.id ASC"},
		{"explicit descending", "order=-updated_at", "a.updated_at DESC, a.id DESC"},
		{"unknown field falls back", "order=hacker", "a.name ASC, a.id ASC"},
		{"injection attempt is dropped", "order=name%3BDROP+TABLE+asset", "a.name ASC, a.id ASC"},
		{"multiple keys", "order=name,-updated_at", "a.name ASC, a.updated_at DESC, a.id DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := orderClause(testContext(t, tt.query), allowed, fallback, "a.id"); got != tt.want {
				t.Errorf("orderClause = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOrderClauseDoesNotMutateFallback(t *testing.T) {
	fallback := []filter.Sort{{Column: "a.name"}}

	orderClause(testContext(t, ""), map[string]string{}, fallback, "a.id")
	orderClause(testContext(t, ""), map[string]string{}, fallback, "a.id")

	if len(fallback) != 1 {
		t.Fatalf("fallback grew to %d entries; the tie-breaker was appended in place", len(fallback))
	}
}

func TestQueryHelpers(t *testing.T) {
	t.Run("absent values yield no filter", func(t *testing.T) {
		c := testContext(t, "")
		if v, err := queryInt64(c, "asset_id"); v != nil || err != nil {
			t.Errorf("queryInt64 = %v, %v", v, err)
		}
		if v := queryStr(c, "q"); v != nil {
			t.Errorf("queryStr = %v", *v)
		}
		if v, err := queryTime(c, "date_from"); v != nil || err != nil {
			t.Errorf("queryTime = %v, %v", v, err)
		}
		if v, err := queryEnum(c, "state", "OPEN"); v != nil || err != nil {
			t.Errorf("queryEnum = %v, %v", v, err)
		}
	})

	t.Run("malformed values are rejected", func(t *testing.T) {
		if _, err := queryInt64(testContext(t, "asset_id=abc"), "asset_id"); err == nil {
			t.Error("queryInt64 accepted a non-numeric id")
		}
		if _, err := queryBool(testContext(t, "is_active=maybe"), "is_active"); err == nil {
			t.Error("queryBool accepted a non-boolean")
		}
		if _, err := queryTime(testContext(t, "date_from=yesterday"), "date_from"); err == nil {
			t.Error("queryTime accepted an unparseable date")
		}
		if _, err := queryEnum(testContext(t, "state=BOGUS"), "state", "OPEN", "CLOSED"); err == nil {
			t.Error("queryEnum accepted a value outside the vocabulary")
		}
	})

	t.Run("valid values parse", func(t *testing.T) {
		v, err := queryInt64(testContext(t, "asset_id=42"), "asset_id")
		if err != nil || v == nil || *v != 42 {
			t.Errorf("queryInt64 = %v, %v", v, err)
		}
		ts, err := queryTime(testContext(t, "date_from=2026-08-14"), "date_from")
		if err != nil || ts == nil || ts.Year() != 2026 {
			t.Errorf("queryTime = %v, %v", ts, err)
		}
		if _, err := queryTime(testContext(t, "date_from=2026-08-14T10:00:00Z"), "date_from"); err != nil {
			t.Errorf("queryTime rejected RFC3339: %v", err)
		}
		e, err := queryEnum(testContext(t, "state=OPEN"), "state", "OPEN", "CLOSED")
		if err != nil || e == nil || *e != "OPEN" {
			t.Errorf("queryEnum = %v, %v", e, err)
		}
	})
}
