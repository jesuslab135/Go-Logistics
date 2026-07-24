package handler

// docNListTireInspections godoc
//
//	@Summary	List tire-inspections
//	@Tags	tire-inspections
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Success	200	{object}	dto.TireInspectionPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/inspections [get]
func docNListTireInspections() {}

// docNCreateTireInspection godoc
//
//	@Summary	Create tire-inspections
//	@Tags	tire-inspections
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	body	body	dto.CreateTireInspectionRequest	true	"body"
//	@Success	201	{object}	dto.TireInspectionResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/inspections [post]
func docNCreateTireInspection() {}

// docNGetTireInspection godoc
//
//	@Summary	Get tire-inspections
//	@Tags	tire-inspections
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	200	{object}	dto.TireInspectionResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/inspections/{child_id} [get]
func docNGetTireInspection() {}

// docNUpdateTireInspection godoc
//
//	@Summary	Update tire-inspections
//	@Tags	tire-inspections
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Param	body	body	dto.UpdateTireInspectionRequest	true	"body"
//	@Success	200	{object}	dto.TireInspectionResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/inspections/{child_id} [put]
func docNUpdateTireInspection() {}

// docNDeleteTireInspection godoc
//
//	@Summary	Delete tire-inspections
//	@Tags	tire-inspections
//	@Security	BearerAuth
//	@Param	id	path	int	true	"parent id"
//	@Param	child_id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/tires/{id}/inspections/{child_id} [delete]
func docNDeleteTireInspection() {}
