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

	authH := NewAuthHandler(d.Tokens, d.Verifier)
	r.POST("/auth/login", authH.Login)
	r.POST("/auth/refresh", authH.Refresh)

	api := r.Group("/api/v1")
	api.Use(middleware.Auth(d.Tokens))

	crud.NewHandler[dto.CompanyResponse, dto.CreateCompanyRequest, dto.UpdateCompanyRequest](
		NewCompanyStore(d.Queries)).Register(api, "/companies")

	// Phase 2: assets
	crud.NewHandler[dto.AssetResponse, dto.CreateAssetRequest, dto.UpdateAssetRequest](
		NewAssetStore(d.Queries)).Register(api, "/assets")

	// Phase 3: parts & inventory
	crud.NewHandler[dto.PartCategoryResponse, dto.CreatePartCategoryRequest, dto.UpdatePartCategoryRequest](
		NewPartCategoryStore(d.Queries)).Register(api, "/part-categories")
	crud.NewHandler[dto.PartManufacturerResponse, dto.CreatePartManufacturerRequest, dto.UpdatePartManufacturerRequest](
		NewPartManufacturerStore(d.Queries)).Register(api, "/part-manufacturers")
	crud.NewHandler[dto.MeasurementUnitResponse, dto.CreateMeasurementUnitRequest, dto.UpdateMeasurementUnitRequest](
		NewMeasurementUnitStore(d.Queries)).Register(api, "/measurement-units")
	crud.NewHandler[dto.PartResponse, dto.CreatePartRequest, dto.UpdatePartRequest](
		NewPartStore(d.Queries)).Register(api, "/parts")
	crud.NewHandler[dto.PartLocationResponse, dto.CreatePartLocationRequest, dto.UpdatePartLocationRequest](
		NewPartLocationStore(d.Queries)).Register(api, "/part-locations")
	crud.NewHandler[dto.InventoryAdjustmentReasonResponse, dto.CreateInventoryAdjustmentReasonRequest, dto.UpdateInventoryAdjustmentReasonRequest](
		NewInventoryAdjustmentReasonStore(d.Queries)).Register(api, "/inventory-adjustment-reasons")
	crud.NewHandler[dto.InventoryJournalEntryResponse, dto.CreateInventoryJournalEntryRequest, dto.UpdateInventoryJournalEntryRequest](
		NewInventoryJournalEntryStore(d.Queries)).Register(api, "/inventory-journal-entries")

	// Phase 4: vendors, work orders & issues
	crud.NewHandler[dto.VendorResponse, dto.CreateVendorRequest, dto.UpdateVendorRequest](NewVendorStore(d.Queries)).Register(api, "/vendors")
	crud.NewHandler[dto.WorkOrderStatusResponse, dto.CreateWorkOrderStatusRequest, dto.UpdateWorkOrderStatusRequest](NewWorkOrderStatusStore(d.Queries)).Register(api, "/work-order-statuses")
	crud.NewHandler[dto.LocationResponse, dto.CreateLocationRequest, dto.UpdateLocationRequest](NewLocationStore(d.Queries)).Register(api, "/locations")
	crud.NewHandler[dto.WorkOrderResponse, dto.CreateWorkOrderRequest, dto.UpdateWorkOrderRequest](NewWorkOrderStore(d.Queries)).Register(api, "/work-orders")
	crud.NewHandler[dto.IssueResponse, dto.CreateIssueRequest, dto.UpdateIssueRequest](NewIssueStore(d.Queries)).Register(api, "/issues")
	crud.NewHandler[dto.IssuePriorityResponse, dto.CreateIssuePriorityRequest, dto.UpdateIssuePriorityRequest](NewIssuePriorityStore(d.Queries)).Register(api, "/issue-priorities")
	crud.NewHandler[dto.FaultResponse, dto.CreateFaultRequest, dto.UpdateFaultRequest](NewFaultStore(d.Queries)).Register(api, "/faults")

	// Phase 5: purchase orders & service
	crud.NewHandler[dto.PurchaseOrderResponse, dto.CreatePurchaseOrderRequest, dto.UpdatePurchaseOrderRequest](NewPurchaseOrderStore(d.Queries)).Register(api, "/purchase-orders")
	crud.NewHandler[dto.ServiceTaskResponse, dto.CreateServiceTaskRequest, dto.UpdateServiceTaskRequest](NewServiceTaskStore(d.Queries)).Register(api, "/service-tasks")
	crud.NewHandler[dto.ServiceReminderResponse, dto.CreateServiceReminderRequest, dto.UpdateServiceReminderRequest](NewServiceReminderStore(d.Queries)).Register(api, "/service-reminders")
	crud.NewHandler[dto.ServiceEntryResponse, dto.CreateServiceEntryRequest, dto.UpdateServiceEntryRequest](NewServiceEntryStore(d.Queries)).Register(api, "/service-entries")

	// Phase 6: tires
	crud.NewHandler[dto.TireResponse, dto.CreateTireRequest, dto.UpdateTireRequest](NewTireStore(d.Queries)).Register(api, "/tires")
	crud.NewHandler[dto.TireModelResponse, dto.CreateTireModelRequest, dto.UpdateTireModelRequest](NewTireModelStore(d.Queries)).Register(api, "/tire-models")
	crud.NewHandler[dto.AxleTemplateResponse, dto.CreateAxleTemplateRequest, dto.UpdateAxleTemplateRequest](NewAxleTemplateStore(d.Queries)).Register(api, "/axle-templates")
	crud.NewHandler[dto.TireAssignmentRequestResponse, dto.CreateTireAssignmentRequestRequest, dto.UpdateTireAssignmentRequestRequest](NewTireAssignmentRequestStore(d.Queries)).Register(api, "/tire-assignment-requests")

	// Phase 7: fuel, inspections, org & misc
	crud.NewHandler[dto.FuelTypeResponse, dto.CreateFuelTypeRequest, dto.UpdateFuelTypeRequest](NewFuelTypeStore(d.Queries)).Register(api, "/fuel-types")
	crud.NewHandler[dto.InspectionFormResponse, dto.CreateInspectionFormRequest, dto.UpdateInspectionFormRequest](NewInspectionFormStore(d.Queries)).Register(api, "/inspection-forms")
	crud.NewHandler[dto.InspectionSubmissionResponse, dto.CreateInspectionSubmissionRequest, dto.UpdateInspectionSubmissionRequest](NewInspectionSubmissionStore(d.Queries)).Register(api, "/inspection-submissions")
	crud.NewHandler[dto.MediumResponse, dto.CreateMediumRequest, dto.UpdateMediumRequest](NewMediumStore(d.Queries)).Register(api, "/media")
	crud.NewHandler[dto.CommentResponse, dto.CreateCommentRequest, dto.UpdateCommentRequest](NewCommentStore(d.Queries)).Register(api, "/comments")
	crud.NewHandler[dto.WarrantyResponse, dto.CreateWarrantyRequest, dto.UpdateWarrantyRequest](NewWarrantyStore(d.Queries)).Register(api, "/warranties")
	crud.NewHandler[dto.WeeklyMileageGoalResponse, dto.CreateWeeklyMileageGoalRequest, dto.UpdateWeeklyMileageGoalRequest](NewWeeklyMileageGoalStore(d.Queries)).Register(api, "/weekly-mileage-goals")
	crud.NewHandler[dto.RoleResponse, dto.CreateRoleRequest, dto.UpdateRoleRequest](NewRoleStore(d.Queries)).Register(api, "/roles")
	// employee: m2m-scoped, password_hash excluded, transactional create
	crud.NewHandler[dto.EmployeeResponse, dto.CreateEmployeeRequest, dto.UpdateEmployeeRequest](NewEmployeeStore(d.Queries, d.Pool)).Register(api, "/employees")

	// Phase 1: asset catalogs & vehicle models (tenant-scoped)
	crud.NewHandler[dto.AssetTypeResponse, dto.CreateAssetTypeRequest, dto.UpdateAssetTypeRequest](
		NewAssetTypeStore(d.Queries)).Register(api, "/asset-types")
	crud.NewHandler[dto.AssetStatusResponse, dto.CreateAssetStatusRequest, dto.UpdateAssetStatusRequest](
		NewAssetStatusStore(d.Queries)).Register(api, "/asset-statuses")
	crud.NewHandler[dto.CatalogOptionResponse, dto.CreateCatalogOptionRequest, dto.UpdateCatalogOptionRequest](
		NewCatalogOptionStore(d.Queries)).Register(api, "/catalog-options")
	crud.NewHandler[dto.VehicleMakeResponse, dto.CreateVehicleMakeRequest, dto.UpdateVehicleMakeRequest](
		NewVehicleMakeStore(d.Queries)).Register(api, "/vehicle-makes")
	crud.NewHandler[dto.VehicleModelResponse, dto.CreateVehicleModelRequest, dto.UpdateVehicleModelRequest](
		NewVehicleModelStore(d.Queries)).Register(api, "/vehicle-models")

	api.POST("/uploads", NewUploadHandler(d.Storage).Upload)

	// Phase 8: nested, parent-scoped child resources
	crud.NewNestedHandler[dto.WorkOrderLineItemResponse, dto.CreateWorkOrderLineItemRequest, dto.UpdateWorkOrderLineItemRequest](NewWorkOrderLineItemStore(d.Queries)).Register(api, "/work-orders", "/line-items")
	crud.NewNestedHandler[dto.WorkOrderStatusLogResponse, dto.CreateWorkOrderStatusLogRequest, dto.UpdateWorkOrderStatusLogRequest](NewWorkOrderStatusLogStore(d.Queries)).Register(api, "/work-orders", "/status-logs")
	crud.NewNestedHandler[dto.PurchaseOrderLineItemResponse, dto.CreatePurchaseOrderLineItemRequest, dto.UpdatePurchaseOrderLineItemRequest](NewPurchaseOrderLineItemStore(d.Queries)).Register(api, "/purchase-orders", "/line-items")
	crud.NewNestedHandler[dto.ServiceTaskPartResponse, dto.CreateServiceTaskPartRequest, dto.UpdateServiceTaskPartRequest](NewServiceTaskPartStore(d.Queries)).Register(api, "/service-tasks", "/parts")
	crud.NewNestedHandler[dto.ServiceEntryLineItemResponse, dto.CreateServiceEntryLineItemRequest, dto.UpdateServiceEntryLineItemRequest](NewServiceEntryLineItemStore(d.Queries)).Register(api, "/service-entries", "/line-items")
	crud.NewNestedHandler[dto.InspectionFormItemResponse, dto.CreateInspectionFormItemRequest, dto.UpdateInspectionFormItemRequest](NewInspectionFormItemStore(d.Queries)).Register(api, "/inspection-forms", "/items")
	crud.NewNestedHandler[dto.InspectionSubmissionItemResponse, dto.CreateInspectionSubmissionItemRequest, dto.UpdateInspectionSubmissionItemRequest](NewInspectionSubmissionItemStore(d.Queries)).Register(api, "/inspection-submissions", "/items")
	crud.NewNestedHandler[dto.AxleDefinitionResponse, dto.CreateAxleDefinitionRequest, dto.UpdateAxleDefinitionRequest](NewAxleDefinitionStore(d.Queries)).Register(api, "/axle-templates", "/definitions")
	crud.NewNestedHandler[dto.AssetTrailerAssignmentResponse, dto.CreateAssetTrailerAssignmentRequest, dto.UpdateAssetTrailerAssignmentRequest](NewAssetTrailerAssignmentStore(d.Queries)).Register(api, "/assets", "/trailer-assignments")
	crud.NewNestedHandler[dto.PartInventoryResponse, dto.CreatePartInventoryRequest, dto.UpdatePartInventoryRequest](NewPartInventoryStore(d.Queries)).Register(api, "/parts", "/inventory")
	crud.NewNestedHandler[dto.FuelEntryResponse, dto.CreateFuelEntryRequest, dto.UpdateFuelEntryRequest](NewFuelEntryStore(d.Queries)).Register(api, "/assets", "/fuel-entries")
	crud.NewNestedHandler[dto.TireInstallationResponse, dto.CreateTireInstallationRequest, dto.UpdateTireInstallationRequest](NewTireInstallationStore(d.Queries)).Register(api, "/tires", "/installations")
	crud.NewNestedHandler[dto.TireInspectionResponse, dto.CreateTireInspectionRequest, dto.UpdateTireInspectionRequest](NewTireInspectionStore(d.Queries)).Register(api, "/tires", "/inspections")
	crud.NewNestedHandler[dto.TireMountLogResponse, dto.CreateTireMountLogRequest, dto.UpdateTireMountLogRequest](NewTireMountLogStore(d.Queries)).Register(api, "/tires", "/mount-logs")

	// Phase 8b: deeper (grandchild) nested resources, scoped up the chain to company
	crud.NewNestedHandler[dto.WorkOrderSubLineItemResponse, dto.CreateWorkOrderSubLineItemRequest, dto.UpdateWorkOrderSubLineItemRequest](NewWorkOrderSubLineItemStore(d.Queries)).Register(api, "/work-order-line-items", "/sub-line-items")
	crud.NewNestedHandler[dto.LaborTimeEntryResponse, dto.CreateLaborTimeEntryRequest, dto.UpdateLaborTimeEntryRequest](NewLaborTimeEntryStore(d.Queries)).Register(api, "/work-order-sub-line-items", "/labor-entries")
	crud.NewNestedHandler[dto.WheelPositionDefinitionResponse, dto.CreateWheelPositionDefinitionRequest, dto.UpdateWheelPositionDefinitionRequest](NewWheelPositionDefinitionStore(d.Queries)).Register(api, "/axle-definitions", "/wheel-positions")
	crud.NewNestedHandler[dto.FuelCommentResponse, dto.CreateFuelCommentRequest, dto.UpdateFuelCommentRequest](NewFuelCommentStore(d.Queries)).Register(api, "/fuel-entries", "/comments")
	crud.NewNestedHandler[dto.FuelPhotoResponse, dto.CreateFuelPhotoRequest, dto.UpdateFuelPhotoRequest](NewFuelPhotoStore(d.Queries)).Register(api, "/fuel-entries", "/photos")

	// Phase 8c: shared-PK 1:1 sub-types (singleton under the asset)
	crud.NewSingletonHandler[dto.VehicleResponse, dto.UpsertVehicleRequest](NewVehicleStore(d.Queries)).Register(api, "/assets", "/vehicle")
	crud.NewSingletonHandler[dto.TrailerResponse, dto.UpsertTrailerRequest](NewTrailerStore(d.Queries)).Register(api, "/assets", "/trailer")
	crud.NewSingletonHandler[dto.VehicleAxleConfigResponse, dto.UpsertVehicleAxleConfigRequest](NewVehicleAxleConfigStore(d.Queries)).Register(api, "/assets", "/axle-config")

	return r
}
