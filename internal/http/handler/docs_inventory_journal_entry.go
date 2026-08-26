package handler

// The ledger is append-only. PUT and DELETE still route — they answer 405
// `route_retired` naming the reversal route — but they are not documented as
// operations, because they are not ones a client should call. The reversal
// route documents itself, on the handler.

// docCreateInventoryJournalEntry godoc
//
//	@Summary		Create inventory-journal-entries
//	@Description	Files a stock movement and applies it: part_inventory.available_quantity moves by adjustment_quantity in the same transaction. adjustment_quantity is signed — negative consumes, positive receives. previous_quantity and current_quantity are server-derived and returned, not sent; user_id is stamped from the caller.
//	@Tags			inventory-journal-entries
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body	dto.CreateInventoryJournalEntryRequest	true	"body"
//	@Success		201		{object}	dto.InventoryJournalEntryResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/inventory-journal-entries [post]
func docCreateInventoryJournalEntry() {}

// docGetInventoryJournalEntry godoc
//
//	@Summary	Get inventory-journal-entries
//	@Tags		inventory-journal-entries
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		int	true	"id"
//	@Success	200	{object}	dto.InventoryJournalEntryResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/api/v1/inventory-journal-entries/{id} [get]
func docGetInventoryJournalEntry() {}
