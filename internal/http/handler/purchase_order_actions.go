package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/domain/purchaseorder"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/dbctx"
)

// PurchaseOrderActionHandler serves the workflow transitions.
//
// Before these existed, PUT /purchase-orders/{id} was the only write: a
// whole-record replace that accepted state and every workflow timestamp from
// the client. An order could be set to CLOSED without passing through approval,
// approved_by_id could name somebody who never approved anything, and nothing
// was recorded either way. The columns implied a state machine that nothing
// enforced.
type PurchaseOrderActionHandler struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewPurchaseOrderActionHandler(q *gen.Queries, pool *pgxpool.Pool) *PurchaseOrderActionHandler {
	return &PurchaseOrderActionHandler{q: q, pool: pool}
}

// Handle runs one named transition. Every route shares this: they differ only
// in the action name, which is what decides the legal source states, the
// timestamp stamped, and whether a reason is required.
func (h *PurchaseOrderActionHandler) Handle(action string) gin.HandlerFunc {
	transition, ok := purchaseorder.Lookup(action)
	if !ok {
		// A registration bug, not a request error: fail at startup rather than
		// serving a route that cannot work.
		panic("purchase order: unknown transition " + action)
	}

	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id < 1 {
			apierr.Abort(c, apierr.BadRequest("invalid id"))
			return
		}

		var req dto.PurchaseOrderTransitionRequest
		// The body is optional for every transition but reject, so a malformed
		// one is only worth refusing when it carries something we need.
		if c.Request.ContentLength > 0 {
			if err := bindJSONValidated(c, &req); err != nil {
				apierr.Abort(c, err)
				return
			}
		}
		reason := strings.TrimSpace(req.Reason)
		if transition.RequiresReason && reason == "" {
			apierr.Abort(c, apierr.Validation(map[string]string{
				"reason": "a reason is required, so the buyer knows what to change",
			}))
			return
		}

		out, err := h.apply(c.Request.Context(), id, transition, reason)
		if err != nil {
			apierr.Abort(c, err)
			return
		}
		c.JSON(http.StatusOK, out)
	}
}

func (h *PurchaseOrderActionHandler) apply(ctx context.Context, id int64, t purchaseorder.Transition, reason string) (dto.PurchaseOrderResponse, error) {
	company := middleware.CompanyFromContext(ctx)

	tx, err := dbctx.Begin(ctx, h.pool)
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	defer tx.Rollback(ctx)
	qtx := h.q.WithTx(tx)

	// Locked, so the state this validates against is the state it replaces.
	// Without it two concurrent approvals both see PENDING_APPROVAL and both
	// stamp an approval.
	order, err := qtx.LockPurchaseOrder(ctx, gen.LockPurchaseOrderParams{ID: id, CompanyID: company})
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	if !t.Allows(order.State) {
		return dto.PurchaseOrderResponse{}, apierr.New(http.StatusConflict, "invalid_transition",
			"cannot "+t.Action+" a purchase order in state "+order.State)
	}

	now := time.Now().UTC()
	actorID, actorType := actorOf(ctx)

	params := gen.ApplyPurchaseOrderTransitionParams{
		ID:        id,
		CompanyID: company,
		State:     t.To,
		UpdatedAt: now,
		// Cleared on every transition but reject, which sets it. A stale reason
		// left on an order that has since been approved reads as a live one.
		RejectionReason: "",
	}
	if t.Action == purchaseorder.ActionReject {
		params.RejectionReason = reason
	}
	stampTransition(&params, t.Action, now, actorID)

	updated, err := qtx.ApplyPurchaseOrderTransition(ctx, params)
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}

	if _, err := qtx.CreatePurchaseOrderStatusLog(ctx, gen.CreatePurchaseOrderStatusLogParams{
		ParentID:        id,
		CompanyID:       company,
		FromState:       order.State,
		ToState:         t.To,
		ActorEmployeeID: actorID,
		ActorType:       actorType,
		Reason:          reason,
		ChangedAt:       now,
	}); err != nil {
		return dto.PurchaseOrderResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	return toPurchaseOrderResponse(updated), nil
}

// stampTransition sets the one timestamp/actor pair the action owns. Each column
// belongs to exactly one action, and the query COALESCEs the rest, so a
// transition never disturbs a stamp it did not make — an order that was
// submitted, rejected and resubmitted keeps the original submitted_at until
// submit runs again, which is what it means.
func stampTransition(p *gen.ApplyPurchaseOrderTransitionParams, action string, now time.Time, actor *int64) {
	switch action {
	case purchaseorder.ActionSubmit:
		p.SubmittedAt, p.SubmittedByID = &now, actor
	case purchaseorder.ActionApprove:
		p.ApprovedAt, p.ApprovedByID = &now, actor
	case purchaseorder.ActionReject:
		p.RejectedAt, p.RejectedByID = &now, actor
	case purchaseorder.ActionPurchase:
		p.PurchasedAt = &now
	case purchaseorder.ActionReceivePartial:
		p.ReceivedPartialAt = &now
	case purchaseorder.ActionReceiveFull:
		p.ReceivedFullAt = &now
	case purchaseorder.ActionClose:
		p.ClosedAt = &now
	case purchaseorder.ActionRevise:
		// Revise moves REJECTED back to DRAFT and stamps nothing: it undoes a
		// decision rather than making one, and the rejection it followed stays
		// in the log.
	}
}
