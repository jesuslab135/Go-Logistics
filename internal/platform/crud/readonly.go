package crud

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
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
