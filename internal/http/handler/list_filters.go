package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/domain/tire"
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

// PartInventory godoc
//
//	@Summary		List part inventory rows
//	@Description	Flat, filterable view of the rows nested under /parts/{id}/inventory, so a form can offer inventory rows as a foreign-key source without knowing the part up front.
//	@Tags			part-inventory
//	@Produce		json
//	@Security		BearerAuth
//	@Param			part_id	query		int	false	"Filter by part"
//	@Param			limit	query		int	false	"Page size"
//	@Param			offset	query		int	false	"Offset"
//	@Success		200		{object}	dto.PartInventoryPage
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Router			/api/v1/part-inventory [get]
func (h *FilteredListHandler) PartInventory(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("prt.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	partID, err := queryInt64(c, "part_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if partID != nil {
		where.Add("pi.part_id", filter.Eq, *partID)
	}

	spec := newListSpec[gen.PartInventory]("part_inventory", "pi").
		join("JOIN part prt ON prt.id = pi.part_id").
		filter(where).
		orderBy("pi.id ASC")

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toPartInventoryResponse)
}

// PurchaseOrderLineItems godoc
//
//	@Summary		List purchase order line items
//	@Description	Flat, filterable view of the rows nested under /purchase-orders/{id}/line-items, so a form can reference a line item without draining every order.
//	@Tags			purchase-order-line-items
//	@Produce		json
//	@Security		BearerAuth
//	@Param			purchase_order_id	query		int	false	"Filter by purchase order"
//	@Param			part_id				query		int	false	"Filter by part"
//	@Param			limit				query		int	false	"Page size"
//	@Param			offset				query		int	false	"Offset"
//	@Success		200					{object}	dto.PurchaseOrderLineItemPage
//	@Failure		400					{object}	dto.ErrorResponse
//	@Failure		401					{object}	dto.ErrorResponse
//	@Failure		403					{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-order-line-items [get]
func (h *FilteredListHandler) PurchaseOrderLineItems(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("po.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	orderID, err := queryInt64(c, "purchase_order_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if orderID != nil {
		where.Add("li.purchase_order_id", filter.Eq, *orderID)
	}
	partID, err := queryInt64(c, "part_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if partID != nil {
		where.Add("li.part_id", filter.Eq, *partID)
	}

	spec := newListSpec[gen.PurchaseOrderLineItem]("purchase_order_line_item", "li").
		join("JOIN purchase_order po ON po.id = li.purchase_order_id").
		filter(where).
		orderBy("li.position ASC, li.id ASC")

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toPurchaseOrderLineItemResponse)
}

// FuelEntries godoc
//
//	@Summary		List fuel entries across the fleet
//	@Description	Company-wide register, so a cross-fleet view does not need one request per asset.
//	@Tags			fuel-entries
//	@Produce		json
//	@Security		BearerAuth
//	@Param			asset_id	query		int		false	"Filter by asset"
//	@Param			fuel_type	query		string	false	"Filter by fuel type"
//	@Param			date_from	query		string	false	"Inclusive lower bound (RFC3339 or YYYY-MM-DD)"
//	@Param			date_to		query		string	false	"Inclusive upper bound (RFC3339 or YYYY-MM-DD)"
//	@Param			order		query		string	false	"date, odometer, quantity, total_cost, updated_at, id (prefix with - for descending)"
//	@Param			limit		query		int		false	"Page size"
//	@Param			offset		query		int		false	"Offset"
//	@Success		200			{object}	dto.FuelEntryPage
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Router			/api/v1/fuel-entries [get]
func (h *FilteredListHandler) FuelEntries(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	// fuel_entry has no company of its own; it is scoped through its asset.
	where := filter.NewWhere(1).Add("a.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	assetID, err := queryInt64(c, "asset_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if assetID != nil {
		where.Add("fe.asset_id", filter.Eq, *assetID)
	}
	// fuel_type is a free varchar in the schema, so it is matched as given
	// rather than validated against a fixed vocabulary.
	if fuelType := queryStr(c, "fuel_type"); fuelType != nil {
		where.Add("fe.fuel_type", filter.Eq, *fuelType)
	}
	from, err := queryTime(c, "date_from")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if from != nil {
		where.Add("fe.date", filter.Gte, *from)
	}
	to, err := queryTime(c, "date_to")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if to != nil {
		where.Add("fe.date", filter.Lte, *to)
	}

	spec := newListSpec[gen.FuelEntry]("fuel_entry", "fe").
		join("JOIN asset a ON a.id = fe.asset_id").
		filter(where).
		orderBy(orderClause(c, map[string]string{
			"date":       "fe.date",
			"odometer":   "fe.odometer",
			"quantity":   "fe.quantity",
			"total_cost": "fe.total_cost",
			"updated_at": "fe.updated_at",
			"id":         "fe.id",
		}, []filter.Sort{{Column: "fe.date", Desc: true}}, "fe.id"))

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toFuelEntryResponse)
}

// TireMountLogs godoc
//
//	@Summary		List tire movements across the fleet
//	@Description	The mount log is the history of tire movement — installs, dismounts and rotations. It is not a TireInstallation collection: an installation records where a tire is now, a mount log records what happened.
//	@Tags			tire-mount-logs
//	@Produce		json
//	@Security		BearerAuth
//	@Param			event_date_from	query		string	false	"Inclusive lower bound (RFC3339 or YYYY-MM-DD)"
//	@Param			event_date_to	query		string	false	"Inclusive upper bound (RFC3339 or YYYY-MM-DD)"
//	@Param			event_type		query		string	false	"INSTALL, DISMOUNT or ROTATION"
//	@Param			tire_id			query		int		false	"Filter by tire"
//	@Param			vehicle_id		query		int		false	"Filter by vehicle"
//	@Param			position_code	query		string	false	"Filter by wheel position"
//	@Param			performed_by_id	query		int		false	"Filter by the employee who performed it"
//	@Param			limit			query		int		false	"Page size"
//	@Param			offset			query		int		false	"Offset"
//	@Success		200				{object}	dto.TireMountLogPage
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		403				{object}	dto.ErrorResponse
//	@Router			/api/v1/tire-mount-logs [get]
func (h *FilteredListHandler) TireMountLogs(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	// The log has no company of its own; it is scoped through its tire.
	where := filter.NewWhere(1).Add("ti.company_id", filter.Eq, middleware.CompanyFromContext(ctx))

	from, err := queryTime(c, "event_date_from")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if from != nil {
		where.Add("tml.event_date", filter.Gte, *from)
	}
	to, err := queryTime(c, "event_date_to")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if to != nil {
		where.Add("tml.event_date", filter.Lte, *to)
	}
	eventType, err := queryEnum(c, "event_type", tire.EventInstall, tire.EventDismount, tire.EventRotation)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if eventType != nil {
		where.Add("tml.event_type", filter.Eq, *eventType)
	}
	tireID, err := queryInt64(c, "tire_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if tireID != nil {
		where.Add("tml.tire_id", filter.Eq, *tireID)
	}
	vehicleID, err := queryInt64(c, "vehicle_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if vehicleID != nil {
		where.Add("tml.vehicle_id", filter.Eq, *vehicleID)
	}
	if positionCode := queryStr(c, "position_code"); positionCode != nil {
		where.Add("tml.position_code", filter.Eq, *positionCode)
	}
	performedBy, err := queryInt64(c, "performed_by_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if performedBy != nil {
		where.Add("tml.performed_by_id", filter.Eq, *performedBy)
	}

	spec := newListSpec[gen.TireMountLog]("tire_mount_log", "tml").
		join("JOIN tire ti ON ti.id = tml.tire_id").
		filter(where).
		orderBy("tml.event_date DESC, tml.id DESC")

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toTireMountLogResponse)
}

// Comments godoc
//
//	@Summary	List comments
//	@Tags		comments
//	@Produce	json
//	@Security	BearerAuth
//	@Param		content_type	query		string	false	"asset, issue, work_order or service_entry"
//	@Param		object_id		query		int		false	"Filter by parent id"
//	@Param		limit			query		int		false	"Page size"
//	@Param		offset			query		int		false	"Offset"
//	@Success	200				{object}	dto.CommentPage
//	@Failure	400				{object}	dto.ErrorResponse
//	@Failure	401				{object}	dto.ErrorResponse
//	@Failure	403				{object}	dto.ErrorResponse
//	@Router		/api/v1/comments [get]
func (h *FilteredListHandler) Comments(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("cm.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	contentType, err := queryEnum(c, "content_type", commentContentTypes...)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if contentType != nil {
		where.Add("cm.content_type", filter.Eq, *contentType)
	}
	objectID, err := queryInt64(c, "object_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if objectID != nil {
		where.Add("cm.object_id", filter.Eq, *objectID)
	}

	spec := newListSpec[gen.Comment]("comment", "cm").
		filter(where).
		orderBy("cm.created_at DESC, cm.id DESC")

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toCommentResponse)
}

// CatalogOptions godoc
//
//	@Summary		List catalog options
//	@Description	Catalog options are grouped by category, so a form filling one dropdown filters to its own category rather than fetching the whole catalog.
//	@Tags			catalog-options
//	@Produce		json
//	@Security		BearerAuth
//	@Param			category	query		string	false	"Filter by category"
//	@Param			limit		query		int		false	"Page size"
//	@Param			offset		query		int		false	"Offset"
//	@Success		200			{object}	dto.CatalogOptionPage
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Router			/api/v1/catalog-options [get]
func (h *FilteredListHandler) CatalogOptions(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("co.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	if category := queryStr(c, "category"); category != nil {
		where.Add("co.category", filter.Eq, *category)
	}

	spec := newListSpec[gen.CatalogOption]("catalog_option", "co").
		filter(where).
		orderBy("co.category ASC, co.value ASC, co.id ASC")

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toCatalogOptionResponse)
}

// Media godoc
//
//	@Summary	List media
//	@Tags		media
//	@Produce	json
//	@Security	BearerAuth
//	@Param		asset_id	query		int	false	"Filter by asset"
//	@Param		limit		query		int	false	"Page size"
//	@Param		offset		query		int	false	"Offset"
//	@Success	200			{object}	dto.MediumPage
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Router		/api/v1/media [get]
func (h *FilteredListHandler) Media(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("m.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	assetID, err := queryInt64(c, "asset_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if assetID != nil {
		where.Add("m.asset_id", filter.Eq, *assetID)
	}

	spec := newListSpec[gen.Medium]("media", "m").
		filter(where).
		orderBy("m.created_at DESC, m.id DESC")

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toMediumResponse)
}

// PurchaseOrders godoc
//
//	@Summary	List purchase orders
//	@Tags		purchase-orders
//	@Produce	json
//	@Security	BearerAuth
//	@Param		vendor_id	query		int		false	"Filter by vendor"
//	@Param		state		query		string	false	"DRAFT, PENDING_APPROVAL, REJECTED, APPROVED, PURCHASED, RECEIVED_PARTIAL, RECEIVED_FULL or CLOSED"
//	@Param		order		query		string	false	"created_at, state, id (prefix with - for descending)"
//	@Param		limit		query		int		false	"Page size"
//	@Param		offset		query		int		false	"Offset"
//	@Success	200			{object}	dto.PurchaseOrderPage
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Router		/api/v1/purchase-orders [get]
func (h *FilteredListHandler) PurchaseOrders(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("po.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	vendorID, err := queryInt64(c, "vendor_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if vendorID != nil {
		where.Add("po.vendor_id", filter.Eq, *vendorID)
	}
	state, err := queryEnum(c, "state", purchaseOrderStates...)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if state != nil {
		where.Add("po.state", filter.Eq, *state)
	}

	spec := newListSpec[gen.PurchaseOrder]("purchase_order", "po").
		filter(where).
		orderBy(orderClause(c, map[string]string{
			"created_at": "po.created_at",
			"state":      "po.state",
			"id":         "po.id",
		}, []filter.Sort{{Column: "po.created_at", Desc: true}}, "po.id"))

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toPurchaseOrderResponse)
}

// TireAssignmentRequests godoc
//
//	@Summary		List tire assignment requests
//	@Description	The approval inbox: filter to PENDING for work waiting on someone.
//	@Tags			tire-assignment-requests
//	@Produce		json
//	@Security		BearerAuth
//	@Param			state			query		string	false	"PENDING, APPROVED or REJECTED"
//	@Param			tire_id			query		int		false	"Filter by tire"
//	@Param			vehicle_id		query		int		false	"Filter by vehicle"
//	@Param			requested_from	query		string	false	"Inclusive lower bound (RFC3339 or YYYY-MM-DD)"
//	@Param			requested_to	query		string	false	"Inclusive upper bound (RFC3339 or YYYY-MM-DD)"
//	@Param			order			query		string	false	"requested_at, resolved_at, id (prefix with - for descending)"
//	@Param			limit			query		int		false	"Page size"
//	@Param			offset			query		int		false	"Offset"
//	@Success		200				{object}	dto.TireAssignmentRequestPage
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		403				{object}	dto.ErrorResponse
//	@Router			/api/v1/tire-assignment-requests [get]
func (h *FilteredListHandler) TireAssignmentRequests(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	where := filter.NewWhere(1).Add("tar.company_id", filter.Eq, middleware.CompanyFromContext(ctx))
	state, err := queryEnum(c, "state", tire.RequestPending, tire.RequestApproved, tire.RequestRejected)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if state != nil {
		where.Add("tar.state", filter.Eq, *state)
	}
	tireID, err := queryInt64(c, "tire_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if tireID != nil {
		where.Add("tar.tire_id", filter.Eq, *tireID)
	}
	vehicleID, err := queryInt64(c, "vehicle_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if vehicleID != nil {
		where.Add("tar.vehicle_id", filter.Eq, *vehicleID)
	}
	from, err := queryTime(c, "requested_from")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if from != nil {
		where.Add("tar.requested_at", filter.Gte, *from)
	}
	to, err := queryTime(c, "requested_to")
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if to != nil {
		where.Add("tar.requested_at", filter.Lte, *to)
	}

	spec := newListSpec[gen.TireAssignmentRequest]("tire_assignment_request", "tar").
		filter(where).
		orderBy(orderClause(c, map[string]string{
			"requested_at": "tar.requested_at",
			"resolved_at":  "tar.resolved_at",
			"id":           "tar.id",
		}, []filter.Sort{{Column: "tar.requested_at", Desc: true}}, "tar.id"))

	rows, total, err := runList(ctx, h.pool, spec, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	renderPage(c, rows, total, p, toTireAssignmentRequestResponse)
}

// Vocabularies accepted by the state filters, kept beside the routes that
// validate them. Ported from the Django models the schema came from.
var (
	issueStates         = []string{"OPEN", "RESOLVED", "CLOSED"}
	purchaseOrderStates = []string{
		"DRAFT", "PENDING_APPROVAL", "REJECTED", "APPROVED",
		"PURCHASED", "RECEIVED_PARTIAL", "RECEIVED_FULL", "CLOSED",
	}
	commentContentTypes = []string{
		dto.CommentContentTypeAsset,
		dto.CommentContentTypeIssue,
		dto.CommentContentTypeWorkOrder,
		dto.CommentContentTypeServiceEntry,
	}
)
