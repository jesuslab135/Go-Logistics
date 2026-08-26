package crud

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
	"fleet/internal/platform/reqbind"
)

// ReadOnlyNestedStore is a nested resource that can be read but not written
// through its own routes. An append-only log is the case this exists for: the
// rows are produced by whatever caused the event, inside that operation's
// transaction, so a route that let a client write them directly would let the
// record disagree with the thing it records.
type ReadOnlyNestedStore[T any] interface {
	List(ctx context.Context, parentID int64, p paginate.Params) ([]T, int64, error)
	Get(ctx context.Context, parentID, id int64) (T, error)
}

type ReadOnlyNestedHandler[T any] struct {
	store ReadOnlyNestedStore[T]
	// replacement names what to do instead, in the 405 body. A client that was
	// writing these rows needs to be told where the behaviour went, not just
	// that it is gone.
	replacement string
}

func NewReadOnlyNestedHandler[T any](store ReadOnlyNestedStore[T], replacement string) *ReadOnlyNestedHandler[T] {
	return &ReadOnlyNestedHandler[T]{store: store, replacement: replacement}
}

// Register wires the two read routes, and answers the three write verbs with
// 405 rather than leaving them unrouted. An unrouted path is a 404, which is
// indistinguishable from a typo; a 405 naming its replacement turns a client
// bug report into a one-line fix.
func (h *ReadOnlyNestedHandler[T]) Register(r gin.IRouter, parent, child string) {
	base := parent + "/:id" + child
	r.GET(base, h.List)
	r.GET(base+"/:child_id", h.Get)

	r.POST(base, h.retired)
	r.PUT(base+"/:child_id", h.retired)
	r.DELETE(base+"/:child_id", h.retired)
}

func (h *ReadOnlyNestedHandler[T]) retired(c *gin.Context) {
	apierr.Abort(c, apierr.New(http.StatusMethodNotAllowed, "route_retired", h.replacement))
}

func (h *ReadOnlyNestedHandler[T]) List(c *gin.Context) {
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

func (h *ReadOnlyNestedHandler[T]) Get(c *gin.Context) {
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

// AppendOnlyStore is a resource whose rows can be created and read but never
// edited or deleted. A ledger is the case this exists for: a row records
// something that happened, and history that can be rewritten records nothing.
type AppendOnlyStore[T any, C any] interface {
	List(ctx context.Context, p paginate.Params) ([]T, int64, error)
	Get(ctx context.Context, id int64) (T, error)
	Create(ctx context.Context, in C) (T, error)
}

type AppendOnlyHandler[T any, C any] struct {
	store       AppendOnlyStore[T, C]
	replacement string
}

func NewAppendOnlyHandler[T any, C any](store AppendOnlyStore[T, C], replacement string) *AppendOnlyHandler[T, C] {
	return &AppendOnlyHandler[T, C]{store: store, replacement: replacement}
}

// Register wires list, create and get, and answers PUT and DELETE with 405
// naming the replacement. As with ReadOnlyNestedHandler, the retired verbs stay
// routed so a client gets told what to do instead of a bare 404.
func (h *AppendOnlyHandler[T, C]) Register(r gin.IRouter, path string) {
	r.GET(path, h.List)
	r.POST(path, h.Create)
	r.GET(path+"/:id", h.Get)

	r.PUT(path+"/:id", h.Retired)
	r.DELETE(path+"/:id", h.Retired)
}

func (h *AppendOnlyHandler[T, C]) Retired(c *gin.Context) {
	apierr.Abort(c, apierr.New(http.StatusMethodNotAllowed, "route_retired", h.replacement))
}

func (h *AppendOnlyHandler[T, C]) List(c *gin.Context) {
	p := paginate.Parse(c)
	items, total, err := h.store.List(c.Request.Context(), p)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, paginate.NewPage(items, total, p))
}

func (h *AppendOnlyHandler[T, C]) Get(c *gin.Context) {
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

func (h *AppendOnlyHandler[T, C]) Create(c *gin.Context) {
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
