package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
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

// assetListRow is an asset plus its 1:1 subtype columns. The embedded row must
// stay first: the projection is built from it, in field order.
type assetListRow struct {
	gen.Asset
	Operator              *string
	TrailerType           *string
	TrailerClassification *string
	TrailerSize           *string
}

// Assets godoc
//
//	@Summary		List assets
//	@Description	Includes the vehicle operator and trailer subtype columns, so a registry (including its DRY_VAN, CONTAINER, CHASSIS and DOLLY tabs) can be drawn from one paginated request.
//	@Tags			assets
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q				query		string	false	"Search name, VIN/serial and license plate"
//	@Param			vehicle_type	query		string	false	"Filter by vehicle type"
//	@Param			status_id		query		int		false	"Filter by status"
//	@Param			order			query		string	false	"name, vin_sn, vehicle_type, updated_at, id (prefix with - for descending)"
//	@Param			limit			query		int		false	"Page size"
//	@Param			offset			query		int		false	"Offset"
//	@Success		200				{object}	dto.AssetPage
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		403				{object}	dto.ErrorResponse
//	@Router			/api/v1/assets [get]
func (h *FilteredListHandler) Assets(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("a.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	if vehicleType := queryStr(c, "vehicle_type"); vehicleType != nil {
		where.Add("a.vehicle_type", filter.Eq, *vehicleType)
	}
	statusID, err := queryInt64(c, "status_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if statusID != nil {
		where.Add("a.status_id", filter.Eq, *statusID)
	}
	if q := queryStr(c, "q"); q != nil {
		pattern := "%" + *q + "%"
		where.Raw("(a.name ILIKE ? OR a.vin_sn ILIKE ? OR a.license_plate ILIKE ?)", pattern, pattern, pattern)
	}

	spec := newListSpec[assetListRow]("asset", "a").
		join("LEFT JOIN vehicle v ON v.asset_id = a.id", "LEFT JOIN trailer t ON t.asset_id = a.id").
		selecting("v.operator", "t.trailer_type", "t.classification", "t.size").
		filter(where).
		orderBy(orderClause(c, map[string]string{
			"name":         "a.name",
			"vin_sn":       "a.vin_sn",
			"vehicle_type": "a.vehicle_type",
			"updated_at":   "a.updated_at",
			"id":           "a.id",
		}, []filter.Sort{{Column: "a.name"}}, "a.id"))

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, func(r assetListRow) dto.AssetResponse {
		out := toAssetResponse(r.Asset)
		out.Operator = r.Operator
		out.TrailerType = r.TrailerType
		out.TrailerClassification = r.TrailerClassification
		out.TrailerSize = r.TrailerSize
		return out
	})
}

// assetTrailerAssignmentListRow is an assignment plus the trailer's name.
type assetTrailerAssignmentListRow struct {
	gen.AssetTrailerAssignment
	TrailerName string
}

// AssetTrailerAssignments godoc
//
//	@Summary		List trailer assignments across the fleet
//	@Description	Answers "what is towing this trailer?" without draining every asset's nested assignments. A partial unique index allows at most one active assignment per trailer, so trailer_id with is_active=true identifies the current one unambiguously.
//	@Tags			asset-trailer-assignments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			trailer_id	query		int		false	"Filter by trailer"
//	@Param			asset_id	query		int		false	"Filter by towing asset"
//	@Param			is_active	query		bool	false	"Filter by active state"
//	@Param			limit		query		int		false	"Page size"
//	@Param			offset		query		int		false	"Offset"
//	@Success		200			{object}	dto.AssetTrailerAssignmentPage
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Router			/api/v1/asset-trailer-assignments [get]
func (h *FilteredListHandler) AssetTrailerAssignments(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	// Scope follows the towing asset; the trailer is joined only for its name.
	where := filter.NewWhere(1).Add("owner.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	trailerID, err := queryInt64(c, "trailer_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if trailerID != nil {
		where.Add("ata.trailer_id", filter.Eq, *trailerID)
	}
	assetID, err := queryInt64(c, "asset_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if assetID != nil {
		where.Add("ata.asset_id", filter.Eq, *assetID)
	}
	isActive, err := queryBool(c, "is_active")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if isActive != nil {
		where.Add("ata.is_active", filter.Eq, *isActive)
	}

	spec := newListSpec[assetTrailerAssignmentListRow]("asset_trailer_assignment", "ata").
		join("JOIN asset owner ON owner.id = ata.asset_id", "JOIN asset t ON t.id = ata.trailer_id").
		selecting("t.name").
		filter(where).
		orderBy("ata.assigned_date DESC, ata.id DESC")

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, func(r assetTrailerAssignmentListRow) dto.AssetTrailerAssignmentResponse {
		return dto.AssetTrailerAssignmentResponse{
			ID:             r.ID,
			AssetID:        r.AssetID,
			TrailerID:      r.TrailerID,
			Position:       r.Position,
			AssignedDate:   r.AssignedDate,
			UnassignedDate: r.UnassignedDate,
			AssignedByID:   r.AssignedByID,
			IsActive:       r.IsActive,
			Notes:          r.Notes,
			TrailerName:    r.TrailerName,
		}
	})
}

// Vocabularies accepted by the state filters, kept beside the routes that
// validate them. Ported from the Django models the schema came from.
var issueStates = []string{"OPEN", "RESOLVED", "CLOSED"}
