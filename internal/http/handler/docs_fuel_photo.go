package handler

// docDListFuelPhotos godoc
//
//	@Summary	List fuel-photos
//	@Tags	fuel-photos
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.FuelPhotoPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-entries/{id}/photos [get]
func docDListFuelPhotos() {}

// docDCreateFuelPhoto godoc
//
//	@Summary	Create fuel-photos
//	@Tags	fuel-photos
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateFuelPhotoRequest	true	"body"
//	@Success	201	{object}	dto.FuelPhotoResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-entries/{id}/photos [post]
func docDCreateFuelPhoto() {}

// docDGetFuelPhoto godoc
//
//	@Summary	Get fuel-photos
//	@Tags	fuel-photos
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.FuelPhotoResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-entries/{id}/photos/{child_id} [get]
func docDGetFuelPhoto() {}

// docDUpdateFuelPhoto godoc
//
//	@Summary	Update fuel-photos
//	@Tags	fuel-photos
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateFuelPhotoRequest	true	"body"
//	@Success	200	{object}	dto.FuelPhotoResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-entries/{id}/photos/{child_id} [put]
func docDUpdateFuelPhoto() {}

// docDDeleteFuelPhoto godoc
//
//	@Summary	Delete fuel-photos
//	@Tags	fuel-photos
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/fuel-entries/{id}/photos/{child_id} [delete]
func docDDeleteFuelPhoto() {}

// docSetFuelPhotoPrimary godoc
//
//	@Summary		Make a fuel photo the primary one
//	@Description	Promotes this photo and demotes the entry's others, in one transaction. At most one photo per entry may be primary, enforced by a partial unique index — which is why is_primary is not a field on create or update. Zero photos, and photos with no primary among them, are both valid: deleting the primary promotes nothing, because auto-promotion would designate a photo nobody chose.
//	@Tags			fuel-photos
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id			path	int	true	"Fuel entry id"
//	@Param			child_id	path	int	true	"Photo id"
//	@Success		200			{object}	dto.FuelPhotoResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Router			/api/v1/fuel-entries/{id}/photos/{child_id}/set-primary [post]
func docSetFuelPhotoPrimary() {}
