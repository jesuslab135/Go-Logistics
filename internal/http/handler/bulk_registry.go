package handler

import (
	"github.com/gin-gonic/gin"

	"fleet/internal/http/dto"
	"fleet/internal/platform/bulk"
)

// bulkGroups are the route groups NewRouter builds, one per module gate.
type bulkGroups struct {
	member, assets, tires, parts, inventory, workOrders, issues, service,
	fuel, vendors, warranties, mileageGoals, employees gin.IRouter
}

type bulkEntry struct {
	group func(bulkGroups) gin.IRouter
	path  string
	imp   bulk.Importer
}

type refs = map[string]bulk.Lookup

// bulkImporters is every importable section. A section is out of this list on
// purpose when creating it is a workflow (purchase orders, tire assignment
// requests), a file upload (media, fuel photos), configuration (roles, custom
// field definitions) or a line item entered inside its parent.
func bulkImporters(d Deps) []bulkEntry {
	q, l := d.Queries, bulkLookups{d}
	member := func(b bulkGroups) gin.IRouter { return b.member }
	assets := func(b bulkGroups) gin.IRouter { return b.assets }
	tires := func(b bulkGroups) gin.IRouter { return b.tires }
	parts := func(b bulkGroups) gin.IRouter { return b.parts }
	inventory := func(b bulkGroups) gin.IRouter { return b.inventory }
	workOrders := func(b bulkGroups) gin.IRouter { return b.workOrders }
	issues := func(b bulkGroups) gin.IRouter { return b.issues }
	service := func(b bulkGroups) gin.IRouter { return b.service }
	fuel := func(b bulkGroups) gin.IRouter { return b.fuel }
	vendors := func(b bulkGroups) gin.IRouter { return b.vendors }
	warranties := func(b bulkGroups) gin.IRouter { return b.warranties }
	mileageGoals := func(b bulkGroups) gin.IRouter { return b.mileageGoals }
	employees := func(b bulkGroups) gin.IRouter { return b.employees }

	return []bulkEntry{
		{assets, "/assets", bulk.Flat[dto.AssetResponse, dto.CreateAssetRequest, dto.UpdateAssetRequest]{
			Name: "assets", Store: NewAssetStore(q, d.Pool, d.Storage, d.Logger), CustomFields: customFieldsFor(d, "assets"),
			Skip: []string{"linked_vehicles", "photo", "owner_company_id"},
			References: refs{"asset_type_id": l.assetTypes(), "status_id": l.assetStatuses(),
				"lease_vendor_id": l.vendors(), "loan_vendor_id": l.vendors(), "trailer.classification_id": l.trailerClassifications()}}},
		{vendors, "/vendors", bulk.Flat[dto.VendorResponse, dto.CreateVendorRequest, dto.UpdateVendorRequest]{
			Name: "vendors", Store: NewVendorStore(q), CustomFields: customFieldsFor(d, "vendors")}},
		{member, "/locations", bulk.Flat[dto.LocationResponse, dto.CreateLocationRequest, dto.UpdateLocationRequest]{
			Name: "locations", Store: NewLocationStore(q)}},
		{parts, "/part-categories", bulk.Flat[dto.PartCategoryResponse, dto.CreatePartCategoryRequest, dto.UpdatePartCategoryRequest]{
			Name: "part-categories", Store: NewPartCategoryStore(q)}},
		{parts, "/part-manufacturers", bulk.Flat[dto.PartManufacturerResponse, dto.CreatePartManufacturerRequest, dto.UpdatePartManufacturerRequest]{
			Name: "part-manufacturers", Store: NewPartManufacturerStore(q)}},
		{parts, "/parts", bulk.Flat[dto.PartResponse, dto.CreatePartRequest, dto.UpdatePartRequest]{
			Name: "parts", Store: NewPartStore(q), CustomFields: customFieldsFor(d, "parts"),
			References: refs{"part_category_id": l.partCategories(), "part_manufacturer_id": l.partManufacturers(), "measurement_unit_id": l.measurementUnits()}}},
		{inventory, "/measurement-units", bulk.Flat[dto.MeasurementUnitResponse, dto.CreateMeasurementUnitRequest, dto.UpdateMeasurementUnitRequest]{
			Name: "measurement-units", Store: NewMeasurementUnitStore(q)}},
		{inventory, "/part-locations", bulk.Flat[dto.PartLocationResponse, dto.CreatePartLocationRequest, dto.UpdatePartLocationRequest]{
			Name: "part-locations", Store: NewPartLocationStore(q), References: refs{"location_id": l.locations()}}},
		{inventory, "/inventory-adjustment-reasons", bulk.Flat[dto.InventoryAdjustmentReasonResponse, dto.CreateInventoryAdjustmentReasonRequest, dto.UpdateInventoryAdjustmentReasonRequest]{
			Name: "inventory-adjustment-reasons", Store: NewInventoryAdjustmentReasonStore(q)}},
		{inventory, "/part-inventory", bulk.Nested[dto.PartInventoryResponse, dto.CreatePartInventoryRequest, dto.UpdatePartInventoryRequest]{
			Name: "part-inventory", Store: NewPartInventoryStore(q), ParentColumn: "part_id",
			References: refs{"part_id": l.parts(), "location_id": l.partLocations()}}},
		{workOrders, "/work-order-statuses", bulk.Flat[dto.WorkOrderStatusResponse, dto.CreateWorkOrderStatusRequest, dto.UpdateWorkOrderStatusRequest]{
			Name: "work-order-statuses", Store: NewWorkOrderStatusStore(q)}},
		{workOrders, "/work-orders", bulk.Flat[dto.WorkOrderResponse, dto.CreateWorkOrderRequest, dto.UpdateWorkOrderRequest]{
			Name: "work-orders", Store: NewWorkOrderStore(q, d.Pool), CustomFields: customFieldsFor(d, "work-orders"),
			References: refs{"location_id": l.locations(), "asset_id": l.assets(), "status_id": l.workOrderStatuses(), "vendor_id": l.vendors(),
				"assigned_to_id": l.employees(), "issued_by_id": l.employees(), "fault_id": l.faults()}}},
		{issues, "/issues", bulk.Flat[dto.IssueResponse, dto.CreateIssueRequest, dto.UpdateIssueRequest]{
			Name: "issues", Store: NewIssueStore(q), CustomFields: customFieldsFor(d, "issues"),
			Skip: []string{"inspection_submission_id", "resolvable_id"},
			References: refs{"asset_id": l.assets(), "priority_id": l.issuePriorities(), "fault_id": l.faults(),
				"reported_by_id": l.employees(), "resolved_by_id": l.employees(), "reopened_by_id": l.employees(), "closed_by_id": l.employees()}}},
		{issues, "/issue-priorities", bulk.Flat[dto.IssuePriorityResponse, dto.CreateIssuePriorityRequest, dto.UpdateIssuePriorityRequest]{
			Name: "issue-priorities", Store: NewIssuePriorityStore(q)}},
		{issues, "/faults", bulk.Flat[dto.FaultResponse, dto.CreateFaultRequest, dto.UpdateFaultRequest]{
			Name: "faults", Store: NewFaultStore(q)}},
		{service, "/service-tasks", bulk.Flat[dto.ServiceTaskResponse, dto.CreateServiceTaskRequest, dto.UpdateServiceTaskRequest]{
			Name: "service-tasks", Store: NewServiceTaskStore(q), References: refs{"parent_task_id": l.serviceTasks()}}},
		{service, "/service-reminders", bulk.Flat[dto.ServiceReminderResponse, dto.CreateServiceReminderRequest, dto.UpdateServiceReminderRequest]{
			Name: "service-reminders", Store: NewServiceReminderStore(q), Skip: []string{"last_service_entry_id"},
			References: refs{"asset_id": l.assets(), "service_task_id": l.serviceTasks()}}},
		{service, "/service-entries", bulk.Flat[dto.ServiceEntryResponse, dto.CreateServiceEntryRequest, dto.UpdateServiceEntryRequest]{
			Name: "service-entries", Store: NewServiceEntryStore(q, d.Pool), CustomFields: customFieldsFor(d, "service-entries"),
			References: refs{"asset_id": l.assets(), "vendor_id": l.vendors()}}},
		{tires, "/tire-models", bulk.Flat[dto.TireModelResponse, dto.CreateTireModelRequest, dto.UpdateTireModelRequest]{
			Name: "tire-models", Store: NewTireModelStore(q)}},
		{tires, "/tires", bulk.Flat[dto.TireResponse, dto.CreateTireRequest, dto.UpdateTireRequest]{
			Name: "tires", Store: NewTireStore(q),
			References: refs{"tire_model_id": l.tireModels(), "current_vehicle_id": l.assets(), "vendor_id": l.vendors()}}},
		{tires, "/tire-inspections", bulk.Nested[dto.TireInspectionResponse, dto.CreateTireInspectionRequest, dto.UpdateTireInspectionRequest]{
			Name: "tire-inspections", Store: NewTireInspectionStore(q), ParentColumn: "tire_id",
			References: refs{"tire_id": l.tires(), "measured_by_id": l.employees()}}},
		{fuel, "/fuel-types", bulk.Flat[dto.FuelTypeResponse, dto.CreateFuelTypeRequest, dto.UpdateFuelTypeRequest]{
			Name: "fuel-types", Store: NewFuelTypeStore(q)}},
		{fuel, "/fuel-entries", bulk.Nested[dto.FuelEntryResponse, dto.CreateFuelEntryRequest, dto.UpdateFuelEntryRequest]{
			Name: "fuel-entries", Store: NewFuelEntryStore(q, d.Pool), ParentColumn: "asset_id",
			References: fuelEntryRefs(l)}},
		{assets, "/trailer-classifications", bulk.Flat[dto.TrailerClassificationResponse, dto.CreateTrailerClassificationRequest, dto.UpdateTrailerClassificationRequest]{
			Name: "trailer-classifications", Store: NewTrailerClassificationStore(q)}},
		{warranties, "/warranties", bulk.Flat[dto.WarrantyResponse, dto.CreateWarrantyRequest, dto.UpdateWarrantyRequest]{
			Name: "warranties", Store: NewWarrantyStore(q),
			References: refs{"provider_id": l.vendors(), "asset_id": l.assets(), "part_id": l.parts()}}},
		{mileageGoals, "/weekly-mileage-goals", bulk.Flat[dto.WeeklyMileageGoalResponse, dto.CreateWeeklyMileageGoalRequest, dto.UpdateWeeklyMileageGoalRequest]{
			Name: "weekly-mileage-goals", Store: NewWeeklyMileageGoalStore(q)}},
		{employees, "/employees", bulk.Flat[dto.EmployeeResponse, dto.CreateEmployeeRequest, dto.UpdateEmployeeRequest]{
			Name: "employees", Store: NewEmployeeStore(q, d.Pool), CustomFields: customFieldsFor(d, "employees"),
			Skip:       []string{"user_id", "default_company_id", "table_preferences", "dashboard_preferences"},
			Kinds:      map[string]bulk.Kind{"role_id": bulk.KindInt},
			References: refs{"group_id": l.groups(), "role_id": l.roles()}}},
		{employees, "/groups", bulk.Flat[dto.GroupResponse, dto.CreateGroupRequest, dto.UpdateGroupRequest]{
			Name: "groups", Store: NewGroupStore(q), References: refs{"parent_id": l.groups()}}},
		{assets, "/asset-types", bulk.Flat[dto.AssetTypeResponse, dto.CreateAssetTypeRequest, dto.UpdateAssetTypeRequest]{
			Name: "asset-types", Store: NewAssetTypeStore(q)}},
		{assets, "/asset-statuses", bulk.Flat[dto.AssetStatusResponse, dto.CreateAssetStatusRequest, dto.UpdateAssetStatusRequest]{
			Name: "asset-statuses", Store: NewAssetStatusStore(q)}},
		{assets, "/catalog-options", bulk.Flat[dto.CatalogOptionResponse, dto.CreateCatalogOptionRequest, dto.UpdateCatalogOptionRequest]{
			Name: "catalog-options", Store: NewCatalogOptionStore(q)}},
		{member, "/vehicle-makes", bulk.Flat[dto.VehicleMakeResponse, dto.CreateVehicleMakeRequest, dto.UpdateVehicleMakeRequest]{
			Name: "vehicle-makes", Store: NewVehicleMakeStore(q)}},
		{member, "/vehicle-models", bulk.Flat[dto.VehicleModelResponse, dto.CreateVehicleModelRequest, dto.UpdateVehicleModelRequest]{
			Name: "vehicle-models", Store: NewVehicleModelStore(q), References: refs{"make_id": l.vehicleMakes()}}},
	}
}

// fuelEntryRefs declares the parent column, plus vendor and employee only if
// CreateFuelEntryRequest has those fields (TestBulkRegistryReferencesExistingColumns
// fails otherwise: remove the ones it names).
func fuelEntryRefs(l bulkLookups) refs {
	return refs{"asset_id": l.assets(), "vendor_id": l.vendors(), "employee_id": l.employees()}
}

// registerBulkImports wires import and template routes for every section.
func registerBulkImports(g bulkGroups, d Deps) {
	h := NewBulkHandler(d.Pool)
	for _, e := range bulkImporters(d) {
		h.Register(e.group(g), e.path, e.imp)
	}
}
