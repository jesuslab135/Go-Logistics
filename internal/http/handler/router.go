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
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
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

	// Interactive API docs. Disabled in production to avoid exposing the schema.
	if !d.Production {
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

	registerCompanyRoutes(api, member, d)

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
	// Django gated roles on IsAdminRole rather than on a module entry.
	roles := member.Group("", middleware.RequireAdminRole())

	// Collections whose filters vary per request take their List verb from
	// here instead of from the generic CRUD handler.
	lists := NewFilteredListHandler(d.Pool)

	// Phase 2: assets
	registerCrudWithList(assets, "/assets",
		crud.NewHandler[dto.AssetResponse, dto.CreateAssetRequest, dto.UpdateAssetRequest](
			NewAssetStore(d.Queries)), lists.Assets)
	// Read-only reverse view of the nested /assets/{id}/trailer-assignments;
	// mutations stay on the nested route that owns the parent scope.
	assets.GET("/asset-trailer-assignments", lists.AssetTrailerAssignments)

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
	registerCrudWithList(inventory, "/inventory-journal-entries",
		crud.NewHandler[dto.InventoryJournalEntryResponse, dto.CreateInventoryJournalEntryRequest, dto.UpdateInventoryJournalEntryRequest](
			NewInventoryJournalEntryStore(d.Queries)), lists.InventoryJournalEntries)

	// Phase 4: vendors, work orders & issues
	crud.NewHandler[dto.VendorResponse, dto.CreateVendorRequest, dto.UpdateVendorRequest](NewVendorStore(d.Queries)).Register(vendors, "/vendors")
	crud.NewHandler[dto.WorkOrderStatusResponse, dto.CreateWorkOrderStatusRequest, dto.UpdateWorkOrderStatusRequest](NewWorkOrderStatusStore(d.Queries)).Register(workOrders, "/work-order-statuses")
	// Django's LocationViewSet required membership only, no module entry.
	crud.NewHandler[dto.LocationResponse, dto.CreateLocationRequest, dto.UpdateLocationRequest](NewLocationStore(d.Queries)).Register(member, "/locations")
	registerCrudWithList(workOrders, "/work-orders",
		crud.NewHandler[dto.WorkOrderResponse, dto.CreateWorkOrderRequest, dto.UpdateWorkOrderRequest](NewWorkOrderStore(d.Queries)), lists.WorkOrders)
	registerCrudWithList(issues, "/issues",
		crud.NewHandler[dto.IssueResponse, dto.CreateIssueRequest, dto.UpdateIssueRequest](NewIssueStore(d.Queries)), lists.Issues)
	crud.NewHandler[dto.IssuePriorityResponse, dto.CreateIssuePriorityRequest, dto.UpdateIssuePriorityRequest](NewIssuePriorityStore(d.Queries)).Register(issues, "/issue-priorities")
	crud.NewHandler[dto.FaultResponse, dto.CreateFaultRequest, dto.UpdateFaultRequest](NewFaultStore(d.Queries)).Register(issues, "/faults")

	// Phase 5: purchase orders & service
	crud.NewHandler[dto.PurchaseOrderResponse, dto.CreatePurchaseOrderRequest, dto.UpdatePurchaseOrderRequest](NewPurchaseOrderStore(d.Queries)).Register(inventory, "/purchase-orders")
	crud.NewHandler[dto.ServiceTaskResponse, dto.CreateServiceTaskRequest, dto.UpdateServiceTaskRequest](NewServiceTaskStore(d.Queries)).Register(service, "/service-tasks")
	crud.NewHandler[dto.ServiceReminderResponse, dto.CreateServiceReminderRequest, dto.UpdateServiceReminderRequest](NewServiceReminderStore(d.Queries)).Register(service, "/service-reminders")
	crud.NewHandler[dto.ServiceEntryResponse, dto.CreateServiceEntryRequest, dto.UpdateServiceEntryRequest](NewServiceEntryStore(d.Queries)).Register(service, "/service-entries")

	// Phase 6: tires
	crud.NewHandler[dto.TireResponse, dto.CreateTireRequest, dto.UpdateTireRequest](NewTireStore(d.Queries)).Register(tires, "/tires")
	crud.NewHandler[dto.TireModelResponse, dto.CreateTireModelRequest, dto.UpdateTireModelRequest](NewTireModelStore(d.Queries)).Register(tires, "/tire-models")
	crud.NewHandler[dto.AxleTemplateResponse, dto.CreateAxleTemplateRequest, dto.UpdateAxleTemplateRequest](NewAxleTemplateStore(d.Queries)).Register(tires, "/axle-templates")
	// Django's TireAssignmentRequestViewSet required membership only; its
	// approve/reject actions carried the extra warehouse-role check.
	crud.NewHandler[dto.TireAssignmentRequestResponse, dto.CreateTireAssignmentRequestRequest, dto.UpdateTireAssignmentRequestRequest](NewTireAssignmentRequestStore(d.Queries)).Register(member, "/tire-assignment-requests")
	// Approving is the warehouse's act, not an ordinary write: Django gated it
	// on tire_approvals/approve, which admins bypass.
	member.POST("/tire-assignment-requests/:id/approve",
		middleware.RequireAction("tire_approvals", "approve"),
		NewTireAssignmentActionHandler(d.Queries, d.Pool).Approve)

	// Phase 7: fuel, inspections, org & misc
	crud.NewHandler[dto.FuelTypeResponse, dto.CreateFuelTypeRequest, dto.UpdateFuelTypeRequest](NewFuelTypeStore(d.Queries)).Register(fuel, "/fuel-types")
	crud.NewHandler[dto.InspectionFormResponse, dto.CreateInspectionFormRequest, dto.UpdateInspectionFormRequest](NewInspectionFormStore(d.Queries)).Register(inspections, "/inspection-forms")
	crud.NewHandler[dto.InspectionSubmissionResponse, dto.CreateInspectionSubmissionRequest, dto.UpdateInspectionSubmissionRequest](NewInspectionSubmissionStore(d.Queries)).Register(inspections, "/inspection-submissions")
	crud.NewHandler[dto.MediumResponse, dto.CreateMediumRequest, dto.UpdateMediumRequest](NewMediumStore(d.Queries, d.Storage)).Register(assets, "/media")
	crud.NewHandler[dto.CommentResponse, dto.CreateCommentRequest, dto.UpdateCommentRequest](NewCommentStore(d.Queries)).Register(issues, "/comments")
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
	crud.NewHandler[dto.CatalogOptionResponse, dto.CreateCatalogOptionRequest, dto.UpdateCatalogOptionRequest](
		NewCatalogOptionStore(d.Queries)).Register(assets, "/catalog-options")
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
	member.GET("/notifications", NewNotificationHandler().List)

	// The dashboard reads across every module, so it is gated on membership
	// rather than on any one of them; the counts it returns are already scoped
	// to the caller's company.
	member.GET("/dashboard/stats", NewDashboardHandler(d.Queries).Stats)

	// Groups are the company's org tree. Django's employee viewset owned them,
	// so they sit behind the same module.
	crud.NewHandler[dto.GroupResponse, dto.CreateGroupRequest, dto.UpdateGroupRequest](
		NewGroupStore(d.Queries)).Register(employees, "/groups")

	// Phase 8: nested, parent-scoped child resources
	crud.NewNestedHandler[dto.WorkOrderLineItemResponse, dto.CreateWorkOrderLineItemRequest, dto.UpdateWorkOrderLineItemRequest](NewWorkOrderLineItemStore(d.Queries)).Register(workOrders, "/work-orders", "/line-items")
	crud.NewNestedHandler[dto.WorkOrderStatusLogResponse, dto.CreateWorkOrderStatusLogRequest, dto.UpdateWorkOrderStatusLogRequest](NewWorkOrderStatusLogStore(d.Queries)).Register(workOrders, "/work-orders", "/status-logs")
	crud.NewNestedHandler[dto.PurchaseOrderLineItemResponse, dto.CreatePurchaseOrderLineItemRequest, dto.UpdatePurchaseOrderLineItemRequest](NewPurchaseOrderLineItemStore(d.Queries)).Register(inventory, "/purchase-orders", "/line-items")
	crud.NewNestedHandler[dto.ServiceTaskPartResponse, dto.CreateServiceTaskPartRequest, dto.UpdateServiceTaskPartRequest](NewServiceTaskPartStore(d.Queries)).Register(service, "/service-tasks", "/parts")
	crud.NewNestedHandler[dto.ServiceEntryLineItemResponse, dto.CreateServiceEntryLineItemRequest, dto.UpdateServiceEntryLineItemRequest](NewServiceEntryLineItemStore(d.Queries)).Register(service, "/service-entries", "/line-items")
	crud.NewNestedHandler[dto.InspectionFormItemResponse, dto.CreateInspectionFormItemRequest, dto.UpdateInspectionFormItemRequest](NewInspectionFormItemStore(d.Queries)).Register(inspections, "/inspection-forms", "/items")
	crud.NewNestedHandler[dto.InspectionSubmissionItemResponse, dto.CreateInspectionSubmissionItemRequest, dto.UpdateInspectionSubmissionItemRequest](NewInspectionSubmissionItemStore(d.Queries)).Register(inspections, "/inspection-submissions", "/items")
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
	crud.NewNestedHandler[dto.WorkOrderSubLineItemResponse, dto.CreateWorkOrderSubLineItemRequest, dto.UpdateWorkOrderSubLineItemRequest](NewWorkOrderSubLineItemStore(d.Queries)).Register(workOrders, "/work-order-line-items", "/sub-line-items")
	crud.NewNestedHandler[dto.LaborTimeEntryResponse, dto.CreateLaborTimeEntryRequest, dto.UpdateLaborTimeEntryRequest](NewLaborTimeEntryStore(d.Queries)).Register(workOrders, "/work-order-sub-line-items", "/labor-entries")
	crud.NewNestedHandler[dto.WheelPositionDefinitionResponse, dto.CreateWheelPositionDefinitionRequest, dto.UpdateWheelPositionDefinitionRequest](NewWheelPositionDefinitionStore(d.Queries)).Register(tires, "/axle-definitions", "/wheel-positions")
	crud.NewNestedHandler[dto.FuelCommentResponse, dto.CreateFuelCommentRequest, dto.UpdateFuelCommentRequest](NewFuelCommentStore(d.Queries)).Register(fuel, "/fuel-entries", "/comments")
	crud.NewNestedHandler[dto.FuelPhotoResponse, dto.CreateFuelPhotoRequest, dto.UpdateFuelPhotoRequest](NewFuelPhotoStore(d.Queries)).Register(fuel, "/fuel-entries", "/photos")

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

// registerCompanyRoutes wires /companies with Django's split gating: creating a
// company requires account ownership (and no membership, since the first company
// is what creates it), while reading and mutating one requires company admin.
func registerCompanyRoutes(api, member *gin.RouterGroup, d Deps) {
	h := crud.NewHandler[dto.CompanyResponse, dto.CreateCompanyRequest, dto.UpdateCompanyRequest](
		NewCompanyStore(d.Queries, d.Pool))

	api.POST("/companies", middleware.RequireAccountOwner(), h.Create)

	admin := member.Group("", middleware.RequireAdminRole())
	admin.GET("/companies", h.List)
	admin.GET("/companies/:id", h.Get)
	admin.PUT("/companies/:id", h.Update)
	admin.DELETE("/companies/:id", h.Delete)
}
