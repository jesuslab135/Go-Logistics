package crud

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
)

// NestedStore is the contract for a child resource nested under a parent. Every
// method receives the parent id from the path; implementations must scope each
// query so the parent belongs to the caller's tenant (join to the company-owning
// ancestor), giving parent + tenant isolation.
type NestedStore[T any, C any, U any] interface {
	List(ctx context.Context, parentID int64, p paginate.Params) ([]T, int64, error)
	Get(ctx context.Context, parentID, id int64) (T, error)
	Create(ctx context.Context, parentID int64, in C) (T, error)
	Update(ctx context.Context, parentID, id int64, in U) (T, error)
	Delete(ctx context.Context, parentID, id int64) error
}

type NestedHandler[T any, C any, U any] struct {
	store NestedStore[T, C, U]
}

func NewNestedHandler[T any, C any, U any](store NestedStore[T, C, U]) *NestedHandler[T, C, U] {
	return &NestedHandler[T, C, U]{store: store}
}

// Register wires /{parent}/:id/{child} (list, create) and
// /{parent}/:id/{child}/:child_id (get, update, delete). The parent segment
// reuses :id so it does not conflict with the parent's own flat :id route.
func (h *NestedHandler[T, C, U]) Register(r gin.IRouter, parent, child string) {
	base := parent + "/:id" + child
	r.GET(base, h.List)
	r.POST(base, h.Create)
	r.GET(base+"/:child_id", h.Get)
	r.PUT(base+"/:child_id", h.Update)
	r.DELETE(base+"/:child_id", h.Delete)
}

func (h *NestedHandler[T, C, U]) List(c *gin.Context) {
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

func (h *NestedHandler[T, C, U]) Get(c *gin.Context) {
	parentID, id, err := parentAndIDParams(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	item, err := h.store.Get(c.Request.Context(), parentID, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *NestedHandler[T, C, U]) Create(c *gin.Context) {
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
	item, err := h.store.Create(c.Request.Context(), parentID, in)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *NestedHandler[T, C, U]) Update(c *gin.Context) {
	parentID, id, err := parentAndIDParams(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	var in U
	if err := c.ShouldBindJSON(&in); err != nil {
		apierr.Abort(c, apierr.BadRequest("invalid request body").Wrap(err))
		return
	}
	item, err := h.store.Update(c.Request.Context(), parentID, id, in)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *NestedHandler[T, C, U]) Delete(c *gin.Context) {
	parentID, id, err := parentAndIDParams(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if err := h.store.Delete(c.Request.Context(), parentID, id); err != nil {
		apierr.Abort(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parentParam(c *gin.Context) (int64, error) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || pid < 1 {
		return 0, apierr.BadRequest("invalid parent id")
	}
	return pid, nil
}

func parentAndIDParams(c *gin.Context) (int64, int64, error) {
	pid, err := parentParam(c)
	if err != nil {
		return 0, 0, err
	}
	id, err := strconv.ParseInt(c.Param("child_id"), 10, 64)
	if err != nil || id < 1 {
		return 0, 0, apierr.BadRequest("invalid id")
	}
	return pid, id, nil
}
