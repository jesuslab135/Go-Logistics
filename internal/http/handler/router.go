package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

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

	authH := NewAuthHandler(d.Tokens, d.Verifier)
	r.POST("/auth/login", authH.Login)
	r.POST("/auth/refresh", authH.Refresh)

	api := r.Group("/api/v1")
	api.Use(middleware.Auth(d.Tokens))

	companies := crud.NewHandler[dto.CompanyResponse, dto.CreateCompanyRequest, dto.UpdateCompanyRequest](
		NewCompanyStore(d.Queries),
	)
	companies.Register(api, "/companies")

	return r
}
