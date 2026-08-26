package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
)

// archived_at existed on five catalogs — assets, parts, vendors, service tasks,
// inspection forms — as a plain writable timestamp, which meant three things
// were true at once: nothing distinguished archiving from any other edit,
// archived rows still appeared in every list and picker, and DELETE went
// straight through whether or not anything referenced the row.
//
// The rule now: a record nothing points at can be deleted; a record something
// points at is archived instead. Hard-deleting a referenced row either breaks a
// foreign key or silently blanks a column somebody's report depends on, and
// neither is a thing to discover from a stack trace.

// archiver is one resource's archive behaviour: how to set and clear the
// timestamp, and how to count what points at a row.
type archiver[T any] struct {
	// resource names the thing in error messages, in the plural the API uses.
	resource string
	archive  func(ctx context.Context, id, companyID int64, at time.Time) (T, error)
	restore  func(ctx context.Context, id, companyID int64, at time.Time) (T, error)
	// references counts every inbound relation. Nothing is exempt by age: a
	// two-year-old purchase order still needs its vendor's name to render.
	references func(ctx context.Context, id int64) (int64, error)
}

// ArchiveHandler wires /{resource}/{id}/archive and /restore.
type ArchiveHandler[T any] struct {
	a archiver[T]
}

func newArchiveHandler[T any](a archiver[T]) *ArchiveHandler[T] {
	return &ArchiveHandler[T]{a: a}
}

// Register adds the two routes. Archiving needs the same permission as editing:
// it is an editorial act, not a privileged one — deciding a vendor is no longer
// used is the same kind of decision as correcting its address.
func (h *ArchiveHandler[T]) Register(r gin.IRouter, path string) {
	r.POST(path+"/:id/archive", h.Archive)
	r.POST(path+"/:id/restore", h.Restore)
}

func (h *ArchiveHandler[T]) Archive(c *gin.Context) { h.run(c, h.a.archive) }

func (h *ArchiveHandler[T]) Restore(c *gin.Context) { h.run(c, h.a.restore) }

func (h *ArchiveHandler[T]) run(c *gin.Context, action func(context.Context, int64, int64, time.Time) (T, error)) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid id"))
		return
	}

	ctx := c.Request.Context()
	out, err := action(ctx, id, middleware.CompanyFromContext(ctx), time.Now().UTC())
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// guardDelete refuses to delete a row something still points at, and says which
// resource is holding it. It is a 409 rather than a 422 because nothing about
// the request is malformed — the record is simply still in use.
func guardDelete(ctx context.Context, resource string, id int64, references func(context.Context, int64) (int64, error)) error {
	n, err := references(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return apierr.New(http.StatusConflict, "resource_referenced",
			"this "+resource+" is referenced by "+strconv.FormatInt(n, 10)+
				" other record(s) and cannot be deleted; archive it instead with POST .../archive")
	}
	return nil
}

// includeArchived reads the opt-in from the request. Archived rows are hidden
// by default: an archived record is one somebody retired, and showing it in the
// default list — and in the pickers built from that list — is how it gets
// referenced again.
//
// The context carries it because the generic CRUD handler owns the List call
// and has no way to pass a per-resource query flag through.
func includeArchived(ctx context.Context) bool {
	v, _ := ctx.Value(includeArchivedKey{}).(bool)
	return v
}

type includeArchivedKey struct{}

// WithIncludeArchived is the middleware that reads ?include_archived=true.
func WithIncludeArchived() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Query("include_archived") == "true" {
			ctx := context.WithValue(c.Request.Context(), includeArchivedKey{}, true)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}
