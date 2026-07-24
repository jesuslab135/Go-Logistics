package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
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

	return r
}
