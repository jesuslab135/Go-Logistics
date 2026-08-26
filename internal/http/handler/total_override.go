package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/reqbind"
)

// An override is the escape hatch for a figure the formula cannot express: a
// negotiated total, a rounding agreed with a vendor, a correction to a
// historical document.
//
// It is audited because it replaces the only number the system can justify, and
// it is a distinct action rather than a writable field so that setting one is a
// decision somebody made rather than a side effect of saving a form.
//
// It does not survive an edit to the line items. That is the point: a stale
// override quietly misstating a total is exactly the failure server-authoritative
// totals exist to prevent, so a line-item write clears it and the document
// visibly returns to its computed total.
type TotalOverrideHandler struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewTotalOverrideHandler(q *gen.Queries, pool *pgxpool.Pool) *TotalOverrideHandler {
	return &TotalOverrideHandler{q: q, pool: pool}
}

// WorkOrder godoc
//
//	@Summary		Override a work order's total
//	@Description	Replaces the computed total with an explicit figure, recording who set it and why. The component columns (parts_subtotal, labor_subtotal, subtotal) stay computed, so the difference between what the document adds up to and what it is being charged at stays visible — which is the reason an override is audited. Send a null amount to drop the override and return to the computed total. Any later edit to the line items clears it automatically.
//	@Tags			work-orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Work order id"
//	@Param			body	body		dto.TotalOverrideRequest	true	"Override"
//	@Success		200		{object}	dto.WorkOrderResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/work-orders/{id}/override-total [post]
func (h *TotalOverrideHandler) WorkOrder(c *gin.Context) {
	id, req, ok := h.bind(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	company := middleware.CompanyFromContext(ctx)
	actor, _ := actorOf(ctx)
	now := time.Now().UTC()

	var out dto.WorkOrderResponse
	err := inTx(ctx, h.pool, h.q, func(qtx *gen.Queries) error {
		if _, err := qtx.SetWorkOrderTotalOverride(ctx, gen.SetWorkOrderTotalOverrideParams{
			ID: id, CompanyID: company,
			TotalOverride: req.Amount, Reason: &req.Reason, Actor: actor, At: &now,
		}); err != nil {
			return err
		}
		// Recompute so total_amount reflects the decision immediately: the
		// override is what the document is charged at, and the components beside
		// it stay computed.
		if err := recalcWorkOrder(ctx, qtx, id, now); err != nil {
			return err
		}
		r, err := qtx.GetWorkOrder(ctx, gen.GetWorkOrderParams{ID: id, CompanyID: company})
		if err != nil {
			return err
		}
		out = toWorkOrderResponse(r)
		return nil
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// PurchaseOrder godoc
//
//	@Summary		Override a purchase order's total
//	@Description	As for work orders: replaces the computed total with an audited figure, null clears it, and a line-item edit clears it automatically.
//	@Tags			purchase-orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Purchase order id"
//	@Param			body	body		dto.TotalOverrideRequest	true	"Override"
//	@Success		200		{object}	dto.PurchaseOrderResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/purchase-orders/{id}/override-total [post]
func (h *TotalOverrideHandler) PurchaseOrder(c *gin.Context) {
	id, req, ok := h.bind(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	company := middleware.CompanyFromContext(ctx)
	actor, _ := actorOf(ctx)
	now := time.Now().UTC()

	var out dto.PurchaseOrderResponse
	err := inTx(ctx, h.pool, h.q, func(qtx *gen.Queries) error {
		if _, err := qtx.SetPurchaseOrderTotalOverride(ctx, gen.SetPurchaseOrderTotalOverrideParams{
			ID: id, CompanyID: company,
			TotalOverride: req.Amount, Reason: &req.Reason, Actor: actor, At: &now,
		}); err != nil {
			return err
		}
		if err := recalcPurchaseOrder(ctx, qtx, id, now); err != nil {
			return err
		}
		r, err := qtx.GetPurchaseOrder(ctx, gen.GetPurchaseOrderParams{ID: id, CompanyID: company})
		if err != nil {
			return err
		}
		out = toPurchaseOrderResponse(r)
		return nil
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// ServiceEntry godoc
//
//	@Summary		Override a service entry's total
//	@Description	As for work orders: replaces the computed total with an audited figure, null clears it, and a line-item edit clears it automatically.
//	@Tags			service-entries
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Service entry id"
//	@Param			body	body		dto.TotalOverrideRequest	true	"Override"
//	@Success		200		{object}	dto.ServiceEntryResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/service-entries/{id}/override-total [post]
func (h *TotalOverrideHandler) ServiceEntry(c *gin.Context) {
	id, req, ok := h.bind(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	company := middleware.CompanyFromContext(ctx)
	actor, _ := actorOf(ctx)
	now := time.Now().UTC()

	var out dto.ServiceEntryResponse
	err := inTx(ctx, h.pool, h.q, func(qtx *gen.Queries) error {
		if _, err := qtx.SetServiceEntryTotalOverride(ctx, gen.SetServiceEntryTotalOverrideParams{
			ID: id, CompanyID: company,
			TotalOverride: req.Amount, Reason: &req.Reason, Actor: actor, At: &now,
		}); err != nil {
			return err
		}
		if err := recalcServiceEntry(ctx, qtx, id, now); err != nil {
			return err
		}
		r, err := qtx.GetServiceEntry(ctx, gen.GetServiceEntryParams{ID: id, CompanyID: company})
		if err != nil {
			return err
		}
		out = toServiceEntryResponse(r)
		return nil
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// bind parses the id and body shared by all three routes. A reason is required
// whenever an amount is given: an unexplained override is a number nobody can
// defend later, which is the situation it exists to avoid.
func (h *TotalOverrideHandler) bind(c *gin.Context) (int64, dto.TotalOverrideRequest, bool) {
	var req dto.TotalOverrideRequest

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid id"))
		return 0, req, false
	}
	if err := reqbind.JSON(c, &req); err != nil {
		apierr.Abort(c, err)
		return 0, req, false
	}

	req.Reason = strings.TrimSpace(req.Reason)
	if req.Amount != nil && req.Reason == "" {
		apierr.Abort(c, apierr.Validation(map[string]string{
			"reason": "a reason is required: an override replaces the computed total, and nobody can defend it later without one",
		}))
		return 0, req, false
	}
	if req.Amount == nil {
		// Clearing has nothing to explain.
		req.Reason = ""
	}
	return id, req, true
}
