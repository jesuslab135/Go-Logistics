package handler

// docListServiceReminders godoc
//
//	@Summary	List service-reminders
//	@Tags	service-reminders
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.ServiceReminderPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-reminders [get]
func docListServiceReminders() {}

// docCreateServiceReminder godoc
//
//	@Summary	Create service-reminders
//	@Tags	service-reminders
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateServiceReminderRequest	true	"body"
//	@Success	201	{object}	dto.ServiceReminderResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-reminders [post]
func docCreateServiceReminder() {}

// docGetServiceReminder godoc
//
//	@Summary	Get service-reminders
//	@Tags	service-reminders
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.ServiceReminderResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-reminders/{id} [get]
func docGetServiceReminder() {}

// docUpdateServiceReminder godoc
//
//	@Summary	Update service-reminders
//	@Tags	service-reminders
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateServiceReminderRequest	true	"body"
//	@Success	200	{object}	dto.ServiceReminderResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-reminders/{id} [put]
func docUpdateServiceReminder() {}

// docDeleteServiceReminder godoc
//
//	@Summary	Delete service-reminders
//	@Tags	service-reminders
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/service-reminders/{id} [delete]
func docDeleteServiceReminder() {}
