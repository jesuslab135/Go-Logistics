package handler

// Archiving retires a record that something still points at.
//
// DELETE remains available and still hard-deletes, but only where nothing
// references the row; otherwise it answers 409 `resource_referenced` and names
// how many records are holding it. There is no age exemption — a two-year-old
// purchase order still needs its vendor's name to render.
//
// Archived records are excluded from every list, and therefore from the pickers
// built on those lists, unless the request passes ?include_archived=true. That
// is what stops a retired record being referenced again.
//
// archived_at is no longer writable through create or update: archiving is a
// decision, and two ways to make it is one too many.

// docArchiveResource godoc
//
//	@Summary		Archive a record
//	@Description	Retires the record. Use this instead of DELETE for anything other records reference — DELETE answers 409 resource_referenced in that case. Archived records disappear from lists and pickers unless ?include_archived=true. Archiving needs the same permission as editing: deciding a vendor is no longer used is the same kind of decision as correcting its address.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path	int	true	"id"
//	@Success		200	"The archived record, in its own resource's shape"
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/{resource}/{id}/archive [post]
func docArchiveResource() {}

// docRestoreResource godoc
//
//	@Summary		Restore an archived record
//	@Description	Clears archived_at, returning the record to the lists and pickers it was hidden from.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path	int	true	"id"
//	@Success		200	"The restored record, in its own resource's shape"
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/{resource}/{id}/restore [post]
func docRestoreResource() {}
