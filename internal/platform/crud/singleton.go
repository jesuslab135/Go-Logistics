package crud

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
	"fleet/internal/platform/reqbind"
)

// SingletonStore is a 1:1 child of a parent (e.g. a Vehicle for an Asset). There
// is no collection: the child is addressed by the parent id, and PUT upserts.
// Implementations scope every query so the parent belongs to the caller's tenant.
type SingletonStore[T any, U any] interface {
	Get(ctx context.Context, parentID int64) (T, error)
	Upsert(ctx context.Context, parentID int64, in U) (T, error)
	Delete(ctx context.Context, parentID int64) error
}

type SingletonHandler[T any, U any] struct {
	store SingletonStore[T, U]
}

func NewSingletonHandler[T any, U any](store SingletonStore[T, U]) *SingletonHandler[T, U] {
	return &SingletonHandler[T, U]{store: store}
}

// Register wires GET/PUT/DELETE on /{parent}/:id/{child} (no child id — 1:1).
func (h *SingletonHandler[T, U]) Register(r gin.IRouter, parent, child string) {
	base := parent + "/:id" + child
	r.GET(base, h.Get)
	r.PUT(base, h.Upsert)
	r.DELETE(base, h.Delete)
}

func (h *SingletonHandler[T, U]) Get(c *gin.Context) {
	parentID, err := parentParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	item, err := h.store.Get(c.Request.Context(), parentID)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *SingletonHandler[T, U]) Upsert(c *gin.Context) {
	parentID, err := parentParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	var in U
	if err := reqbind.JSON(c, &in); err != nil {
		apierr.Abort(c, err)
		return
	}
	item, err := h.store.Upsert(c.Request.Context(), parentID, in)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *SingletonHandler[T, U]) Delete(c *gin.Context) {
	parentID, err := parentParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if err := h.store.Delete(c.Request.Context(), parentID); err != nil {
		apierr.Abort(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
