package paginate

import (
	"math"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	defaultLimit = envInt("PAGE_DEFAULT_SIZE", 25)
	maxLimit     = envInt("PAGE_MAX_SIZE", 100)
)

type Params struct {
	Limit  int
	Offset int
}

// Parse reads pagination from the query string. It accepts either page-based
// (?page=2&page_size=50) or explicit (?limit=50&offset=100) params; explicit
// values win when both are present. Limit is always clamped to [1, PAGE_MAX_SIZE]
// and offset to [0, math.MaxInt32].
func Parse(c *gin.Context) Params {
	limit := clamp(queryInt(c, "page_size", defaultLimit), 1, maxLimit)

	page := max(queryInt(c, "page", 1), 1)
	offset := (page - 1) * limit

	if v, ok := c.GetQuery("limit"); ok {
		if n, err := strconv.Atoi(v); err == nil {
			limit = clamp(n, 1, maxLimit)
		}
	}
	if v, ok := c.GetQuery("offset"); ok {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	// Every call site narrows Offset to int32 for Postgres, so an unclamped
	// value truncates to a negative one and the query fails with
	// invalid_row_count_in_result_offset_clause — a 500 on any list route.
	return Params{Limit: limit, Offset: clamp(offset, 0, math.MaxInt32)}
}

type Page[T any] struct {
	Data    []T   `json:"data"`
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasNext bool  `json:"has_next"`
}

func NewPage[T any](data []T, total int64, p Params) Page[T] {
	if data == nil {
		data = []T{}
	}
	return Page[T]{
		Data:    data,
		Total:   total,
		Limit:   p.Limit,
		Offset:  p.Offset,
		HasNext: int64(p.Offset+len(data)) < total,
	}
}

func queryInt(c *gin.Context, key string, fallback int) int {
	if v, ok := c.GetQuery(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func clamp(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}

func envInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
