package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/domain/tire"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
)

type TireAssignmentActionHandler struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewTireAssignmentActionHandler(q *gen.Queries, pool *pgxpool.Pool) *TireAssignmentActionHandler {
	return &TireAssignmentActionHandler{q: q, pool: pool}
}

// Approve godoc
//
//	@Summary		Approve a tire assignment request
//	@Description	Resolves the request, creates the installation, appends an INSTALL mount log and updates the tire — all in one transaction.
//	@Tags			tire-assignment-requests
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		int											true	"Request id"
//	@Param			installation	body	dto.ApproveTireAssignmentRequestRequest	true	"Installation readings"
//	@Success		200			{object}	dto.TireAssignmentRequestResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Failure		409			{object}	dto.ErrorResponse
//	@Failure		422			{object}	dto.ErrorResponse
//	@Router			/api/v1/tire-assignment-requests/{id}/approve [post]
func (h *TireAssignmentActionHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid id"))
		return
	}

	var req dto.ApproveTireAssignmentRequestRequest
	if err := bindJSONValidated(c, &req); err != nil {
		apierr.Abort(c, err)
		return
	}

	ctx := c.Request.Context()
	company := middleware.CompanyFromContext(ctx)
	employee := middleware.EmployeeFromContext(ctx)
	now := time.Now().UTC()

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	defer tx.Rollback(ctx)
	qtx := h.q.WithTx(tx)

	// Both rows are locked before anything is validated, so a concurrent
	// approval of the same request or tire waits rather than racing us.
	request, err := qtx.GetTireAssignmentRequestForUpdate(ctx, gen.GetTireAssignmentRequestForUpdateParams{
		ID:        id,
		CompanyID: company,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if request.State != tire.RequestPending {
		apierr.Abort(c, apierr.New(http.StatusConflict, "request_already_resolved",
			fmt.Sprintf("request is already %s", request.State)))
		return
	}

	t, err := qtx.GetTireForUpdate(ctx, gen.GetTireForUpdateParams{ID: request.TireID, CompanyID: company})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	switch t.Status {
	case tire.StatusInStock:
		// the only mountable state
	case tire.StatusMounted:
		apierr.Abort(c, apierr.New(http.StatusConflict, "tire_already_mounted", "the tire is already mounted"))
		return
	default:
		apierr.Abort(c, apierr.New(http.StatusConflict, "tire_unavailable",
			fmt.Sprintf("the tire is %s and cannot be mounted", t.Status)))
		return
	}

	// tire_installation carries UNIQUE(tire_id) and UNIQUE(vehicle_id,
	// position_code); checking first turns what would be a 500 from the
	// constraint into a diagnosable conflict.
	installed, err := qtx.TireHasInstallation(ctx, request.TireID)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if installed {
		apierr.Abort(c, apierr.New(http.StatusConflict, "tire_already_mounted", "the tire already has an active installation"))
		return
	}
	occupied, err := qtx.TirePositionOccupied(ctx, gen.TirePositionOccupiedParams{
		VehicleID:    request.VehicleID,
		PositionCode: request.PositionCode,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if occupied {
		apierr.Abort(c, apierr.New(http.StatusConflict, "position_occupied",
			fmt.Sprintf("position %q on vehicle %d already carries a tire", request.PositionCode, request.VehicleID)))
		return
	}

	resolved, err := qtx.ResolveTireAssignmentRequest(ctx, gen.ResolveTireAssignmentRequestParams{
		ID:           id,
		CompanyID:    company,
		State:        tire.RequestApproved,
		ApprovedByID: authorFromContext(ctx),
		ResolvedAt:   &now,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	notes := fmt.Sprintf("Approved from tire assignment request #%d", id)
	if req.Notes != nil && *req.Notes != "" {
		notes = *req.Notes
	}
	if _, err := qtx.CreateTireInstallation(ctx, gen.CreateTireInstallationParams{
		ParentID:                 request.TireID,
		CompanyID:                company,
		VehicleID:                request.VehicleID,
		PositionCode:             request.PositionCode,
		InstallDate:              now,
		OdometerAtInstall:        req.OdometerAtInstall,
		TreadDepthAtInstall32nds: req.TreadDepthAtInstall32nds,
		PsiAtInstall:             req.PsiAtInstall,
		InstalledByID:            authorFromContext(ctx),
		Notes:                    notes,
	}); err != nil {
		apierr.Abort(c, err)
		return
	}

	odometer := req.OdometerAtInstall
	if _, err := qtx.CreateTireMountLog(ctx, gen.CreateTireMountLogParams{
		ParentID:        request.TireID,
		CompanyID:       company,
		VehicleID:       request.VehicleID,
		PositionCode:    request.PositionCode,
		EventType:       tire.EventInstall,
		EventDate:       now,
		Odometer:        &odometer,
		TreadDepth32nds: req.TreadDepthAtInstall32nds,
		Psi:             req.PsiAtInstall,
		PerformedByID:   authorFromContext(ctx),
		Reason:          fmt.Sprintf("Approved tire assignment request #%d", id),
	}); err != nil {
		apierr.Abort(c, err)
		return
	}

	// Django left these stale; the registry reads them, so they are updated
	// inside the same transaction as the installation they describe.
	if err := qtx.MountTire(ctx, gen.MountTireParams{
		ID:                  request.TireID,
		CurrentVehicleID:    &request.VehicleID,
		CurrentPositionCode: request.PositionCode,
	}); err != nil {
		apierr.Abort(c, err)
		return
	}

	// Tell the requester their tire was approved. Sent in the same transaction
	// as the approval, so a notification can never describe a rollback.
	if resolved.RequestedByID != nil && *resolved.RequestedByID != employee {
		if err := qtx.CreateNotification(ctx, gen.CreateNotificationParams{
			CompanyID:  company,
			EmployeeID: *resolved.RequestedByID,
			Kind:       NotificationTireAssignmentApproved,
			Title:      "Tire assignment approved",
			Body:       fmt.Sprintf("Request #%d was approved.", id),
			Url:        fmt.Sprintf("/tire-assignment-requests/%d", id),
			CreatedAt:  now,
		}); err != nil {
			apierr.Abort(c, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, toTireAssignmentRequestResponse(resolved))
}
