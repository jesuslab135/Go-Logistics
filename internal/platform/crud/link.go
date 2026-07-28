package crud

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
)

// LinkStore is the contract for a many-to-many join table exposed under its
// parent. Unlike NestedStore the join row itself is never addressable: a link is
// identified by the pair (parent, target), so there is no Get or Update — a link
// either exists or it does not.
//
// T is the *linked entity* (the employee, the issue), not the join row: clients
// asking for an issue's assignees want the assignees, not the row ids that
// connect them. C is the create input carrying the target id.
type LinkStore[T any, C any] interface {
	List(ctx context.Context, parentID int64, p paginate.Params) ([]T, int64, error)
	Link(ctx context.Context, parentID int64, in C) (T, error)
	Unlink(ctx context.Context, parentID, targetID int64) error
}

type LinkHandler[T any, C any] struct {
	store LinkStore[T, C]
	// targetParam is the path parameter naming the far side of the link
	// ("employee_id", "issue_id"), captured at Register time.
	targetParam string
}

func NewLinkHandler[T any, C any](store LinkStore[T, C]) *LinkHandler[T, C] {
	return &LinkHandler[T, C]{store: store}
}

// Register wires /{parent}/:id/{child} (list, link) and
// /{parent}/:id/{child}/:{targetParam} (unlink). As with NestedHandler the
// parent segment reuses :id so it cannot conflict with the parent's flat route.
func (h *LinkHandler[T, C]) Register(r gin.IRouter, parent, child, targetParam string) {
	h.targetParam = targetParam

	base := parent + "/:id" + child
	r.GET(base, h.List)
	r.POST(base, h.Link)
	r.DELETE(base+"/:"+targetParam, h.Unlink)
}

func (h *LinkHandler[T, C]) List(c *gin.Context) {
	parentID, err := parentParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	p := paginate.Parse(c)
	items, total, err := h.store.List(c.Request.Context(), parentID, p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, paginate.NewPage(items, total, p))
}

// Link is idempotent: re-posting an existing link returns the linked entity with
// 201 rather than a conflict, so a client that retries a dropped response does
// not have to distinguish the two cases.
func (h *LinkHandler[T, C]) Link(c *gin.Context) {
	parentID, err := parentParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	var in C
	if err := c.ShouldBindJSON(&in); err != nil {
		apierr.Abort(c, apierr.BadRequest("invalid request body").Wrap(err))
		return
	}
	item, err := h.store.Link(c.Request.Context(), parentID, in)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *LinkHandler[T, C]) Unlink(c *gin.Context) {
	parentID, err := parentParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	targetID, err := strconv.ParseInt(c.Param(h.targetParam), 10, 64)
	if err != nil || targetID < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid "+h.targetParam))
		return
	}
	if err := h.store.Unlink(c.Request.Context(), parentID, targetID); err != nil {
		apierr.Abort(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
