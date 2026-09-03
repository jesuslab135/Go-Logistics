package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"fleet/internal/http/dto"
	"fleet/internal/platform/filter"
)

// getCtx builds a *gin.Context for a GET request carrying rawQuery, for tests
// that need to call a *FacetWhere builder directly (they take a *gin.Context,
// not a spec function).
func getCtx(t *testing.T, rawQuery string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?"+rawQuery, nil)
	return c
}

func TestFacetQueryGroupsWithLabelJoin(t *testing.T) {
	where := filter.NewWhere(1).Add("wo.company_id", filter.Eq, int64(7))
	sql, args := facetQuery(
		"FROM work_order wo JOIN work_order_status s ON s.id = wo.status_id",
		where, "wo.status_id", "s.name")
	for _, want := range []string{
		"SELECT wo.status_id", "s.name", "count(*)",
		"FROM work_order wo JOIN work_order_status s",
		"wo.company_id =", "GROUP BY wo.status_id, s.name",
	} {
		if !strings.Contains(sql, want) {
			t.Errorf("missing %q in:\n%s", want, sql)
		}
	}
	if len(args) != 1 || args[0].(int64) != 7 {
		t.Errorf("args = %v, want [7]", args)
	}
}

func TestFacetCountFormatsValueAndLabel(t *testing.T) {
	cases := []struct {
		name  string
		value any
		label *string
		want  dto.FacetCount
	}{
		{
			name:  "nil value and nil label becomes the no-status sentinel",
			value: nil,
			label: nil,
			want:  dto.FacetCount{Value: "", Label: "No status"},
		},
		{
			name:  "non-nil value with nil label defaults the label to the value",
			value: int64(3),
			label: nil,
			want:  dto.FacetCount{Value: "3", Label: "3"},
		},
		{
			name:  "non-nil value and non-nil label keeps both",
			value: int64(3),
			label: strPtr("Active"),
			want:  dto.FacetCount{Value: "3", Label: "Active"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := facetCount(tc.value, tc.label)
			if got.Value != tc.want.Value || got.Label != tc.want.Label {
				t.Errorf("facetCount(%v, %v) = %+v, want Value=%q Label=%q",
					tc.value, tc.label, got, tc.want.Value, tc.want.Label)
			}
		})
	}
}

func strPtr(s string) *string { return &s }

func TestWorkOrderFacetsExcludeStatusDimension(t *testing.T) {
	c := getCtx(t, "asset_id=3&status_id=9")
	h := &FilteredListHandler{}
	where, err := h.workOrderFacetWhere(c, 7)
	if err != nil {
		t.Fatal(err)
	}
	sql := where.SQL()
	if !strings.Contains(sql, "wo.asset_id =") {
		t.Errorf("asset_id filter should apply to facets: %s", sql)
	}
	if strings.Contains(sql, "wo.status_id =") {
		t.Errorf("status_id must be excluded when grouping by status: %s", sql)
	}
}

func TestIssueFacetsExcludeStateDimension(t *testing.T) {
	c := getCtx(t, "asset_id=3&state=OPEN")
	h := &FilteredListHandler{}
	where, err := h.issueFacetWhere(c, 7)
	if err != nil {
		t.Fatal(err)
	}
	sql := where.SQL()
	if !strings.Contains(sql, "i.asset_id =") {
		t.Errorf("asset_id filter should apply to facets: %s", sql)
	}
	if strings.Contains(sql, "i.state =") {
		t.Errorf("state must be excluded when grouping by state: %s", sql)
	}
}

func TestAssetVehicleTypeFacetExcludesVehicleTypeDimension(t *testing.T) {
	c := getCtx(t, "vehicle_type=TRUCK&status_id=2&q=abc")
	h := &FilteredListHandler{}
	where, err := h.assetVehicleTypeFacetWhere(c, c.Request.Context(), 7)
	if err != nil {
		t.Fatal(err)
	}
	sql := where.SQL()
	if strings.Contains(sql, "a.vehicle_type =") {
		t.Errorf("vehicle_type must be excluded when grouping by vehicle_type: %s", sql)
	}
	if !strings.Contains(sql, "a.status_id =") {
		t.Errorf("status_id filter should apply: %s", sql)
	}
	if !strings.Contains(sql, "a.name ILIKE") {
		t.Errorf("q filter should apply: %s", sql)
	}
	if !strings.Contains(sql, "a.archived_at IS NULL") {
		t.Errorf("archived-visibility rule should apply by default: %s", sql)
	}
}

func TestAssetStatusFacetExcludesStatusDimension(t *testing.T) {
	c := getCtx(t, "vehicle_type=TRUCK&status_id=2")
	h := &FilteredListHandler{}
	where, err := h.assetStatusFacetWhere(c, c.Request.Context(), 7)
	if err != nil {
		t.Fatal(err)
	}
	sql := where.SQL()
	if strings.Contains(sql, "a.status_id =") {
		t.Errorf("status_id must be excluded when grouping by status: %s", sql)
	}
	if !strings.Contains(sql, "a.vehicle_type =") {
		t.Errorf("vehicle_type filter should apply: %s", sql)
	}
}

func TestAssetTrailerTypeFacetAppliesAllListFilters(t *testing.T) {
	c := getCtx(t, "vehicle_type=TRAILER&status_id=2")
	h := &FilteredListHandler{}
	where, err := h.assetTrailerTypeFacetWhere(c, c.Request.Context(), 7)
	if err != nil {
		t.Fatal(err)
	}
	sql := where.SQL()
	for _, want := range []string{"a.vehicle_type =", "a.status_id ="} {
		if !strings.Contains(sql, want) {
			t.Errorf("missing %q in:\n%s", want, sql)
		}
	}
}

func TestAssetFacetsQueryShapes(t *testing.T) {
	c := getCtx(t, "")
	h := &FilteredListHandler{}

	vtWhere, err := h.assetVehicleTypeFacetWhere(c, c.Request.Context(), 7)
	if err != nil {
		t.Fatal(err)
	}
	vtSQL, _ := facetQuery("FROM asset a", vtWhere, "a.vehicle_type", "a.vehicle_type")
	if !strings.Contains(vtSQL, "GROUP BY a.vehicle_type, a.vehicle_type") {
		t.Errorf("vehicle_type facet query shape wrong: %s", vtSQL)
	}

	stWhere, err := h.assetStatusFacetWhere(c, c.Request.Context(), 7)
	if err != nil {
		t.Fatal(err)
	}
	// asset.status_id is nullable, so this must be a LEFT (not INNER) join: an
	// INNER join would silently drop null-status assets from the facet
	// entirely, breaking the "counts stay exact" guarantee.
	stSQL, _ := facetQuery("FROM asset a LEFT JOIN asset_status st ON st.id = a.status_id", stWhere, "a.status_id", "st.name")
	if !strings.Contains(stSQL, "LEFT JOIN asset_status st ON st.id = a.status_id") ||
		!strings.Contains(stSQL, "GROUP BY a.status_id, st.name") {
		t.Errorf("status facet query shape wrong: %s", stSQL)
	}

	ttWhere, err := h.assetTrailerTypeFacetWhere(c, c.Request.Context(), 7)
	if err != nil {
		t.Fatal(err)
	}
	ttSQL, _ := facetQuery("FROM asset a JOIN trailer t ON t.asset_id = a.id", ttWhere, "t.trailer_type", "t.trailer_type")
	if !strings.Contains(ttSQL, "JOIN trailer t ON t.asset_id = a.id") ||
		!strings.Contains(ttSQL, "GROUP BY t.trailer_type, t.trailer_type") {
		t.Errorf("trailer_type facet query shape wrong: %s", ttSQL)
	}
}
