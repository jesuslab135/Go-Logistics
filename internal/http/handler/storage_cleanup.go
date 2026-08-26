package handler

import (
	"context"
	"errors"
	"log/slog"

	"fleet/internal/platform/storage"
)

// Replacing or clearing a file URL used to leave the previous object in the
// bucket forever, with nothing referencing it. The lifecycle policy is to
// delete the old object once the row that referenced it has been written.
//
// The delete happens after the write, not inside it: an object store cannot
// join a database transaction, and deleting first would destroy a live file if
// the write then failed. The cost of this ordering is that a crash between the
// two leaves one orphan — far better than deleting something still in use.
// Failures are logged and swallowed for the same reason: the user's write
// succeeded, and reporting an error for unreclaimed bytes would be misleading.

// reclaimObjects deletes objects that are no longer referenced. Values may be
// storage keys or the absolute URLs older rows still hold; either resolves to a
// key. Values that are empty, or URLs this storage does not own (an externally
// hosted image, say), are skipped rather than treated as errors.
func reclaimObjects(ctx context.Context, files storage.Storage, log *slog.Logger, refs ...string) {
	if files == nil {
		return
	}
	if log == nil {
		log = slog.Default()
	}

	for _, ref := range refs {
		key, ok := fileKey(files, ref)
		if !ok {
			continue
		}
		if err := files.Delete(ctx, key); err != nil && !errors.Is(err, storage.ErrNotFound) {
			// Worth knowing about — it means bytes are accumulating — but not
			// worth failing a request that already did what was asked.
			log.WarnContext(ctx, "orphaned upload object could not be deleted", "key", key, "error", err)
		}
	}
}

// fileOwner is embedded by the stores whose rows hold upload URLs, so each one
// does not have to carry the same two dependencies by hand.
type fileOwner struct {
	files storage.Storage
	log   *slog.Logger
}

func newFileOwner(files storage.Storage, log *slog.Logger) fileOwner {
	return fileOwner{files: files, log: log}
}

func (o fileOwner) reclaim(ctx context.Context, refs ...string) {
	reclaimObjects(ctx, o.files, o.log, refs...)
}

// reclaimReplaced deletes the previous object when a write replaced or cleared
// the reference, and does nothing when it was left alone.
func (o fileOwner) reclaimReplaced(ctx context.Context, before, after *string) {
	if previous, ok := changedURL(before, after); ok {
		o.reclaim(ctx, previous)
	}
}

// changedURL reports the previous reference when a write replaced or cleared it,
// and therefore when that object is no longer referenced.
func changedURL(before, after *string) (string, bool) {
	previous := ""
	if before != nil {
		previous = *before
	}
	current := ""
	if after != nil {
		current = *after
	}
	if previous != "" && previous != current {
		return previous, true
	}
	return "", false
}
