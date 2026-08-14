package handler

// docListLocations godoc
//
//	@Summary	List locations
//	@Tags	locations
//	@Security	BearerAuth
//	@Produce	json
//	@Param	limit	query	int	false	"Page size"
//	@Param	offset	query	int	false	"Offset"
//	@Success	200	{object}	dto.LocationPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/locations [get]
func docListLocations() {}

// docCreateLocation godoc
//
//	@Summary	Create locations
//	@Tags	locations
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateLocationRequest	true	"body"
//	@Success	201	{object}	dto.LocationResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/locations [post]
func docCreateLocation() {}

// docGetLocation godoc
//
//	@Summary	Get locations
//	@Tags	locations
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.LocationResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/locations/{id} [get]
func docGetLocation() {}

// docUpdateLocation godoc
//
//	@Summary	Update locations
//	@Tags	locations
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateLocationRequest	true	"body"
//	@Success	200	{object}	dto.LocationResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/locations/{id} [put]
func docUpdateLocation() {}

// docDeleteLocation godoc
//
//	@Summary	Delete locations
//	@Tags	locations
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/locations/{id} [delete]
func docDeleteLocation() {}
