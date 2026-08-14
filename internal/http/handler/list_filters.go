package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/filter"
	"fleet/internal/platform/paginate"
)

// FilteredListHandler serves the collections whose filters vary per request.
// They cannot come from sqlc, which generates one fixed statement per query, so
// they are assembled through listSpec instead. Everything else about them —
// tenant scope, page envelope, permission gate — is unchanged; only the List
// verb is taken over, via registerCrudWithList.
type FilteredListHandler struct {
	pool *pgxpool.Pool
}

func NewFilteredListHandler(pool *pgxpool.Pool) *FilteredListHandler {
	return &FilteredListHandler{pool: pool}
}

// renderPage maps rows through their DTO conversion and writes the page.
func renderPage[R any, T any](c *gin.Context, rows []R, total int64, p paginate.Params, convert func(R) T) {
	out := make([]T, 0, len(rows))
	for _, r := range rows {
		out = append(out, convert(r))
	}
	c.JSON(http.StatusOK, paginate.NewPage(out, total, p))
}

// WorkOrders godoc
//
//	@Summary	List work orders
//	@Tags		work-orders
//	@Produce	json
//	@Security	BearerAuth
//	@Param		asset_id	query		int		false	"Filter by asset"
//	@Param		status_id	query		int		false	"Filter by status"
//	@Param		order		query		string	false	"issued_at, id (prefix with - for descending)"
//	@Param		limit		query		int		false	"Page size"
//	@Param		offset		query		int		false	"Offset"
//	@Success	200			{object}	dto.WorkOrderPage
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Router		/api/v1/work-orders [get]
func (h *FilteredListHandler) WorkOrders(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("wo.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	assetID, err := queryInt64(c, "asset_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if assetID != nil {
		where.Add("wo.asset_id", filter.Eq, *assetID)
	}
	statusID, err := queryInt64(c, "status_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if statusID != nil {
		where.Add("wo.status_id", filter.Eq, *statusID)
	}

	spec := newListSpec[gen.WorkOrder]("work_order", "wo").
		filter(where).
		orderBy(orderClause(c,
			map[string]string{"issued_at": "wo.issued_at", "id": "wo.id"},
			[]filter.Sort{{Column: "wo.issued_at", Desc: true}}, "wo.id"))

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toWorkOrderResponse)
}

// Issues godoc
//
//	@Summary	List issues
//	@Tags		issues
//	@Produce	json
//	@Security	BearerAuth
//	@Param		asset_id	query		int		false	"Filter by asset"
//	@Param		state		query		string	false	"OPEN, RESOLVED or CLOSED"
//	@Param		order		query		string	false	"created_at, due_date, state, id (prefix with - for descending)"
//	@Param		limit		query		int		false	"Page size"
//	@Param		offset		query		int		false	"Offset"
//	@Success	200			{object}	dto.IssuePage
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Router		/api/v1/issues [get]
func (h *FilteredListHandler) Issues(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("i.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	assetID, err := queryInt64(c, "asset_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if assetID != nil {
		where.Add("i.asset_id", filter.Eq, *assetID)
	}
	state, err := queryEnum(c, "state", issueStates...)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if state != nil {
		where.Add("i.state", filter.Eq, *state)
	}

	spec := newListSpec[gen.Issue]("issue", "i").
		filter(where).
		orderBy(orderClause(c,
			map[string]string{"created_at": "i.created_at", "due_date": "i.due_date", "state": "i.state", "id": "i.id"},
			[]filter.Sort{{Column: "i.created_at", Desc: true}}, "i.id"))

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toIssueResponse)
}

// InventoryJournalEntries godoc
//
//	@Summary		List inventory journal entries
//	@Description	The ledger grows forever, so it is filtered in SQL and always returned newest first.
//	@Tags			inventory-journal-entries
//	@Produce		json
//	@Security		BearerAuth
//	@Param			part_id			query		int		false	"Filter by part"
//	@Param			created_from	query		string	false	"Inclusive lower bound (RFC3339 or YYYY-MM-DD)"
//	@Param			created_to		query		string	false	"Inclusive upper bound (RFC3339 or YYYY-MM-DD)"
//	@Param			limit			query		int		false	"Page size"
//	@Param			offset			query		int		false	"Offset"
//	@Success		200				{object}	dto.InventoryJournalEntryPage
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		403				{object}	dto.ErrorResponse
//	@Router			/api/v1/inventory-journal-entries [get]
func (h *FilteredListHandler) InventoryJournalEntries(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("ije.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	partID, err := queryInt64(c, "part_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if partID != nil {
		where.Add("ije.part_id", filter.Eq, *partID)
	}
	from, err := queryTime(c, "created_from")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if from != nil {
		where.Add("ije.created_at", filter.Gte, *from)
	}
	to, err := queryTime(c, "created_to")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if to != nil {
		where.Add("ije.created_at", filter.Lte, *to)
	}

	spec := newListSpec[gen.InventoryJournalEntry]("inventory_journal_entry", "ije").
		filter(where).
		orderBy("ije.created_at DESC, ije.id DESC")

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toInventoryJournalEntryResponse)
}

// Vocabularies accepted by the state filters, kept beside the routes that
// validate them. Ported from the Django models the schema came from.
var issueStates = []string{"OPEN", "RESOLVED", "CLOSED"}
