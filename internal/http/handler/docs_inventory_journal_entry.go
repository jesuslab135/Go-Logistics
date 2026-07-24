package handler

// docListInventoryJournalEntries godoc
//
//	@Summary	List inventory-journal-entries
//	@Tags	inventory-journal-entries
//	@Security	BearerAuth
//	@Produce	json
//	@Param	page	query	int	false	"Page"
//	@Param	page_size	query	int	false	"Size"
//	@Success	200	{object}	dto.InventoryJournalEntryPage
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inventory-journal-entries [get]
func docListInventoryJournalEntries() {}

// docCreateInventoryJournalEntry godoc
//
//	@Summary	Create inventory-journal-entries
//	@Tags	inventory-journal-entries
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateInventoryJournalEntryRequest	true	"body"
//	@Success	201	{object}	dto.InventoryJournalEntryResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/inventory-journal-entries [post]
func docCreateInventoryJournalEntry() {}

// docGetInventoryJournalEntry godoc
//
//	@Summary	Get inventory-journal-entries
//	@Tags	inventory-journal-entries
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.InventoryJournalEntryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inventory-journal-entries/{id} [get]
func docGetInventoryJournalEntry() {}

// docUpdateInventoryJournalEntry godoc
//
//	@Summary	Update inventory-journal-entries
//	@Tags	inventory-journal-entries
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateInventoryJournalEntryRequest	true	"body"
//	@Success	200	{object}	dto.InventoryJournalEntryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/inventory-journal-entries/{id} [put]
func docUpdateInventoryJournalEntry() {}

// docDeleteInventoryJournalEntry godoc
//
//	@Summary	Delete inventory-journal-entries
//	@Tags	inventory-journal-entries
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/inventory-journal-entries/{id} [delete]
func docDeleteInventoryJournalEntry() {}
