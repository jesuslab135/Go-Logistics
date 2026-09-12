package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/filter"
)

// facetCount formats one scanned facet row into its DTO. It is split out from
// scanFacets, which is DB-touching and untested, so this formatting — the part
// with actual branching to get wrong — has direct unit coverage.
//
// A nil value means the grouped dimension is itself null (e.g. an asset with
// no status, via a LEFT join): fmt.Sprint(nil) would render the literal string
// "<nil>", which is not an acceptable facet key, so that bucket gets the
// sentinel Value "" / Label "No status" instead. A non-nil value with no label
// (an unjoined dimension that is its own label) falls back the label to the
// value's string form.
func facetCount(value any, label *string) dto.FacetCount {
	if value == nil {
		l := "No status"
		if label != nil {
			l = *label
		}
		return dto.FacetCount{Value: "", Label: l}
	}
	v := fmt.Sprint(value)
	l := v
	if label != nil {
		l = *label
	}
	return dto.FacetCount{Value: v, Label: l}
}

// scanFacets executes a facetQuery statement and collects its rows. It is a
// thin DB-touching shell around facetQuery's pure SQL and facetCount's pure
// formatting, kept small because the DB round trip itself is not unit-tested.
func scanFacets(ctx context.Context, pool *pgxpool.Pool, sql string, args []any) ([]dto.FacetCount, error) {
	// Like runList (see listquery.go), this queries the pool directly rather
	// than through dbctx.DB, so it does not join a transaction the request
	// context carries. Facet counts are read-only and never run inside an
	// import's transaction, so there is nothing to join; both files are
	// allowlisted by TestHandlersUseDbctxRatherThanThePoolDirectly for exactly
	// this reason.
	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dto.FacetCount
	for rows.Next() {
		var value any
		var label *string
		var count int64
		if err := rows.Scan(&value, &label, &count); err != nil {
			return nil, err
		}
		fc := facetCount(value, label)
		fc.Count = count
		out = append(out, fc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// workOrderFacetWhere is the work-orders list's filter set minus status_id,
// the dimension WorkOrderFacets groups by: applying it here would make every
// bucket describe only its own status.
func (h *FilteredListHandler) workOrderFacetWhere(c *gin.Context, company int64) (*filter.Where, error) {
	where := filter.NewWhere(1).Add("wo.company_id", filter.Eq, company)
	assetID, err := queryInt64(c, "asset_id")
	if err != nil {
		return nil, err
	}
	if assetID != nil {
		where.Add("wo.asset_id", filter.Eq, *assetID)
	}
	if err := applyCustomFieldFilters(c, where, "wo.custom_fields", "work-orders", gen.New(h.pool)); err != nil {
		return nil, err
	}
	return where, nil
}

// WorkOrderFacets godoc
//
//	@Summary	Work order counts by status
//	@Tags		work-orders
//	@Produce	json
//	@Security	BearerAuth
//	@Param		asset_id	query		int	false	"Filter by asset"
//	@Success	200			{object}	dto.WorkOrderFacetsResponse
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Router		/api/v1/work-orders/facets [get]
func (h *FilteredListHandler) WorkOrderFacets(c *gin.Context) {
	ctx := c.Request.Context()
	where, err := h.workOrderFacetWhere(c, middleware.CompanyFromContext(ctx))
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	sql, args := facetQuery(
		"FROM work_order wo JOIN work_order_status s ON s.id = wo.status_id",
		where, "wo.status_id", "s.name")
	counts, err := scanFacets(ctx, h.pool, sql, args)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.WorkOrderFacetsResponse{Status: counts})
}

// issueFacetWhere is the issues list's filter set minus state, the dimension
// IssueFacets groups by.
func (h *FilteredListHandler) issueFacetWhere(c *gin.Context, company int64) (*filter.Where, error) {
	where := filter.NewWhere(1).Add("i.company_id", filter.Eq, company)
	assetID, err := queryInt64(c, "asset_id")
	if err != nil {
		return nil, err
	}
	if assetID != nil {
		where.Add("i.asset_id", filter.Eq, *assetID)
	}
	if err := applyCustomFieldFilters(c, where, "i.custom_fields", "issues", gen.New(h.pool)); err != nil {
		return nil, err
	}
	return where, nil
}

// IssueFacets godoc
//
//	@Summary	Issue counts by state
//	@Tags		issues
//	@Produce	json
//	@Security	BearerAuth
//	@Param		asset_id	query		int	false	"Filter by asset"
//	@Success	200			{object}	dto.IssueFacetsResponse
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Router		/api/v1/issues/facets [get]
func (h *FilteredListHandler) IssueFacets(c *gin.Context) {
	ctx := c.Request.Context()
	where, err := h.issueFacetWhere(c, middleware.CompanyFromContext(ctx))
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	sql, args := facetQuery("FROM issue i", where, "i.state", "i.state")
	counts, err := scanFacets(ctx, h.pool, sql, args)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.IssueFacetsResponse{State: counts})
}

// assetFacetBaseWhere seeds the assets list's tenant scope, archived-visibility
// rule and every filter shared by all three asset facets (q, custom fields).
// vehicle_type and status_id are added by the caller, per dimension, since one
// of them is always the grouped-by column and must be omitted for that facet.
func (h *FilteredListHandler) assetFacetBaseWhere(c *gin.Context, ctx context.Context, company int64) (*filter.Where, error) {
	where := filter.NewWhere(1).Add("a.company_id", filter.Eq, company)
	if !includeArchived(ctx) {
		where.Raw("a.archived_at IS NULL")
	}
	if q := queryStr(c, "q"); q != nil {
		pattern := "%" + *q + "%"
		where.Raw("(a.name ILIKE ? OR a.vin_sn ILIKE ? OR a.license_plate ILIKE ?)", pattern, pattern, pattern)
	}
	if err := applyCustomFieldFilters(c, where, "a.custom_fields", "assets", gen.New(h.pool)); err != nil {
		return nil, err
	}
	return where, nil
}

// assetVehicleTypeFacetWhere is the assets list's filter set minus vehicle_type,
// the dimension the vehicle_type facet groups by.
func (h *FilteredListHandler) assetVehicleTypeFacetWhere(c *gin.Context, ctx context.Context, company int64) (*filter.Where, error) {
	where, err := h.assetFacetBaseWhere(c, ctx, company)
	if err != nil {
		return nil, err
	}
	statusID, err := queryInt64(c, "status_id")
	if err != nil {
		return nil, err
	}
	if statusID != nil {
		where.Add("a.status_id", filter.Eq, *statusID)
	}
	return where, nil
}

// assetStatusFacetWhere is the assets list's filter set minus status_id, the
// dimension the status facet groups by.
func (h *FilteredListHandler) assetStatusFacetWhere(c *gin.Context, ctx context.Context, company int64) (*filter.Where, error) {
	where, err := h.assetFacetBaseWhere(c, ctx, company)
	if err != nil {
		return nil, err
	}
	if vehicleType := queryStr(c, "vehicle_type"); vehicleType != nil {
		where.Add("a.vehicle_type", filter.Eq, *vehicleType)
	}
	return where, nil
}

// assetTrailerTypeFacetWhere applies every asset list filter: trailer_type is
// not a filterable dimension on the assets list, so nothing needs omitting.
func (h *FilteredListHandler) assetTrailerTypeFacetWhere(c *gin.Context, ctx context.Context, company int64) (*filter.Where, error) {
	where, err := h.assetFacetBaseWhere(c, ctx, company)
	if err != nil {
		return nil, err
	}
	if vehicleType := queryStr(c, "vehicle_type"); vehicleType != nil {
		where.Add("a.vehicle_type", filter.Eq, *vehicleType)
	}
	statusID, err := queryInt64(c, "status_id")
	if err != nil {
		return nil, err
	}
	if statusID != nil {
		where.Add("a.status_id", filter.Eq, *statusID)
	}
	return where, nil
}

// AssetFacets godoc
//
//	@Summary		Asset counts by vehicle type, trailer type and status
//	@Description	Applies the assets list's archived-visibility rule and filters, each query omitting only the dimension it groups by.
//	@Tags			assets
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q					query		string	false	"Search name, VIN/serial and license plate"
//	@Param			vehicle_type		query		string	false	"Filter by vehicle type"
//	@Param			status_id			query		int		false	"Filter by status"
//	@Param			include_archived	query		bool	false	"Include archived records"
//	@Success		200					{object}	dto.AssetFacetsResponse
//	@Failure		400					{object}	dto.ErrorResponse
//	@Failure		401					{object}	dto.ErrorResponse
//	@Failure		403					{object}	dto.ErrorResponse
//	@Router			/api/v1/assets/facets [get]
func (h *FilteredListHandler) AssetFacets(c *gin.Context) {
	ctx := c.Request.Context()
	company := middleware.CompanyFromContext(ctx)

	vehicleTypeWhere, err := h.assetVehicleTypeFacetWhere(c, ctx, company)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	statusWhere, err := h.assetStatusFacetWhere(c, ctx, company)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	trailerTypeWhere, err := h.assetTrailerTypeFacetWhere(c, ctx, company)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	vtSQL, vtArgs := facetQuery("FROM asset a", vehicleTypeWhere, "a.vehicle_type", "a.vehicle_type")
	vehicleTypeCounts, err := scanFacets(ctx, h.pool, vtSQL, vtArgs)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	// LEFT, not INNER: asset.status_id is nullable, and an unassigned asset is
	// a real, reachable state — an INNER join would drop it from the facet
	// silently instead of counting it in a null bucket.
	stSQL, stArgs := facetQuery("FROM asset a LEFT JOIN asset_status st ON st.id = a.status_id",
		statusWhere, "a.status_id", "st.name")
	statusCounts, err := scanFacets(ctx, h.pool, stSQL, stArgs)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	ttSQL, ttArgs := facetQuery("FROM asset a JOIN trailer t ON t.asset_id = a.id",
		trailerTypeWhere, "t.trailer_type", "t.trailer_type")
	trailerTypeCounts, err := scanFacets(ctx, h.pool, ttSQL, ttArgs)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.AssetFacetsResponse{
		VehicleType: vehicleTypeCounts,
		TrailerType: trailerTypeCounts,
		Status:      statusCounts,
	})
}
