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

// docArchiveAsset godoc
//
//	@Summary		Archive an asset
//	@Description	Retires the record. Use instead of DELETE for anything other records reference — DELETE answers 409 resource_referenced in that case. Archived records disappear from lists and pickers unless ?include_archived=true.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"id"
//	@Success		200	{object}	dto.AssetResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/assets/{id}/archive [post]
func docArchiveAsset() {}

// docRestoreAsset godoc
//
//	@Summary		Restore an archived asset
//	@Description	Clears archived_at, returning the record to the lists and pickers it was hidden from.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"id"
//	@Success		200	{object}	dto.AssetResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/assets/{id}/restore [post]
func docRestoreAsset() {}

// docArchivePart godoc
//
//	@Summary		Archive a part
//	@Description	Retires the record. Use instead of DELETE for anything other records reference — DELETE answers 409 resource_referenced in that case. Archived records disappear from lists and pickers unless ?include_archived=true.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"id"
//	@Success		200	{object}	dto.PartResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/parts/{id}/archive [post]
func docArchivePart() {}

// docRestorePart godoc
//
//	@Summary		Restore an archived part
//	@Description	Clears archived_at, returning the record to the lists and pickers it was hidden from.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"id"
//	@Success		200	{object}	dto.PartResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/parts/{id}/restore [post]
func docRestorePart() {}

// docArchiveVendor godoc
//
//	@Summary		Archive a vendor
//	@Description	Retires the record. Use instead of DELETE for anything other records reference — DELETE answers 409 resource_referenced in that case. Archived records disappear from lists and pickers unless ?include_archived=true.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"id"
//	@Success		200	{object}	dto.VendorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/vendors/{id}/archive [post]
func docArchiveVendor() {}

// docRestoreVendor godoc
//
//	@Summary		Restore an archived vendor
//	@Description	Clears archived_at, returning the record to the lists and pickers it was hidden from.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"id"
//	@Success		200	{object}	dto.VendorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/vendors/{id}/restore [post]
func docRestoreVendor() {}

// docArchiveServiceTask godoc
//
//	@Summary		Archive a service task
//	@Description	Retires the record. Use instead of DELETE for anything other records reference — DELETE answers 409 resource_referenced in that case. Archived records disappear from lists and pickers unless ?include_archived=true.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"id"
//	@Success		200	{object}	dto.ServiceTaskResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/service-tasks/{id}/archive [post]
func docArchiveServiceTask() {}

// docRestoreServiceTask godoc
//
//	@Summary		Restore an archived service task
//	@Description	Clears archived_at, returning the record to the lists and pickers it was hidden from.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"id"
//	@Success		200	{object}	dto.ServiceTaskResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/service-tasks/{id}/restore [post]
func docRestoreServiceTask() {}

// docArchiveInspectionForm godoc
//
//	@Summary		Archive an inspection form
//	@Description	Retires the record. Use instead of DELETE for anything other records reference — DELETE answers 409 resource_referenced in that case. Archived records disappear from lists and pickers unless ?include_archived=true.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"id"
//	@Success		200	{object}	dto.InspectionFormResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/inspection-forms/{id}/archive [post]
func docArchiveInspectionForm() {}

// docRestoreInspectionForm godoc
//
//	@Summary		Restore an archived inspection form
//	@Description	Clears archived_at, returning the record to the lists and pickers it was hidden from.
//	@Tags			archive
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"id"
//	@Success		200	{object}	dto.InspectionFormResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/inspection-forms/{id}/restore [post]
func docRestoreInspectionForm() {}
