package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"fleet/internal/auth"
	"fleet/internal/db/gen"
	"fleet/internal/domain/purchaseorder"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/crud"
	"fleet/internal/platform/storage"
)

// Deps are the collaborators the HTTP layer needs. cmd/api constructs these
// (pgx pool -> gen.Queries, token service, etc.) and calls NewRouter.
type Deps struct {
	Queries     *gen.Queries
	Pool        *pgxpool.Pool
	Tokens      *auth.TokenService
	Verifier    CredentialVerifier
	Storage     storage.Storage
	Logger      *slog.Logger
	CORSOrigins []string
	Production  bool
	// Swagger serves the interactive docs at /swagger/*. Independent of
	// Production: the docs are a disclosure choice, not a runtime mode.
	Swagger bool
}

// NewRouter assembles the Gin engine: global middleware, public auth/health
// routes, and the authenticated /api/v1 resource routes.
func NewRouter(d Deps) *gin.Engine {
	if d.Logger == nil {
		d.Logger = slog.Default()
	}
	if d.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Logger(d.Logger),
		middleware.Recovery(d.Logger),
		middleware.CORS(d.CORSOrigins),
	)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Interactive API docs. Off by default in production (SWAGGER_ENABLED)
	// because serving them publishes the whole schema.
	if d.Swagger {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	authH := NewAuthHandler(d.Tokens, d.Verifier, d.Queries)
	r.POST("/auth/login", authH.Login)
	r.POST("/auth/refresh", authH.Refresh)
	r.POST("/auth/logout", authH.Logout)
	// Switching tenants needs a valid access token but deliberately no company
	// gate — the caller is leaving the company the token is scoped to.
	r.POST("/auth/switch-company", middleware.Auth(d.Tokens), authH.SwitchCompany)

	api := r.Group("/api/v1")
	api.Use(middleware.Auth(d.Tokens), middleware.RequireIdentity(NewIdentityLoader(d.Queries)))

	// Authorization mirrors api/permissions.py: every resource requires company
	// membership, and most additionally require the caller's role to grant the
	// module the resource belonged to in Django (its viewset's module_name).
	// Groups below are named for that module.
	member := api.Group("", middleware.RequireCompanyMember())
	// Archived rows are hidden from every list unless the caller opts in.
	member.Use(WithIncludeArchived())

	registerCompanyRoutes(api, member, d)
	registerAdminRoutes(api, d)
	registerAccountEmployeeRoutes(api, d)

	assets := member.Group("", middleware.RequireModule("assets"))
	tires := member.Group("", middleware.RequireModule("tires"))
	parts := member.Group("", middleware.RequireModule("parts"))
	inventory := member.Group("", middleware.RequireModule("inventory"))
	workOrders := member.Group("", middleware.RequireModule("work_orders"))
	issues := member.Group("", middleware.RequireModule("issues"))
	service := member.Group("", middleware.RequireModule("service"))
	fuel := member.Group("", middleware.RequireModule("fuel"))
	inspections := member.Group("", middleware.RequireModule("inspections"))
	vendors := member.Group("", middleware.RequireModule("vendors"))
	warranties := member.Group("", middleware.RequireModule("warranties"))
	mileageGoals := member.Group("", middleware.RequireModule("mileage_goals"))
	employees := member.Group("", middleware.RequireModule("employees"))
	purchaseOrders := member.Group("", middleware.RequireModule("purchase_orders"))
	// Django gated roles on IsAdminRole rather than on a module entry.
	roles := member.Group("", middleware.RequireAdminRole())

	// Custom field definitions describe what may go in each resource's
	// custom_fields document. Admin-gated: they shape everyone's forms.
	customFields := member.Group("", middleware.RequireAdminRole(), WithResourceFilter())
	crud.NewHandler[dto.CustomFieldDefinitionResponse, dto.CreateCustomFieldDefinitionRequest, dto.UpdateCustomFieldDefinitionRequest](
		NewCustomFieldDefinitionStore(d.Queries)).Register(customFields, "/custom-field-definitions")

	// Archive/restore for the five catalogs that carry archived_at. A referenced
	// record cannot be deleted; archiving is what retires it instead.
	registerArchiveRoutes(assets, parts, vendors, service, inspections, d.Queries)

	// Collections whose filters vary per request take their List verb from
	// here instead of from the generic CRUD handler.
	lists := NewFilteredListHandler(d.Pool, d.Storage)

	// Phase 2: assets
	registerCrudWithList(assets, "/assets",
		crud.NewHandler[dto.AssetResponse, dto.CreateAssetRequest, dto.UpdateAssetRequest](
			NewAssetStore(d.Queries, d.Pool, d.Storage, d.Logger)), lists.Assets)
	// Read-only reverse view of the nested /assets/{id}/trailer-assignments;
	// mutations stay on the nested route that owns the parent scope.
	assets.GET("/asset-trailer-assignments", lists.AssetTrailerAssignments)
	// Cheap aggregate counts alongside the list, so a registry can draw its
	// tab/facet chips from one grouped query instead of one list call per bucket.
	assets.GET("/assets/facets", lists.AssetFacets)

	// Phase 3: parts & inventory
	crud.NewHandler[dto.PartCategoryResponse, dto.CreatePartCategoryRequest, dto.UpdatePartCategoryRequest](
		NewPartCategoryStore(d.Queries)).Register(parts, "/part-categories")
	crud.NewHandler[dto.PartManufacturerResponse, dto.CreatePartManufacturerRequest, dto.UpdatePartManufacturerRequest](
		NewPartManufacturerStore(d.Queries)).Register(parts, "/part-manufacturers")
	crud.NewHandler[dto.MeasurementUnitResponse, dto.CreateMeasurementUnitRequest, dto.UpdateMeasurementUnitRequest](
		NewMeasurementUnitStore(d.Queries)).Register(inventory, "/measurement-units")
	crud.NewHandler[dto.PartResponse, dto.CreatePartRequest, dto.UpdatePartRequest](
		NewPartStore(d.Queries)).Register(parts, "/parts")
	crud.NewHandler[dto.PartLocationResponse, dto.CreatePartLocationRequest, dto.UpdatePartLocationRequest](
		NewPartLocationStore(d.Queries)).Register(inventory, "/part-locations")
	crud.NewHandler[dto.InventoryAdjustmentReasonResponse, dto.CreateInventoryAdjustmentReasonRequest, dto.UpdateInventoryAdjustmentReasonRequest](
		NewInventoryAdjustmentReasonStore(d.Queries)).Register(inventory, "/inventory-adjustment-reasons")
	// Flat, filterable views of rows that are otherwise reachable only under a
	// parent. Forms need them as foreign-key sources; mutations stay nested.
	inventory.GET("/part-inventory", lists.PartInventory)
	inventory.GET("/purchase-order-line-items", lists.PurchaseOrderLineItems)
	registerInventoryJournalRoutes(inventory, d, lists.InventoryJournalEntries)

	// Phase 4: vendors, work orders & issues
	crud.NewHandler[dto.VendorResponse, dto.CreateVendorRequest, dto.UpdateVendorRequest](NewVendorStore(d.Queries)).Register(vendors, "/vendors")
	crud.NewHandler[dto.WorkOrderStatusResponse, dto.CreateWorkOrderStatusRequest, dto.UpdateWorkOrderStatusRequest](NewWorkOrderStatusStore(d.Queries)).Register(workOrders, "/work-order-statuses")
	// Django's LocationViewSet required membership only, no module entry.
	crud.NewHandler[dto.LocationResponse, dto.CreateLocationRequest, dto.UpdateLocationRequest](NewLocationStore(d.Queries)).Register(member, "/locations")
	registerCrudWithList(workOrders, "/work-orders",
		crud.NewHandler[dto.WorkOrderResponse, dto.CreateWorkOrderRequest, dto.UpdateWorkOrderRequest](NewWorkOrderStore(d.Queries, d.Pool)), lists.WorkOrders)
	workOrders.GET("/work-orders/facets", lists.WorkOrderFacets)
	// Totals are computed; this is the audited way to depart from the formula.
	overrides := NewTotalOverrideHandler(d.Queries, d.Pool)
	workOrders.POST("/work-orders/:id/override-total", overrides.WorkOrder)
	service.POST("/service-entries/:id/override-total", overrides.ServiceEntry)
	registerCrudWithList(issues, "/issues",
		crud.NewHandler[dto.IssueResponse, dto.CreateIssueRequest, dto.UpdateIssueRequest](NewIssueStore(d.Queries)), lists.Issues)
	issues.GET("/issues/facets", lists.IssueFacets)
	crud.NewHandler[dto.IssuePriorityResponse, dto.CreateIssuePriorityRequest, dto.UpdateIssuePriorityRequest](NewIssuePriorityStore(d.Queries)).Register(issues, "/issue-priorities")
	crud.NewHandler[dto.FaultResponse, dto.CreateFaultRequest, dto.UpdateFaultRequest](NewFaultStore(d.Queries)).Register(issues, "/faults")

	// Phase 5: purchase orders & service
	registerPurchaseOrderRoutes(purchaseOrders, d, lists.PurchaseOrders)
	crud.NewHandler[dto.ServiceTaskResponse, dto.CreateServiceTaskRequest, dto.UpdateServiceTaskRequest](NewServiceTaskStore(d.Queries)).Register(service, "/service-tasks")
	crud.NewHandler[dto.ServiceReminderResponse, dto.CreateServiceReminderRequest, dto.UpdateServiceReminderRequest](NewServiceReminderStore(d.Queries)).Register(service, "/service-reminders")
	crud.NewHandler[dto.ServiceEntryResponse, dto.CreateServiceEntryRequest, dto.UpdateServiceEntryRequest](NewServiceEntryStore(d.Queries, d.Pool)).Register(service, "/service-entries")

	// Phase 6: tires
	crud.NewHandler[dto.TireResponse, dto.CreateTireRequest, dto.UpdateTireRequest](NewTireStore(d.Queries)).Register(tires, "/tires")
	crud.NewHandler[dto.TireModelResponse, dto.CreateTireModelRequest, dto.UpdateTireModelRequest](NewTireModelStore(d.Queries)).Register(tires, "/tire-models")
	crud.NewHandler[dto.AxleTemplateResponse, dto.CreateAxleTemplateRequest, dto.UpdateAxleTemplateRequest](NewAxleTemplateStore(d.Queries)).Register(tires, "/axle-templates")
	// Django's TireAssignmentRequestViewSet required membership only; its
	// approve/reject actions carried the extra warehouse-role check.
	registerCrudWithList(member, "/tire-assignment-requests",
		crud.NewHandler[dto.TireAssignmentRequestResponse, dto.CreateTireAssignmentRequestRequest, dto.UpdateTireAssignmentRequestRequest](NewTireAssignmentRequestStore(d.Queries, d.Pool)), lists.TireAssignmentRequests)
	// Approving is the warehouse's act, not an ordinary write: Django gated it
	// on tire_approvals/approve, which admins bypass.
	member.POST("/tire-assignment-requests/:id/approve",
		middleware.RequireAction("tire_approvals", "approve"),
		NewTireAssignmentActionHandler(d.Queries, d.Pool).Approve)

	// Phase 7: fuel, inspections, org & misc
	crud.NewHandler[dto.FuelTypeResponse, dto.CreateFuelTypeRequest, dto.UpdateFuelTypeRequest](NewFuelTypeStore(d.Queries)).Register(fuel, "/fuel-types")
	// The trailer classification vocabulary, a Settings catalog like the others.
	crud.NewHandler[dto.TrailerClassificationResponse, dto.CreateTrailerClassificationRequest, dto.UpdateTrailerClassificationRequest](
		NewTrailerClassificationStore(d.Queries)).Register(assets, "/trailer-classifications")
	crud.NewHandler[dto.InspectionFormResponse, dto.CreateInspectionFormRequest, dto.UpdateInspectionFormRequest](NewInspectionFormStore(d.Queries)).Register(inspections, "/inspection-forms")
	crud.NewHandler[dto.InspectionSubmissionResponse, dto.CreateInspectionSubmissionRequest, dto.UpdateInspectionSubmissionRequest](NewInspectionSubmissionStore(d.Queries, d.Storage, d.Logger)).Register(inspections, "/inspection-submissions")
	registerCrudWithList(assets, "/media",
		crud.NewHandler[dto.MediumResponse, dto.CreateMediumRequest, dto.UpdateMediumRequest](NewMediumStore(d.Queries, d.Storage, d.Logger)), lists.Media)
	registerCrudWithList(issues, "/comments",
		crud.NewHandler[dto.CommentResponse, dto.CreateCommentRequest, dto.UpdateCommentRequest](NewCommentStore(d.Queries)), lists.Comments)
	crud.NewHandler[dto.WarrantyResponse, dto.CreateWarrantyRequest, dto.UpdateWarrantyRequest](NewWarrantyStore(d.Queries)).Register(warranties, "/warranties")
	crud.NewHandler[dto.WeeklyMileageGoalResponse, dto.CreateWeeklyMileageGoalRequest, dto.UpdateWeeklyMileageGoalRequest](NewWeeklyMileageGoalStore(d.Queries)).Register(mileageGoals, "/weekly-mileage-goals")
	crud.NewHandler[dto.RoleResponse, dto.CreateRoleRequest, dto.UpdateRoleRequest](NewRoleStore(d.Queries)).Register(roles, "/roles")
	// employee: m2m-scoped, password_hash excluded, transactional create
	crud.NewHandler[dto.EmployeeResponse, dto.CreateEmployeeRequest, dto.UpdateEmployeeRequest](NewEmployeeStore(d.Queries, d.Pool)).Register(employees, "/employees")
	// Credential provisioning is an admin act, so it sits above the module's
	// ordinary write permission.
	employees.POST("/employees/:id/set-password", middleware.RequireAdminRole(), NewEmployeeActionHandler(d.Queries).SetPassword)

	// Phase 1: asset catalogs & vehicle models (tenant-scoped)
	crud.NewHandler[dto.AssetTypeResponse, dto.CreateAssetTypeRequest, dto.UpdateAssetTypeRequest](
		NewAssetTypeStore(d.Queries)).Register(assets, "/asset-types")
	crud.NewHandler[dto.AssetStatusResponse, dto.CreateAssetStatusRequest, dto.UpdateAssetStatusRequest](
		NewAssetStatusStore(d.Queries)).Register(assets, "/asset-statuses")
	registerCrudWithList(assets, "/catalog-options",
		crud.NewHandler[dto.CatalogOptionResponse, dto.CreateCatalogOptionRequest, dto.UpdateCatalogOptionRequest](
			NewCatalogOptionStore(d.Queries)), lists.CatalogOptions)
	// Django's vehicle make/model viewsets required membership only.
	crud.NewHandler[dto.VehicleMakeResponse, dto.CreateVehicleMakeRequest, dto.UpdateVehicleMakeRequest](
		NewVehicleMakeStore(d.Queries)).Register(member, "/vehicle-makes")
	crud.NewHandler[dto.VehicleModelResponse, dto.CreateVehicleModelRequest, dto.UpdateVehicleModelRequest](
		NewVehicleModelStore(d.Queries)).Register(member, "/vehicle-models")

	member.POST("/uploads", NewUploadHandler(d.Storage, d.Logger).Upload)

	// The caller's own profile and permissions. Gated on identity alone: a
	// client must be able to discover what it may do — including that it may do
	// nothing — without first being refused by a module gate.
	api.GET("/me/permissions", NewMeHandler(d.Queries).Permissions)

	// Notifications are per-caller, so membership is the only gate that applies.
	notifications := NewNotificationHandler(d.Queries)
	member.GET("/notifications", notifications.List)
	member.POST("/notifications/:id/read", notifications.MarkRead)

	// The dashboard reads across every module, so it is gated on membership
	// rather than on any one of them; the counts it returns are already scoped
	// to the caller's company.
	member.GET("/dashboard/stats", NewDashboardHandler(d.Queries).Stats)

	// Groups are the company's org tree. Django's employee viewset owned them,
	// so they sit behind the same module.
	crud.NewHandler[dto.GroupResponse, dto.CreateGroupRequest, dto.UpdateGroupRequest](
		NewGroupStore(d.Queries)).Register(employees, "/groups")

	// Phase 8: nested, parent-scoped child resources
	crud.NewNestedHandler[dto.WorkOrderLineItemResponse, dto.CreateWorkOrderLineItemRequest, dto.UpdateWorkOrderLineItemRequest](NewWorkOrderLineItemStore(d.Queries, d.Pool)).Register(workOrders, "/work-orders", "/line-items")
	crud.NewReadOnlyNestedHandler[dto.WorkOrderStatusLogResponse](
		NewWorkOrderStatusLogStore(d.Queries),
		"the status history is append-only: change the work order's status_id with PUT /api/v1/work-orders/{id} and the transition is recorded automatically",
	).Register(workOrders, "/work-orders", "/status-logs")
	crud.NewNestedHandler[dto.PurchaseOrderLineItemResponse, dto.CreatePurchaseOrderLineItemRequest, dto.UpdatePurchaseOrderLineItemRequest](NewPurchaseOrderLineItemStore(d.Queries, d.Pool)).Register(purchaseOrders, "/purchase-orders", "/line-items")
	crud.NewNestedHandler[dto.ServiceTaskPartResponse, dto.CreateServiceTaskPartRequest, dto.UpdateServiceTaskPartRequest](NewServiceTaskPartStore(d.Queries)).Register(service, "/service-tasks", "/parts")
	crud.NewNestedHandler[dto.ServiceEntryLineItemResponse, dto.CreateServiceEntryLineItemRequest, dto.UpdateServiceEntryLineItemRequest](NewServiceEntryLineItemStore(d.Queries, d.Pool)).Register(service, "/service-entries", "/line-items")
	crud.NewNestedHandler[dto.InspectionFormItemResponse, dto.CreateInspectionFormItemRequest, dto.UpdateInspectionFormItemRequest](NewInspectionFormItemStore(d.Queries)).Register(inspections, "/inspection-forms", "/items")
	crud.NewNestedHandler[dto.InspectionSubmissionItemResponse, dto.CreateInspectionSubmissionItemRequest, dto.UpdateInspectionSubmissionItemRequest](NewInspectionSubmissionItemStore(d.Queries, d.Storage, d.Logger)).Register(inspections, "/inspection-submissions", "/items")
	crud.NewNestedHandler[dto.AxleDefinitionResponse, dto.CreateAxleDefinitionRequest, dto.UpdateAxleDefinitionRequest](NewAxleDefinitionStore(d.Queries)).Register(tires, "/axle-templates", "/definitions")
	crud.NewNestedHandler[dto.AssetTrailerAssignmentResponse, dto.CreateAssetTrailerAssignmentRequest, dto.UpdateAssetTrailerAssignmentRequest](NewAssetTrailerAssignmentStore(d.Queries)).Register(assets, "/assets", "/trailer-assignments")
	crud.NewNestedHandler[dto.PartInventoryResponse, dto.CreatePartInventoryRequest, dto.UpdatePartInventoryRequest](NewPartInventoryStore(d.Queries)).Register(inventory, "/parts", "/inventory")
	crud.NewNestedHandler[dto.FuelEntryResponse, dto.CreateFuelEntryRequest, dto.UpdateFuelEntryRequest](NewFuelEntryStore(d.Queries, d.Pool)).Register(fuel, "/assets", "/fuel-entries")
	// Cross-fleet register. Sits alongside /fuel-entries/{id}/comments and
	// /photos, which are keyed by entry rather than by asset.
	fuel.GET("/fuel-entries", lists.FuelEntries)
	crud.NewNestedHandler[dto.TireInstallationResponse, dto.CreateTireInstallationRequest, dto.UpdateTireInstallationRequest](NewTireInstallationStore(d.Queries)).Register(tires, "/tires", "/installations")
	crud.NewNestedHandler[dto.TireInspectionResponse, dto.CreateTireInspectionRequest, dto.UpdateTireInspectionRequest](NewTireInspectionStore(d.Queries)).Register(tires, "/tires", "/inspections")
	crud.NewNestedHandler[dto.TireMountLogResponse, dto.CreateTireMountLogRequest, dto.UpdateTireMountLogRequest](NewTireMountLogStore(d.Queries)).Register(tires, "/tires", "/mount-logs")
	// Fleet-wide movement history, as opposed to one tire's log.
	tires.GET("/tire-mount-logs", lists.TireMountLogs)

	// Phase 8b: deeper (grandchild) nested resources, scoped up the chain to company
	crud.NewNestedHandler[dto.WorkOrderSubLineItemResponse, dto.CreateWorkOrderSubLineItemRequest, dto.UpdateWorkOrderSubLineItemRequest](NewWorkOrderSubLineItemStore(d.Queries, d.Pool)).Register(workOrders, "/work-order-line-items", "/sub-line-items")
	crud.NewNestedHandler[dto.LaborTimeEntryResponse, dto.CreateLaborTimeEntryRequest, dto.UpdateLaborTimeEntryRequest](NewLaborTimeEntryStore(d.Queries)).Register(workOrders, "/work-order-sub-line-items", "/labor-entries")
	crud.NewNestedHandler[dto.WheelPositionDefinitionResponse, dto.CreateWheelPositionDefinitionRequest, dto.UpdateWheelPositionDefinitionRequest](NewWheelPositionDefinitionStore(d.Queries)).Register(tires, "/axle-definitions", "/wheel-positions")
	crud.NewNestedHandler[dto.FuelCommentResponse, dto.CreateFuelCommentRequest, dto.UpdateFuelCommentRequest](NewFuelCommentStore(d.Queries)).Register(fuel, "/fuel-entries", "/comments")
	fuelPhotos := NewFuelPhotoStore(d.Queries, d.Pool, d.Storage, d.Logger)
	crud.NewNestedHandler[dto.FuelPhotoResponse, dto.CreateFuelPhotoRequest, dto.UpdateFuelPhotoRequest](fuelPhotos).Register(fuel, "/fuel-entries", "/photos")
	// is_primary is not a field on those writes: at most one photo per entry may
	// hold it, so promoting one has to demote the others in the same transaction.
	fuel.POST("/fuel-entries/:id/photos/:child_id/set-primary", func(c *gin.Context) {
		parentID, id, err := crud.ParentAndIDParams(c)
		if err != nil {
			apierr.Abort(c, err)
			return
		}
		out, err := fuelPhotos.SetPrimary(c.Request.Context(), parentID, id)
		if err != nil {
			apierr.Abort(c, err)
			return
		}
		c.JSON(http.StatusOK, out)
	})

	// Phase 8c: shared-PK 1:1 sub-types (singleton under the asset). Django put
	// vehicle/trailer under 'assets' but the axle config under 'tires'.
	crud.NewSingletonHandler[dto.VehicleResponse, dto.UpsertVehicleRequest](NewVehicleStore(d.Queries)).Register(assets, "/assets", "/vehicle")
	crud.NewSingletonHandler[dto.TrailerResponse, dto.UpsertTrailerRequest](NewTrailerStore(d.Queries)).Register(assets, "/assets", "/trailer")
	crud.NewSingletonHandler[dto.VehicleAxleConfigResponse, dto.UpsertVehicleAxleConfigRequest](NewVehicleAxleConfigStore(d.Queries)).Register(tires, "/assets", "/axle-config")

	// Phase 8d: m2m join tables, exposed under the owning side. Each returns the
	// linked entities rather than the join rows, so a client that wants an
	// issue's assignees gets employees back in one request.
	crud.NewLinkHandler[dto.EmployeeResponse, dto.LinkEmployeeRequest](NewIssueAssigneeStore(d.Queries)).Register(issues, "/issues", "/assigned-to", "employee_id")
	crud.NewLinkHandler[dto.EmployeeResponse, dto.LinkEmployeeRequest](NewIssueWatcherStore(d.Queries)).Register(issues, "/issues", "/watchers", "employee_id")
	crud.NewLinkHandler[dto.IssueResponse, dto.LinkIssueRequest](NewWorkOrderIssueStore(d.Queries)).Register(workOrders, "/work-orders", "/issues", "issue_id")
	crud.NewLinkHandler[dto.FaultResponse, dto.LinkFaultRequest](NewWorkOrderFaultStore(d.Queries)).Register(workOrders, "/work-orders", "/faults", "fault_id")
	crud.NewLinkHandler[dto.IssueResponse, dto.LinkIssueRequest](NewServiceEntryLineItemIssueStore(d.Queries)).Register(service, "/service-entry-line-items", "/issues", "issue_id")
	crud.NewLinkHandler[dto.IssueResponse, dto.LinkIssueRequest](NewWorkOrderLineItemIssueStore(d.Queries)).Register(workOrders, "/work-order-line-items", "/issues", "issue_id")

	return r
}

// registerCrudWithList wires the standard CRUD shape but serves the collection
// from a custom handler, for resources whose list takes filters the generic one
// cannot express. crud's own doc comment points here: "for read-only or partial
// resources, wire the exported handlers individually".
func registerCrudWithList[T, C, U any](r gin.IRouter, path string, h *crud.Handler[T, C, U], list gin.HandlerFunc) {
	r.GET(path, list)
	r.POST(path, h.Create)
	r.GET(path+"/:id", h.Get)
	r.PUT(path+"/:id", h.Update)
	r.DELETE(path+"/:id", h.Delete)
}

// registerPurchaseOrderRoutes wires the order plus its workflow.
//
// The transitions are separate routes rather than fields on the PUT because
// each one has to decide whether it is legal from the current state, stamp its
// own timestamp, and record who did it. A whole-record replace can do none of
// that: it takes whatever the client sends.
//
// Approve and reject additionally require the purchase_orders.approve action.
// Deciding to commit money is the privileged act; submitting an order for that
// decision, and receiving the goods afterwards, are ordinary work and need only
// update. The pattern is the one tire_approvals already uses.
func registerPurchaseOrderRoutes(r *gin.RouterGroup, d Deps, list gin.HandlerFunc) {
	const path = "/purchase-orders"

	// registerCrudWithList, not Register: the list comes from the filtered
	// handler, which is the one that understands vendor and state.
	registerCrudWithList(r, path,
		crud.NewHandler[dto.PurchaseOrderResponse, dto.CreatePurchaseOrderRequest, dto.UpdatePurchaseOrderRequest](
			NewPurchaseOrderStore(d.Queries, d.Pool)), list)

	r.POST(path+"/:id/override-total", NewTotalOverrideHandler(d.Queries, d.Pool).PurchaseOrder)

	actions := NewPurchaseOrderActionHandler(d.Queries, d.Pool)
	approve := r.Group("", middleware.RequireAction("purchase_orders", "approve"))

	for _, action := range purchaseorder.Actions() {
		t, _ := purchaseorder.Lookup(action)
		group := r
		if t.RequiresApproval {
			group = approve
		}
		group.POST(path+"/:id/"+action, actions.Handle(action))
	}

	crud.NewReadOnlyNestedHandler[dto.PurchaseOrderStatusLogResponse](
		NewPurchaseOrderStatusLogStore(d.Queries),
		"the transition history is append-only: move the order with POST /api/v1/purchase-orders/{id}/{action} and the transition is recorded automatically",
	).Register(r, path, "/status-logs")
}

// registerInventoryJournalRoutes wires the ledger. It is not registerCrudWithList
// because the ledger is append-only: PUT and DELETE are retired, answering 405
// with the replacement rather than 404, and a reversal route takes their place.
func registerInventoryJournalRoutes(r gin.IRouter, d Deps, list gin.HandlerFunc) {
	const path = "/inventory-journal-entries"

	store := NewInventoryJournalEntryStore(d.Queries, d.Pool)
	entries := crud.NewAppendOnlyHandler[dto.InventoryJournalEntryResponse, dto.CreateInventoryJournalEntryRequest](
		store, journalRetiredMessage)

	// The verbs are wired individually rather than through Register because the
	// list comes from the filtered handler: the generic one has no part_id or
	// date filter, and this is the one collection that never plateaus.
	r.GET(path, list)
	r.POST(path, entries.Create)
	r.GET(path+"/:id", entries.Get)
	r.PUT(path+"/:id", entries.Retired)
	r.DELETE(path+"/:id", entries.Retired)
	r.POST(path+"/:id/reverse", NewInventoryJournalEntryHandler(store).Reverse)
}

// registerCompanyRoutes wires /companies with Django's split gating: creating a
// company requires account ownership (and no membership, since the first company
// is what creates it), while reading and mutating one requires company admin.
func registerCompanyRoutes(api, member *gin.RouterGroup, d Deps) {
	h := crud.NewHandler[dto.CompanyResponse, dto.CreateCompanyRequest, dto.UpdateCompanyRequest](
		NewCompanyStore(d.Queries, d.Pool, d.Storage, d.Logger))

	api.POST("/companies", middleware.RequireAccountOwner(), h.Create)

	admin := member.Group("", middleware.RequireAdminRole())
	admin.GET("/companies", h.List)
	admin.GET("/companies/:id", h.Get)
	admin.PUT("/companies/:id", h.Update)
	admin.DELETE("/companies/:id", h.Delete)
}

// registerAdminRoutes wires the cross-company namespace. Every route here reads
// or writes another tenant's data on purpose — the company-scoped routes cannot
// answer "who belongs to company X" or "did company X get seeded".
//
// It is gated on RequirePlatformAdmin, not RequireAdminRole. A tenant's own
// administrator is not a platform operator: an admin role is granted per
// company membership, so RequireAdminRole only proves the caller administers
// the one company their token is scoped to, and that grants no standing over
// another tenant's employees. Granting the platform-admin flag is a CLI act
// (cmd/cli platform-admin); nothing in the API hands it out.
//
// It is registered on api, not member: platform staff belong to no account and
// hold no company memberships, so gating this namespace on RequireCompanyMember
// would lock every genuine platform admin out of it. RequirePlatformAdmin is
// gate enough — company membership is irrelevant to a cross-tenant route.
func registerAdminRoutes(api *gin.RouterGroup, d Deps) {
	admin := api.Group("", middleware.RequirePlatformAdmin())

	employees := NewAdminEmployeeHandler(d.Queries, d.Pool)
	admin.GET("/admin/employees", employees.List)
	admin.GET("/admin/employees/:id/companies", employees.ListCompanies)
	admin.PUT("/admin/employees/:id/companies", employees.ReplaceCompanies)

	companies := NewAdminCompanyHandler(d.Queries, d.Pool)
	admin.GET("/admin/companies/:id/owner", companies.Owner)
	admin.POST("/admin/companies/:id/set-owner", companies.SetOwner)
	admin.GET("/admin/companies/:id/roles", companies.Roles)
	admin.GET("/admin/companies/:id/work-order-statuses", companies.WorkOrderStatuses)

	accounts := NewAdminAccountHandler(d.Queries, d.Pool)
	admin.POST("/admin/accounts", accounts.Create)
	admin.GET("/admin/accounts", accounts.List)
	admin.POST("/admin/accounts/:id/set-owner", accounts.SetOwner)
}

// registerAccountEmployeeRoutes wires the routes that let an account owner
// appoint directors: list the account's people with the role they hold in
// each company, and replace one person's memberships-with-roles.
//
// Account-scoped, deliberately NOT under /api/v1/admin: that namespace is
// cross-tenant and platform-admin only. This one operates strictly inside the
// caller's own account.
//
// Registered on api, not member: these are gated on RequireAccountOwner, and
// an owner with no company of their own yet is not a company member —
// registering on member would lock out exactly the person the routes exist to
// serve. registerCompanyRoutes' POST /companies is the same precedent.
func registerAccountEmployeeRoutes(api *gin.RouterGroup, d Deps) {
	accountOwner := api.Group("", middleware.RequireAccountOwner())
	accountEmployees := NewAccountEmployeeHandler(d.Queries, d.Pool)
	accountOwner.GET("/account/employees", accountEmployees.List)
	accountOwner.PUT("/account/employees/:id/companies", accountEmployees.ReplaceCompanies)
}
