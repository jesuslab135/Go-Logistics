package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/paginate"
)

func specForQuery(t *testing.T, rawQuery string) (string, []any) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?"+rawQuery, nil)
	spec, err := inventoryJournalListSpec(c, 7)
	if err != nil {
		t.Fatalf("build spec: %v", err)
	}
	sql, args, _, _ := spec.build(paginate.Params{Limit: 25, Offset: 0})
	return sql, args
}

func TestInventoryJournalAdjustmentTypeFilter(t *testing.T) {
	sql, _ := specForQuery(t, "adjustment_type=transfer")
	if !strings.Contains(sql, "ije.adjustment_type =") {
		t.Errorf("expected adjustment_type predicate, got:\n%s", sql)
	}
}

func TestInventoryJournalLocationFilterJoinsPartInventory(t *testing.T) {
	sql, _ := specForQuery(t, "location_id=5")
	if !strings.Contains(sql, "JOIN part_inventory pi ON pi.id = ije.part_location_detail_id") {
		t.Errorf("expected part_inventory join, got:\n%s", sql)
	}
	if !strings.Contains(sql, "pi.location_id =") {
		t.Errorf("expected pi.location_id predicate, got:\n%s", sql)
	}
}

func TestInventoryJournalKeepsTenantAndPartFilters(t *testing.T) {
	sql, _ := specForQuery(t, "part_id=3&location_id=5")
	for _, want := range []string{"ije.company_id =", "ije.part_id =", "pi.location_id ="} {
		if !strings.Contains(sql, want) {
			t.Errorf("missing %q in:\n%s", want, sql)
		}
	}
}
