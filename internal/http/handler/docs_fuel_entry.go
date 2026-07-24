package handler

// docNListFuelEntries godoc
//
//	@Summary	List fuel-entries
//	@Tags	fuel-entries
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.FuelEntryPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/fuel-entries [get]
func docNListFuelEntries() {}

// docNCreateFuelEntry godoc
//
//	@Summary	Create fuel-entries
//	@Tags	fuel-entries
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateFuelEntryRequest	true	"body"
//	@Success	201	{object}	dto.FuelEntryResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/fuel-entries [post]
func docNCreateFuelEntry() {}

// docNGetFuelEntry godoc
//
//	@Summary	Get fuel-entries
//	@Tags	fuel-entries
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.FuelEntryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/fuel-entries/{child_id} [get]
func docNGetFuelEntry() {}

// docNUpdateFuelEntry godoc
//
//	@Summary	Update fuel-entries
//	@Tags	fuel-entries
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateFuelEntryRequest	true	"body"
//	@Success	200	{object}	dto.FuelEntryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/fuel-entries/{child_id} [put]
func docNUpdateFuelEntry() {}

// docNDeleteFuelEntry godoc
//
//	@Summary	Delete fuel-entries
//	@Tags	fuel-entries
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/assets/{id}/fuel-entries/{child_id} [delete]
func docNDeleteFuelEntry() {}
