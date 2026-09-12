package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/dbctx"
	"fleet/internal/platform/paginate"
)

// InventoryJournalEntryStore owns the inventory ledger. An entry is a stock
// movement: creating one moves part_inventory.available_quantity in the same
// transaction, which is what previous_quantity / adjustment_quantity /
// current_quantity have always claimed to describe.
//
// Django's Excel importer wrote them exactly that way — read the quantity, add
// the adjustment, save, then record the three values around the move — but its
// ordinary endpoint, and this port's, let the client supply all three and moved
// nothing at all. The columns recorded a movement the inventory never made, and
// the two drifted with no warning.
//
// The ledger is append-only, so there is no Update or Delete: a mistake is
// corrected by an offsetting entry, which is a movement in its own right and is
// visible as one.
type InventoryJournalEntryStore struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewInventoryJournalEntryStore(q *gen.Queries, pool *pgxpool.Pool) *InventoryJournalEntryStore {
	return &InventoryJournalEntryStore{q: q, pool: pool}
}

func (s *InventoryJournalEntryStore) List(ctx context.Context, p paginate.Params) ([]dto.InventoryJournalEntryResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListInventoryJournalEntries(ctx, gen.ListInventoryJournalEntriesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountInventoryJournalEntries(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InventoryJournalEntryResponse, len(rows))
	for i, r := range rows {
		out[i] = toInventoryJournalEntryResponse(r)
	}
	return out, total, nil
}

func (s *InventoryJournalEntryStore) Get(ctx context.Context, id int64) (dto.InventoryJournalEntryResponse, error) {
	r, err := s.q.GetInventoryJournalEntry(ctx, gen.GetInventoryJournalEntryParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	return toInventoryJournalEntryResponse(r), nil
}

// Create moves the stock and records the movement, together.
//
// previous_quantity and current_quantity are no longer read from the request.
// They are the quantities before and after this adjustment, which only the
// server can know: a client's copy is whatever it last read, and honouring it
// would let two concurrent adjustments each record a "previous" that was already
// stale. The request supplies adjustment_quantity, and the row records what
// actually happened around it.
func (s *InventoryJournalEntryStore) Create(ctx context.Context, in dto.CreateInventoryJournalEntryRequest) (dto.InventoryJournalEntryResponse, error) {
	company := middleware.CompanyFromContext(ctx)

	tx, err := dbctx.Begin(ctx, s.pool)
	if err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	entry, err := applyAdjustment(ctx, qtx, adjustment{
		company:              company,
		partID:               in.PartID,
		partLocationDetailID: in.PartLocationDetailID,
		quantity:             in.AdjustmentQuantity,
		unitCost:             in.UnitCost,
		reasonID:             in.ReasonID,
		workOrderID:          in.WorkOrderID,
		purchaseOrderLineID:  in.PurchaseOrderLineID,
		vendorID:             in.VendorID,
		adjustmentType:       orDefault(in.AdjustmentType, "manual"),
		transferLocationID:   in.TransferPartLocationID,
		notes:                in.Notes,
	})
	if err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	return toInventoryJournalEntryResponse(entry), nil
}

// adjustment is one stock movement to apply. It exists so Create and Reverse
// share the read-lock-move-record sequence rather than each writing their own
// and drifting.
type adjustment struct {
	company              int64
	partID               int64
	partLocationDetailID int64
	quantity             decimal.Decimal
	unitCost             decimal.Decimal
	reasonID             *int64
	workOrderID          *int64
	purchaseOrderLineID  *int64
	vendorID             *int64
	adjustmentType       string
	transferLocationID   *int64
	notes                string
	reversalOf           *int64
}

// applyAdjustment locks the stock row, moves it, and records the movement. The
// order matters: the lock is taken before the quantity is read, so the
// previous/current pair the entry records is the pair this transaction actually
// saw.
func applyAdjustment(ctx context.Context, qtx *gen.Queries, a adjustment) (gen.InventoryJournalEntry, error) {
	stock, err := qtx.LockPartInventory(ctx, gen.LockPartInventoryParams{
		ID:        a.partLocationDetailID,
		CompanyID: a.company,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return gen.InventoryJournalEntry{}, apierr.Validation(map[string]string{
				"part_location_detail_id": "no stock record for this id in your company",
			})
		}
		return gen.InventoryJournalEntry{}, err
	}
	if stock.PartID != a.partID {
		return gen.InventoryJournalEntry{}, apierr.Validation(map[string]string{
			"part_id": "does not match the part this stock record belongs to",
		})
	}

	now := time.Now().UTC()
	moved, err := qtx.ApplyPartInventoryAdjustment(ctx, gen.ApplyPartInventoryAdjustmentParams{
		ID:                 a.partLocationDetailID,
		AdjustmentQuantity: a.quantity,
		UpdatedAt:          &now,
	})
	if err != nil {
		return gen.InventoryJournalEntry{}, err
	}

	return qtx.CreateInventoryJournalEntry(ctx, gen.CreateInventoryJournalEntryParams{
		CompanyID:              a.company,
		PartID:                 a.partID,
		PartLocationDetailID:   a.partLocationDetailID,
		UserID:                 authorFromContext(ctx),
		PreviousQuantity:       stock.AvailableQuantity,
		AdjustmentQuantity:     a.quantity,
		CurrentQuantity:        moved.AvailableQuantity,
		UnitCost:               a.unitCost,
		ReasonID:               a.reasonID,
		WorkOrderID:            a.workOrderID,
		PurchaseOrderLineID:    a.purchaseOrderLineID,
		VendorID:               a.vendorID,
		AdjustmentType:         a.adjustmentType,
		TransferPartLocationID: a.transferLocationID,
		Notes:                  a.notes,
		ReversalOfID:           a.reversalOf,
		CreatedAt:              now,
	})
}

func toInventoryJournalEntryResponse(r gen.InventoryJournalEntry) dto.InventoryJournalEntryResponse {
	return dto.InventoryJournalEntryResponse{
		ID:                     r.ID,
		CompanyID:              r.CompanyID,
		PartID:                 r.PartID,
		PartLocationDetailID:   r.PartLocationDetailID,
		UserID:                 r.UserID,
		PreviousQuantity:       r.PreviousQuantity,
		AdjustmentQuantity:     r.AdjustmentQuantity,
		CurrentQuantity:        r.CurrentQuantity,
		UnitCost:               r.UnitCost,
		ReasonID:               r.ReasonID,
		WorkOrderID:            r.WorkOrderID,
		PurchaseOrderLineID:    r.PurchaseOrderLineID,
		VendorID:               r.VendorID,
		AdjustmentType:         r.AdjustmentType,
		TransferPartLocationID: r.TransferPartLocationID,
		Notes:                  r.Notes,
		ReversalOfID:           r.ReversalOfID,
		CreatedAt:              r.CreatedAt,
	}
}

// InventoryJournalEntryHandler serves the two routes the generic CRUD handler
// cannot: reversal, and a 405 for the retired write verbs.
type InventoryJournalEntryHandler struct {
	store *InventoryJournalEntryStore
}

func NewInventoryJournalEntryHandler(store *InventoryJournalEntryStore) *InventoryJournalEntryHandler {
	return &InventoryJournalEntryHandler{store: store}
}

const journalRetiredMessage = "the inventory ledger is append-only: correct an entry with POST /api/v1/inventory-journal-entries/{id}/reverse, which files an offsetting entry and moves the stock back"

// Retired answers the removed edit and delete verbs.
func (h *InventoryJournalEntryHandler) Retired(c *gin.Context) {
	apierr.Abort(c, apierr.New(http.StatusMethodNotAllowed, "route_retired", journalRetiredMessage))
}

// Reverse godoc
//
//	@Summary		Reverse an inventory journal entry
//	@Description	Files an offsetting entry and moves the stock back, in one transaction. The original is left untouched — the correction is itself a movement, and both rows stay visible. An entry can be reversed once: a second attempt is 409, because repeated reversals would drive the stock arbitrarily far from what the ledger sums to. A reversal cannot itself be reversed; reverse the original instead.
//	@Tags			inventory-journal-entries
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Entry id"
//	@Success		201	{object}	dto.InventoryJournalEntryResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/api/v1/inventory-journal-entries/{id}/reverse [post]
func (h *InventoryJournalEntryHandler) Reverse(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid id"))
		return
	}

	out, err := h.store.Reverse(c.Request.Context(), id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// Reverse files the offsetting entry. It is a create, not an undo: the original
// row is never touched, so what the ledger shows is that a movement happened and
// was then reversed — which is what did happen.
func (s *InventoryJournalEntryStore) Reverse(ctx context.Context, id int64) (dto.InventoryJournalEntryResponse, error) {
	company := middleware.CompanyFromContext(ctx)

	tx, err := dbctx.Begin(ctx, s.pool)
	if err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	original, err := qtx.GetInventoryJournalEntry(ctx, gen.GetInventoryJournalEntryParams{ID: id, CompanyID: company})
	if err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	if original.ReversalOfID != nil {
		return dto.InventoryJournalEntryResponse{}, apierr.Conflict(
			"this entry is itself a reversal; reverse the original entry instead")
	}

	// The unique index enforces this too, but checking here turns a constraint
	// violation into an error that says which rule was broken.
	if _, err := qtx.GetInventoryJournalReversal(ctx, &id); err == nil {
		return dto.InventoryJournalEntryResponse{}, apierr.Conflict("this entry has already been reversed")
	} else if err != pgx.ErrNoRows {
		return dto.InventoryJournalEntryResponse{}, err
	}

	entry, err := applyAdjustment(ctx, qtx, adjustment{
		company:              company,
		partID:               original.PartID,
		partLocationDetailID: original.PartLocationDetailID,
		quantity:             original.AdjustmentQuantity.Neg(),
		unitCost:             original.UnitCost,
		reasonID:             original.ReasonID,
		workOrderID:          original.WorkOrderID,
		purchaseOrderLineID:  original.PurchaseOrderLineID,
		vendorID:             original.VendorID,
		adjustmentType:       original.AdjustmentType,
		transferLocationID:   original.TransferPartLocationID,
		notes:                "Reversal of entry " + strconv.FormatInt(original.ID, 10),
		reversalOf:           &original.ID,
	})
	if err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	return toInventoryJournalEntryResponse(entry), nil
}
