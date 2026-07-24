package handler

// docListWeeklyMileageGoals godoc
//
//	@Summary	List weekly-mileage-goals
//	@Tags	weekly-mileage-goals
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.WeeklyMileageGoalPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/weekly-mileage-goals [get]
func docListWeeklyMileageGoals() {}

// docCreateWeeklyMileageGoal godoc
//
//	@Summary	Create weekly-mileage-goals
//	@Tags	weekly-mileage-goals
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateWeeklyMileageGoalRequest	true	"body"
//	@Success	201	{object}	dto.WeeklyMileageGoalResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/weekly-mileage-goals [post]
func docCreateWeeklyMileageGoal() {}

// docGetWeeklyMileageGoal godoc
//
//	@Summary	Get weekly-mileage-goals
//	@Tags	weekly-mileage-goals
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.WeeklyMileageGoalResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/weekly-mileage-goals/{id} [get]
func docGetWeeklyMileageGoal() {}

// docUpdateWeeklyMileageGoal godoc
//
//	@Summary	Update weekly-mileage-goals
//	@Tags	weekly-mileage-goals
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateWeeklyMileageGoalRequest	true	"body"
//	@Success	200	{object}	dto.WeeklyMileageGoalResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/weekly-mileage-goals/{id} [put]
func docUpdateWeeklyMileageGoal() {}

// docDeleteWeeklyMileageGoal godoc
//
//	@Summary	Delete weekly-mileage-goals
//	@Tags	weekly-mileage-goals
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/weekly-mileage-goals/{id} [delete]
func docDeleteWeeklyMileageGoal() {}
