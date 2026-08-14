package crud

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
	"fleet/internal/platform/reqbind"
)

// Store is the persistence contract a resource must satisfy to get generic REST
// endpoints. T is the entity, C the create input, U the update input (typically
// request DTOs). Implementations wrap the sqlc repository and translate DTOs.
type Store[T any, C any, U any] interface {
	List(ctx context.Context, p paginate.Params) ([]T, int64, error)
	Get(ctx context.Context, id int64) (T, error)
	Create(ctx context.Context, in C) (T, error)
	Update(ctx context.Context, id int64, in U) (T, error)
	Delete(ctx context.Context, id int64) error
}

// Handler adapts a Store to Gin. Handlers stay thin: parse, delegate, render;
// all business logic lives behind the Store.
type Handler[T any, C any, U any] struct {
	store Store[T, C, U]
}

func NewHandler[T any, C any, U any](store Store[T, C, U]) *Handler[T, C, U] {
	return &Handler[T, C, U]{store: store}
}

// Register wires the standard five REST routes under path (e.g. "/companies").
// For read-only or partial resources, wire the exported handlers individually.
func (h *Handler[T, C, U]) Register(r gin.IRouter, path string) {
	r.GET(path, h.List)
	r.POST(path, h.Create)
	r.GET(path+"/:id", h.Get)
	r.PUT(path+"/:id", h.Update)
	r.DELETE(path+"/:id", h.Delete)
}

func (h *Handler[T, C, U]) List(c *gin.Context) {
	p := paginate.Parse(c)
	items, total, err := h.store.List(c.Request.Context(), p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, paginate.NewPage(items, total, p))
}

func (h *Handler[T, C, U]) Get(c *gin.Context) {
	id, err := idParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	item, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler[T, C, U]) Create(c *gin.Context) {
	var in C
	if err := reqbind.JSON(c, &in); err != nil {
		apierr.Abort(c, err)
		return
	}
	item, err := h.store.Create(c.Request.Context(), in)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler[T, C, U]) Update(c *gin.Context) {
	id, err := idParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	var in U
	if err := reqbind.JSON(c, &in); err != nil {
		apierr.Abort(c, err)
		return
	}
	item, err := h.store.Update(c.Request.Context(), id, in)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler[T, C, U]) Delete(c *gin.Context) {
	id, err := idParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if err := h.store.Delete(c.Request.Context(), id); err != nil {
		apierr.Abort(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func idParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, apierr.BadRequest("invalid id")
	}
	return id, nil
}
