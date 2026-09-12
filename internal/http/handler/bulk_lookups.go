package handler

import (
	"context"
	"strings"

	"fleet/internal/domain/customfield"
	"fleet/internal/http/dto"
	"fleet/internal/platform/bulk"
	"fleet/internal/platform/paginate"
)

// bulkLookups resolves reference columns. Each lookup pages through the
// referenced section's own store, which is scoped to the caller's company, so
// a name or id can only ever resolve to the caller's own record.
type bulkLookups struct{ d Deps }

// named builds a lookup with both loads over the same list: the unbounded one
// the import's reference index needs, and the bounded one the template's
// catalogue uses so a company with 50,000 assets does not page all of them out
// of Postgres on every template GET.
func named[T any](label string, list func(context.Context, paginate.Params) ([]T, int64, error), id func(T) int64, name func(T) string) bulk.Lookup {
	return bulk.Lookup{
		Label:      label,
		Load:       bulk.Entries(list, id, name),
		LoadCapped: bulk.EntriesUpTo(list, id, name),
	}
}

func (l bulkLookups) assetTypes() bulk.Lookup {
	return named("tipo de activo", NewAssetTypeStore(l.d.Queries).List,
		func(r dto.AssetTypeResponse) int64 { return r.ID }, func(r dto.AssetTypeResponse) string { return r.Name })
}

func (l bulkLookups) assetStatuses() bulk.Lookup {
	return named("estado de activo", NewAssetStatusStore(l.d.Queries).List,
		func(r dto.AssetStatusResponse) int64 { return r.ID }, func(r dto.AssetStatusResponse) string { return r.Name })
}

func (l bulkLookups) assets() bulk.Lookup {
	return named("activo", NewAssetStore(l.d.Queries, l.d.Pool, l.d.Storage, l.d.Logger).List,
		func(r dto.AssetResponse) int64 { return r.ID }, func(r dto.AssetResponse) string { return r.Name })
}

func (l bulkLookups) vendors() bulk.Lookup {
	return named("proveedor", NewVendorStore(l.d.Queries).List,
		func(r dto.VendorResponse) int64 { return r.ID }, func(r dto.VendorResponse) string { return r.Name })
}

func (l bulkLookups) locations() bulk.Lookup {
	return named("ubicación", NewLocationStore(l.d.Queries).List,
		func(r dto.LocationResponse) int64 { return r.ID }, func(r dto.LocationResponse) string { return r.Name })
}

func (l bulkLookups) partCategories() bulk.Lookup {
	return named("categoría de refacción", NewPartCategoryStore(l.d.Queries).List,
		func(r dto.PartCategoryResponse) int64 { return r.ID }, func(r dto.PartCategoryResponse) string { return r.Name })
}

func (l bulkLookups) partManufacturers() bulk.Lookup {
	return named("fabricante", NewPartManufacturerStore(l.d.Queries).List,
		func(r dto.PartManufacturerResponse) int64 { return r.ID }, func(r dto.PartManufacturerResponse) string { return r.Name })
}

func (l bulkLookups) measurementUnits() bulk.Lookup {
	return named("unidad de medida", NewMeasurementUnitStore(l.d.Queries).List,
		func(r dto.MeasurementUnitResponse) int64 { return r.ID }, func(r dto.MeasurementUnitResponse) string { return r.Name })
}

func (l bulkLookups) parts() bulk.Lookup {
	return named("refacción (número de parte)", NewPartStore(l.d.Queries).List,
		func(r dto.PartResponse) int64 { return r.ID }, func(r dto.PartResponse) string { return r.PartNumber })
}

func (l bulkLookups) partLocations() bulk.Lookup {
	return named("almacén", NewPartLocationStore(l.d.Queries).List,
		func(r dto.PartLocationResponse) int64 { return r.ID }, func(r dto.PartLocationResponse) string { return r.Name })
}

func (l bulkLookups) workOrderStatuses() bulk.Lookup {
	return named("estado de orden", NewWorkOrderStatusStore(l.d.Queries).List,
		func(r dto.WorkOrderStatusResponse) int64 { return r.ID }, func(r dto.WorkOrderStatusResponse) string { return r.Name })
}

func (l bulkLookups) workOrders() bulk.Lookup {
	return named("orden de trabajo (número)", NewWorkOrderStore(l.d.Queries, l.d.Pool).List,
		func(r dto.WorkOrderResponse) int64 { return r.ID }, func(r dto.WorkOrderResponse) string { return r.Number })
}

func (l bulkLookups) employees() bulk.Lookup {
	return named("empleado (correo)", NewEmployeeStore(l.d.Queries, l.d.Pool).List,
		func(r dto.EmployeeResponse) int64 { return r.ID }, func(r dto.EmployeeResponse) string { return r.Email })
}

func (l bulkLookups) faults() bulk.Lookup {
	return named("falla (código)", NewFaultStore(l.d.Queries).List,
		func(r dto.FaultResponse) int64 { return r.ID }, func(r dto.FaultResponse) string { return r.Code })
}

func (l bulkLookups) issuePriorities() bulk.Lookup {
	return named("prioridad", NewIssuePriorityStore(l.d.Queries).List,
		func(r dto.IssuePriorityResponse) int64 { return r.ID }, func(r dto.IssuePriorityResponse) string { return r.Name })
}

func (l bulkLookups) serviceTasks() bulk.Lookup {
	return named("tarea de servicio", NewServiceTaskStore(l.d.Queries).List,
		func(r dto.ServiceTaskResponse) int64 { return r.ID }, func(r dto.ServiceTaskResponse) string { return r.Name })
}

func (l bulkLookups) tireModels() bulk.Lookup {
	return named("modelo de llanta (marca modelo medida)", NewTireModelStore(l.d.Queries).List,
		func(r dto.TireModelResponse) int64 { return r.ID },
		func(r dto.TireModelResponse) string {
			return strings.Join(strings.Fields(r.Brand+" "+r.ModelName+" "+r.Size), " ")
		})
}

func (l bulkLookups) tires() bulk.Lookup {
	return named("llanta (número de identificación)", NewTireStore(l.d.Queries).List,
		func(r dto.TireResponse) int64 { return r.ID }, func(r dto.TireResponse) string { return r.TireIdentificationNumber })
}

func (l bulkLookups) trailerClassifications() bulk.Lookup {
	return named("clasificación de remolque", NewTrailerClassificationStore(l.d.Queries).List,
		func(r dto.TrailerClassificationResponse) int64 { return r.ID }, func(r dto.TrailerClassificationResponse) string { return r.Name })
}

func (l bulkLookups) groups() bulk.Lookup {
	return named("grupo", NewGroupStore(l.d.Queries).List,
		func(r dto.GroupResponse) int64 { return r.ID }, func(r dto.GroupResponse) string { return r.Name })
}

func (l bulkLookups) roles() bulk.Lookup {
	return named("rol", NewRoleStore(l.d.Queries).List,
		func(r dto.RoleResponse) int64 { return r.ID }, func(r dto.RoleResponse) string { return r.Name })
}

func (l bulkLookups) vehicleMakes() bulk.Lookup {
	return named("marca de vehículo", NewVehicleMakeStore(l.d.Queries).List,
		func(r dto.VehicleMakeResponse) int64 { return r.ID }, func(r dto.VehicleMakeResponse) string { return r.Name })
}

// customFieldsFor returns the definitions loader for a section with custom
// fields, or nil without a database (unit tests build the registry on Deps{}).
func customFieldsFor(d Deps, resource string) func(context.Context) ([]customfield.Definition, error) {
	if d.Queries == nil {
		return nil
	}
	return func(ctx context.Context) ([]customfield.Definition, error) {
		return definitionsFor(ctx, d.Queries, resource)
	}
}
