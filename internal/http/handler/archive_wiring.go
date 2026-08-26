package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
)

// The five archivable resources. Each is the same three functions — archive,
// restore, count references — differing only in which table they touch.

func registerArchiveRoutes(assets, parts, vendors, service, inspections gin.IRouter, q *gen.Queries) {
	newArchiveHandler(archiver[dto.AssetResponse]{
		resource: "asset",
		archive: func(ctx context.Context, id, company int64, at time.Time) (dto.AssetResponse, error) {
			r, err := q.ArchiveAsset(ctx, gen.ArchiveAssetParams{ID: id, CompanyID: company, ArchivedAt: &at})
			return toAssetResponse(r), err
		},
		restore: func(ctx context.Context, id, company int64, at time.Time) (dto.AssetResponse, error) {
			r, err := q.RestoreAsset(ctx, gen.RestoreAssetParams{ID: id, CompanyID: company, UpdatedAt: at})
			return toAssetResponse(r), err
		},
		references: q.AssetReferences,
	}).Register(assets, "/assets")

	newArchiveHandler(archiver[dto.PartResponse]{
		resource: "part",
		archive: func(ctx context.Context, id, company int64, at time.Time) (dto.PartResponse, error) {
			r, err := q.ArchivePart(ctx, gen.ArchivePartParams{ID: id, CompanyID: company, ArchivedAt: &at})
			return toPartResponse(r), err
		},
		restore: func(ctx context.Context, id, company int64, at time.Time) (dto.PartResponse, error) {
			r, err := q.RestorePart(ctx, gen.RestorePartParams{ID: id, CompanyID: company, UpdatedAt: at})
			return toPartResponse(r), err
		},
		references: q.PartReferences,
	}).Register(parts, "/parts")

	newArchiveHandler(archiver[dto.VendorResponse]{
		resource: "vendor",
		archive: func(ctx context.Context, id, company int64, at time.Time) (dto.VendorResponse, error) {
			r, err := q.ArchiveVendor(ctx, gen.ArchiveVendorParams{ID: id, CompanyID: company, ArchivedAt: &at})
			return toVendorResponse(r), err
		},
		restore: func(ctx context.Context, id, company int64, at time.Time) (dto.VendorResponse, error) {
			r, err := q.RestoreVendor(ctx, gen.RestoreVendorParams{ID: id, CompanyID: company, UpdatedAt: at})
			return toVendorResponse(r), err
		},
		references: q.VendorReferences,
	}).Register(vendors, "/vendors")

	newArchiveHandler(archiver[dto.ServiceTaskResponse]{
		resource: "service task",
		archive: func(ctx context.Context, id, company int64, at time.Time) (dto.ServiceTaskResponse, error) {
			r, err := q.ArchiveServiceTask(ctx, gen.ArchiveServiceTaskParams{ID: id, CompanyID: company, ArchivedAt: &at})
			return toServiceTaskResponse(r), err
		},
		restore: func(ctx context.Context, id, company int64, at time.Time) (dto.ServiceTaskResponse, error) {
			r, err := q.RestoreServiceTask(ctx, gen.RestoreServiceTaskParams{ID: id, CompanyID: company, UpdatedAt: at})
			return toServiceTaskResponse(r), err
		},
		references: q.ServiceTaskReferences,
	}).Register(service, "/service-tasks")

	newArchiveHandler(archiver[dto.InspectionFormResponse]{
		resource: "inspection form",
		archive: func(ctx context.Context, id, company int64, at time.Time) (dto.InspectionFormResponse, error) {
			r, err := q.ArchiveInspectionForm(ctx, gen.ArchiveInspectionFormParams{ID: id, CompanyID: company, ArchivedAt: &at})
			return toInspectionFormResponse(r), err
		},
		restore: func(ctx context.Context, id, company int64, at time.Time) (dto.InspectionFormResponse, error) {
			r, err := q.RestoreInspectionForm(ctx, gen.RestoreInspectionFormParams{ID: id, CompanyID: company, UpdatedAt: at})
			return toInspectionFormResponse(r), err
		},
		references: q.InspectionFormReferences,
	}).Register(inspections, "/inspection-forms")
}
