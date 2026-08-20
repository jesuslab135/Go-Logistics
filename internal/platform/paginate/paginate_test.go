package paginate

import (
	"math"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func parseQuery(t *testing.T, query string) Params {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?"+query, nil)
	return Parse(c)
}

// Every call site narrows Offset to int32 for Postgres. An offset past that
// range truncates to a negative number, which Postgres rejects with
// invalid_row_count_in_result_offset_clause — a 500 on any list route.
func TestParseClampsOffsetToInt32(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "an ordinary offset is kept", query: "offset=100", want: 100},
		{name: "a zero offset is kept", query: "offset=0", want: 0},
		{name: "an offset past int32 is clamped", query: "offset=3000000000", want: math.MaxInt32},
		{name: "a negative offset is ignored", query: "offset=-5", want: 0},
		{name: "a page far past int32 is clamped", query: "page=999999999&page_size=100", want: math.MaxInt32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseQuery(t, tt.query)
			if got.Offset != tt.want {
				t.Errorf("offset = %d, want %d", got.Offset, tt.want)
			}
			if int(int32(got.Offset)) != got.Offset || got.Offset < 0 {
				t.Errorf("offset %d does not survive the int32 conversion every call site makes", got.Offset)
			}
		})
	}
}
