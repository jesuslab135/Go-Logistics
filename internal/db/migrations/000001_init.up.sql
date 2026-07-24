-- 000001_init.up.sql
-- Consolidated initial schema, ported from the Django app (Fleet-Personal-Backend,
-- api app, migrations 0001..0062, final state). Postgres, USE_TZ=True so all
-- datetimes are timestamptz. Table names are word-separated snake_case (no `api_` prefix).
--
-- Notes on the port:
--   * FKs to Django-internal tables (auth_user, django_content_type) are kept as
--     plain bigint columns with NO foreign key, pending the Go auth rework.
--   * Django's PositiveIntegerField emits no DB CHECK constraint (validation is
--     app-level only), so none are reproduced here.
--   * auto_now / auto_now_add / timezone.now defaults are set by the app layer,
--     so those columns carry NOT NULL but no DB default. All other Django
--     `default=` values are reproduced as DB defaults.
--   * FK constraints are declared in a single ALTER block so table order is
--     irrelevant. Django's cryptic auto index names are replaced with clean ones.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ---------------------------------------------------------------------------
-- Organization
-- ---------------------------------------------------------------------------

CREATE TABLE company (
    id                    bigserial PRIMARY KEY,
    name                  varchar(200) NOT NULL,
    tax_id                varchar(50)  NOT NULL UNIQUE,
    address               text         NOT NULL,
    created_at            timestamptz  NOT NULL,
    phone                 varchar(20)  NOT NULL,
    email                 varchar(254) NOT NULL,
    website               varchar(200) NOT NULL,
    logo                  varchar(100),
    city                  varchar(100) NOT NULL,
    region                varchar(50)  NOT NULL,
    postal_code           varchar(20)  NOT NULL,
    country               varchar(50)  NOT NULL DEFAULT 'MX',
    timezone              varchar(50)  NOT NULL DEFAULT 'America/Mexico_City',
    currency              varchar(3)   NOT NULL DEFAULT 'MXN',
    system_of_measurement varchar(10)  NOT NULL DEFAULT 'metric'
);

CREATE TABLE role (
    id          bigserial PRIMARY KEY,
    company_id  bigint       NOT NULL,
    name        varchar(100) NOT NULL,
    is_admin    boolean      NOT NULL DEFAULT false,
    permissions jsonb        NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE "group" (
    id         bigserial PRIMARY KEY,
    company_id bigint       NOT NULL,
    name       varchar(100) NOT NULL,
    parent_id  bigint,
    ancestry   varchar(500) NOT NULL,
    is_default boolean      NOT NULL DEFAULT false,
    created_at timestamptz  NOT NULL,
    updated_at timestamptz  NOT NULL,
    CONSTRAINT uq_group UNIQUE (company_id, name)
);

CREATE TABLE employee (
    id                    bigserial PRIMARY KEY,
    user_id               bigint UNIQUE,           -- auth_user (no FK)
    default_company_id    bigint,
    first_name            varchar(100) NOT NULL,
    last_name             varchar(100) NOT NULL,
    employee_id           varchar(50)  NOT NULL,
    role_id               bigint,
    is_active             boolean      NOT NULL DEFAULT true,
    email                 varchar(254) NOT NULL,
    mobile_phone          varchar(20)  NOT NULL,
    work_phone            varchar(20)  NOT NULL,
    job_title             varchar(100) NOT NULL,
    start_date            date,
    leave_date            date,
    birth_date            date,
    hourly_labor_rate     numeric(8,2),
    is_technician         boolean      NOT NULL DEFAULT false,
    is_vehicle_operator   boolean      NOT NULL DEFAULT false,
    is_account_owner      boolean      NOT NULL DEFAULT false,
    license_class         varchar(10)  NOT NULL,
    license_number        varchar(50)  NOT NULL,
    license_state         varchar(50)  NOT NULL,
    license_expiry        date,
    street_address        varchar(200) NOT NULL,
    city                  varchar(100) NOT NULL,
    region                varchar(50)  NOT NULL,
    postal_code           varchar(20)  NOT NULL,
    country               varchar(50)  NOT NULL,
    group_id              bigint,
    custom_fields         jsonb        NOT NULL DEFAULT '{}'::jsonb,
    table_preferences     jsonb        NOT NULL DEFAULT '{}'::jsonb,
    dashboard_preferences jsonb        NOT NULL DEFAULT '{}'::jsonb,
    updated_at            timestamptz  NOT NULL
);

-- ---------------------------------------------------------------------------
-- Assets
-- ---------------------------------------------------------------------------

CREATE TABLE asset_type (
    id          bigserial PRIMARY KEY,
    company_id  bigint       NOT NULL,
    name        varchar(100) NOT NULL,
    category    varchar(50)  NOT NULL,
    description text         NOT NULL
);

CREATE TABLE asset_status (
    id         bigserial PRIMARY KEY,
    company_id bigint      NOT NULL,
    name       varchar(50) NOT NULL,
    color_code varchar(7)  NOT NULL DEFAULT '#28a745',
    CONSTRAINT uq_asset_status UNIQUE (company_id, name)
);

CREATE TABLE catalog_option (
    id         bigserial PRIMARY KEY,
    company_id bigint       NOT NULL,
    category   varchar(30)  NOT NULL,
    value      varchar(100) NOT NULL,
    CONSTRAINT uq_catalog_option UNIQUE (company_id, category, value)
);

CREATE TABLE asset (
    id                            bigserial PRIMARY KEY,
    company_id                    bigint       NOT NULL,
    name                          varchar(100) NOT NULL,
    vin_sn                        varchar(100) NOT NULL,
    msrp                          numeric(12,2),
    generate_expenses             boolean      NOT NULL DEFAULT false,
    asset_type_id                 bigint,
    status_id                     bigint,
    lease_vendor_id               bigint,
    vehicle_type                  varchar(20)  NOT NULL DEFAULT 'CAR',
    ownership_type                varchar(20)  NOT NULL DEFAULT 'OWNED',
    labels                        jsonb        NOT NULL DEFAULT '[]'::jsonb,
    linked_vehicles               jsonb        NOT NULL DEFAULT '[]'::jsonb,
    loan_start_date               date,
    loan_end_date                 date,
    monthly_payment               numeric(12,2),
    number_of_payments            integer,
    lease_number                  varchar(100) NOT NULL,
    lease_start_date              date,
    lease_end_date                date,
    excess_mileage_charge         numeric(10,2),
    owner_company_id              bigint,
    year                          integer,
    make                          varchar(100) NOT NULL,
    model                         varchar(100) NOT NULL,
    trim                          varchar(100) NOT NULL,
    color                         varchar(50)  NOT NULL,
    license_plate                 varchar(20)  NOT NULL,
    "group"                       varchar(100) NOT NULL,
    photo                         varchar(100),
    meter_unit                    varchar(5)   NOT NULL DEFAULT 'mi',
    current_meter                 numeric(15,2),
    secondary_meter_unit          varchar(20)  NOT NULL,
    secondary_meter_value         numeric(15,2),
    fuel_type                     varchar(20)  NOT NULL,
    body_type                     varchar(100) NOT NULL,
    body_subtype                  varchar(100) NOT NULL,
    registration_state            varchar(50)  NOT NULL,
    purchase_date                 date,
    purchase_price                numeric(12,2),
    purchase_vendor               varchar(200) NOT NULL,
    purchase_meter                numeric(15,2),
    in_service_date               date,
    in_service_meter              numeric(15,2),
    out_of_service_date           date,
    out_of_service_meter          numeric(15,2),
    estimated_service_months      integer,
    estimated_replacement_mileage integer,
    estimated_resale_price        numeric(12,2),
    acquisition_type              varchar(20)  NOT NULL,
    monthly_cost                  numeric(12,2),
    acquisition_date              date,
    loan_amount                   numeric(12,2),
    capitalized_cost              numeric(12,2),
    down_payment                  numeric(12,2),
    annual_percentage_rate        numeric(5,2),
    first_payment_date            date,
    residual_value                numeric(12,2),
    mileage_cap                   integer,
    notes                         text         NOT NULL,
    archived_at                   timestamptz,
    external_id                   varchar(100) NOT NULL,
    custom_fields                 jsonb        NOT NULL DEFAULT '{}'::jsonb,
    fuel_volume_units             varchar(20)  NOT NULL DEFAULT 'liters',
    current_meter_date            date,
    loan_account_number           varchar(100) NOT NULL,
    loan_notes                    text         NOT NULL,
    loan_vendor_id                bigint,
    loan_started_at               date,
    loan_ended_at                 date,
    updated_at                    timestamptz  NOT NULL,
    CONSTRAINT uq_asset UNIQUE (company_id, vin_sn)
);

CREATE TABLE vehicle (
    asset_id                bigint PRIMARY KEY,
    engine_serial           varchar(100) NOT NULL,
    engine_description      varchar(200) NOT NULL,
    engine_brand            varchar(100) NOT NULL,
    engine_cylinders        integer,
    engine_displacement     varchar(50)  NOT NULL,
    max_hp                  varchar(50)  NOT NULL,
    max_torque              varchar(50)  NOT NULL,
    oil_capacity            varchar(50)  NOT NULL,
    engine_aspiration       varchar(50)  NOT NULL,
    engine_block_type       varchar(50)  NOT NULL,
    engine_compression      varchar(50)  NOT NULL,
    fuel_induction          varchar(50)  NOT NULL,
    transmission_description varchar(200) NOT NULL,
    transmission_brand      varchar(100) NOT NULL,
    transmission_type       varchar(50)  NOT NULL,
    transmission_gears      integer,
    drive_type              varchar(20)  NOT NULL,
    brake_system            varchar(20)  NOT NULL,
    axles                   integer      NOT NULL DEFAULT 2,
    differential            varchar(100) NOT NULL,
    differential_ratio      varchar(20)  NOT NULL,
    suspension              varchar(100) NOT NULL,
    tire_size               varchar(50)  NOT NULL,
    height                  varchar(50)  NOT NULL,
    length                  varchar(50)  NOT NULL,
    width                   varchar(50)  NOT NULL,
    wheelbase               varchar(50)  NOT NULL,
    curb_weight             numeric(10,2),
    gvwr                    numeric(10,2),
    max_weight_capacity     numeric(10,2),
    max_payload             numeric(10,2),
    towing_capacity         numeric(10,2),
    fuel_tank_capacity      numeric(10,2),
    fuel_tank_2_capacity    numeric(10,2),
    tank_1_security         varchar(10)  NOT NULL,
    tank_2_security         varchar(10)  NOT NULL,
    interior_volume         varchar(50)  NOT NULL,
    passenger_volume        varchar(50)  NOT NULL,
    ground_clearance        varchar(50)  NOT NULL,
    engine_bore             varchar(50)  NOT NULL,
    redline_rpm             varchar(50)  NOT NULL,
    stroke                  varchar(50)  NOT NULL,
    valves                  integer,
    front_track_width       varchar(50)  NOT NULL,
    rear_track_width        varchar(50)  NOT NULL,
    front_wheel_diameter    varchar(50)  NOT NULL,
    rear_wheel_diameter     varchar(50)  NOT NULL,
    fuel_quality            varchar(100) NOT NULL,
    cab_type                varchar(20)  NOT NULL,
    truck_config            varchar(20)  NOT NULL,
    telematics_system       varchar(100) NOT NULL,
    emission_standard       varchar(50)  NOT NULL,
    emission_active         boolean,
    sweetspot_rpm           varchar(50)  NOT NULL,
    pedal_speed_limit       integer,
    cruise_speed_limit      integer,
    idle_shutdown           varchar(50)  NOT NULL,
    apu_type                varchar(50)  NOT NULL,
    has_spare_tire_rack     boolean      NOT NULL DEFAULT false,
    fuel_group              varchar(100) NOT NULL,
    cell                    varchar(100) NOT NULL,
    service_type            varchar(100) NOT NULL,
    supervisor              varchar(200) NOT NULL,
    management              varchar(200) NOT NULL,
    is_active_dispatch      boolean      NOT NULL DEFAULT true,
    is_active_company       boolean      NOT NULL DEFAULT true,
    duty_type               varchar(20)  NOT NULL,
    weight_class            varchar(50)  NOT NULL,
    cargo_volume            numeric(10,2),
    bed_length              numeric(10,2),
    front_tire_psi          numeric(5,1),
    rear_tire_psi           numeric(5,1),
    front_tire_type         varchar(50)  NOT NULL,
    rear_tire_type          varchar(50)  NOT NULL,
    rear_axle_type          varchar(100) NOT NULL,
    operator                varchar(255) NOT NULL,
    epa_city                numeric(5,1),
    epa_highway             numeric(5,1),
    epa_combined            numeric(5,1)
);

CREATE TABLE trailer (
    asset_id               bigint PRIMARY KEY,
    trailer_type           varchar(20)  NOT NULL,
    classification         varchar(100) NOT NULL,
    classification_2       varchar(100) NOT NULL,
    size                   varchar(20)  NOT NULL,
    suspension             varchar(100) NOT NULL,
    owner_name             varchar(200) NOT NULL,
    financing              varchar(20)  NOT NULL,
    supplier               varchar(200) NOT NULL,
    license_plate_us       varchar(20)  NOT NULL,
    license_plate_us_state varchar(10)  NOT NULL,
    license_plate_mx       varchar(20)  NOT NULL,
    doors                  varchar(50)  NOT NULL,
    walls                  varchar(50)  NOT NULL,
    rail_post              varchar(50)  NOT NULL,
    skylight               boolean      NOT NULL DEFAULT false,
    floor_type             varchar(50)  NOT NULL,
    roof_type              varchar(50)  NOT NULL,
    hazmat_type            varchar(100) NOT NULL,
    aero_kit_type          varchar(50)  NOT NULL,
    gps_provider           varchar(100) NOT NULL,
    gps_serial             varchar(100) NOT NULL,
    gps_signal_status      varchar(50)  NOT NULL,
    gps_contract_start     date,
    gps_contract_end       date,
    gps_contract_reference varchar(200) NOT NULL,
    contract_start         date,
    contract_end           date,
    contract_period        varchar(50)  NOT NULL,
    contract_reference     varchar(100) NOT NULL,
    fumigation_date        date,
    fumigation_cert        boolean      NOT NULL DEFAULT false,
    registration_date      date,
    waterproofing_date     date,
    decommission_reason    text         NOT NULL,
    decommission_date      date,
    operational_use        varchar(20)  NOT NULL,
    operation_zone         varchar(100) NOT NULL
);

CREATE TABLE asset_trailer_assignment (
    id              bigserial PRIMARY KEY,
    asset_id        bigint      NOT NULL,
    trailer_id      bigint      NOT NULL,
    position        integer     NOT NULL DEFAULT 1,
    assigned_date   timestamptz NOT NULL,
    unassigned_date timestamptz,
    assigned_by_id  bigint,
    is_active       boolean     NOT NULL DEFAULT true,
    notes           text        NOT NULL
);

CREATE TABLE vehicle_make (
    id         bigserial PRIMARY KEY,
    company_id bigint       NOT NULL,
    name       varchar(100) NOT NULL,
    created_at timestamptz  NOT NULL,
    updated_at timestamptz  NOT NULL,
    CONSTRAINT uq_vehicle_make UNIQUE (company_id, name)
);

CREATE TABLE vehicle_model (
    id         bigserial PRIMARY KEY,
    company_id bigint       NOT NULL,
    name       varchar(100) NOT NULL,
    make_id    bigint,
    created_at timestamptz  NOT NULL,
    updated_at timestamptz  NOT NULL,
    CONSTRAINT uq_vehicle_model UNIQUE (company_id, name, make_id)
);

-- ---------------------------------------------------------------------------
-- Parts / inventory
-- ---------------------------------------------------------------------------

CREATE TABLE part_category (
    id          bigserial PRIMARY KEY,
    company_id  bigint       NOT NULL,
    name        varchar(100) NOT NULL,
    description text         NOT NULL,
    created_at  timestamptz  NOT NULL,
    CONSTRAINT uq_part_category UNIQUE (company_id, name)
);

CREATE TABLE part_manufacturer (
    id         bigserial PRIMARY KEY,
    company_id bigint       NOT NULL,
    name       varchar(200) NOT NULL,
    website    varchar(200) NOT NULL,
    created_at timestamptz  NOT NULL,
    CONSTRAINT uq_part_manufacturer UNIQUE (company_id, name)
);

CREATE TABLE measurement_unit (
    id           bigserial PRIMARY KEY,
    company_id   bigint      NOT NULL,
    name         varchar(50) NOT NULL,
    abbreviation varchar(10) NOT NULL,
    created_at   timestamptz NOT NULL,
    CONSTRAINT uq_measurement_unit UNIQUE (company_id, name)
);

CREATE TABLE part (
    id                       bigserial PRIMARY KEY,
    company_id               bigint       NOT NULL,
    part_number              varchar(100) NOT NULL,
    description              text         NOT NULL,
    useful_life_months       integer,
    useful_life_distance     integer,
    part_category_id         bigint,
    part_manufacturer_id     bigint,
    measurement_unit_id      bigint,
    manufacturer_part_number varchar(100) NOT NULL,
    supplier_part_number     varchar(100) NOT NULL DEFAULT '',
    upc                      varchar(50)  NOT NULL,
    unit_cost                numeric(12,2),
    inventory_item           boolean      NOT NULL DEFAULT true,
    archived_at              timestamptz,
    custom_fields            jsonb        NOT NULL DEFAULT '{}'::jsonb,
    created_at               timestamptz  NOT NULL,
    updated_at               timestamptz  NOT NULL,
    CONSTRAINT uq_part UNIQUE (company_id, part_number)
);

CREATE TABLE location (
    id         bigserial PRIMARY KEY,
    company_id bigint       NOT NULL,
    name       varchar(100) NOT NULL,
    is_active  boolean      NOT NULL DEFAULT true,
    CONSTRAINT uq_location UNIQUE (company_id, name)
);

CREATE TABLE part_location (
    id          bigserial PRIMARY KEY,
    company_id  bigint       NOT NULL,
    name        varchar(100) NOT NULL,
    address     varchar(200) NOT NULL,
    city        varchar(100) NOT NULL,
    region      varchar(50)  NOT NULL,
    location_id bigint,
    created_at  timestamptz  NOT NULL,
    updated_at  timestamptz  NOT NULL,
    CONSTRAINT uq_part_location UNIQUE (company_id, name)
);

-- PartLocationDetail: Meta.db_table override -> part_inventory
CREATE TABLE part_inventory (
    id                            bigserial PRIMARY KEY,
    part_id                       bigint        NOT NULL,
    location_id                   bigint        NOT NULL,
    available_quantity            numeric(10,2) NOT NULL,
    expiry_date                   date,
    aisle                         varchar(50)   NOT NULL,
    "row"                         varchar(50)   NOT NULL,
    bin                           varchar(50)   NOT NULL,
    reorder_point                 integer,
    reorder_point_enabled         boolean       NOT NULL DEFAULT false,
    reorder_quantity              integer,
    reorder_point_lead_time_days  integer,
    active                        boolean       NOT NULL DEFAULT true,
    track_inventory               boolean       NOT NULL DEFAULT true,
    average_unit_cost             numeric(12,2),
    available_quantity_updated_at timestamptz,
    created_at                    timestamptz   NOT NULL,
    updated_at                    timestamptz   NOT NULL,
    CONSTRAINT uq_part_inventory UNIQUE (part_id, location_id)
);

CREATE TABLE inventory_adjustment_reason (
    id         bigserial PRIMARY KEY,
    company_id bigint       NOT NULL,
    name       varchar(100) NOT NULL,
    created_at timestamptz  NOT NULL,
    CONSTRAINT uq_inventory_adjustment_reason UNIQUE (company_id, name)
);

CREATE TABLE inventory_journal_entry (
    id                       bigserial PRIMARY KEY,
    company_id               bigint        NOT NULL,
    part_id                  bigint        NOT NULL,
    part_location_detail_id  bigint        NOT NULL,
    user_id                  bigint,
    previous_quantity        numeric(10,2) NOT NULL,
    adjustment_quantity      numeric(10,2) NOT NULL,
    current_quantity         numeric(10,2) NOT NULL,
    unit_cost                numeric(12,2) NOT NULL DEFAULT 0,
    reason_id                bigint,
    work_order_id            bigint,
    purchase_order_line_id   bigint,
    vendor_id                bigint,
    adjustment_type          varchar(20)   NOT NULL DEFAULT 'manual',
    transfer_part_location_id bigint,
    notes                    text          NOT NULL,
    created_at               timestamptz   NOT NULL
);

-- ---------------------------------------------------------------------------
-- Vendors
-- ---------------------------------------------------------------------------

CREATE TABLE vendor (
    id                     bigserial PRIMARY KEY,
    company_id             bigint       NOT NULL,
    name                   varchar(200) NOT NULL,
    is_mobile_service      boolean      NOT NULL DEFAULT false,
    street_address         varchar(200) NOT NULL,
    street_address_line_2  varchar(200) NOT NULL,
    city                   varchar(100) NOT NULL,
    region                 varchar(50)  NOT NULL,
    postal_code            varchar(20)  NOT NULL,
    country                varchar(50)  NOT NULL,
    phone                  varchar(20)  NOT NULL,
    website                varchar(200) NOT NULL,
    contact_name           varchar(200) NOT NULL,
    contact_phone          varchar(20)  NOT NULL,
    contact_email          varchar(254) NOT NULL,
    external_id            varchar(100) NOT NULL,
    latitude               numeric(10,7),
    longitude              numeric(10,7),
    is_fuel_vendor         boolean      NOT NULL DEFAULT false,
    is_service_vendor      boolean      NOT NULL DEFAULT false,
    is_parts_vendor        boolean      NOT NULL DEFAULT false,
    labels                 jsonb        NOT NULL DEFAULT '[]'::jsonb,
    archived_at            timestamptz,
    custom_fields          jsonb        NOT NULL DEFAULT '{}'::jsonb,
    created_at             timestamptz  NOT NULL,
    updated_at             timestamptz  NOT NULL,
    CONSTRAINT uq_vendor UNIQUE (company_id, name)
);

-- ---------------------------------------------------------------------------
-- Work orders
-- ---------------------------------------------------------------------------

CREATE TABLE work_order_status (
    id                 bigserial PRIMARY KEY,
    company_id         bigint       NOT NULL,
    name               varchar(100) NOT NULL,
    description        text         NOT NULL DEFAULT '',
    color              varchar(7)   NOT NULL DEFAULT '#6C757D',
    is_default         boolean      NOT NULL DEFAULT false,
    marks_as_completed boolean      NOT NULL DEFAULT false,
    position           integer      NOT NULL DEFAULT 0,
    CONSTRAINT uq_work_order_status UNIQUE (company_id, name)
);

CREATE TABLE work_order (
    id                      bigserial PRIMARY KEY,
    location_id             bigint,
    company_id              bigint        NOT NULL,
    number                  varchar(50)   NOT NULL DEFAULT '',
    description             text          NOT NULL DEFAULT '',
    asset_id                bigint        NOT NULL,
    status_id               bigint        NOT NULL,
    vendor_id               bigint,
    assigned_to_id          bigint,
    issued_by_id            bigint,
    fault_id                bigint,
    issued_at               timestamptz   NOT NULL,
    scheduled_at            timestamptz,
    started_at              timestamptz,
    expected_completed_at   timestamptz,
    completed_at            timestamptz,
    starting_meter          numeric(15,2),
    ending_meter            numeric(15,2),
    duration_seconds        integer,
    labor_time_seconds      integer,
    parts_markup_type       varchar(10)   NOT NULL DEFAULT 'PERCENTAGE',
    parts_markup            numeric(12,2) NOT NULL DEFAULT 0,
    parts_markup_percentage numeric(5,2)  NOT NULL DEFAULT 0,
    labor_markup_type       varchar(10)   NOT NULL DEFAULT 'PERCENTAGE',
    labor_markup            numeric(12,2) NOT NULL DEFAULT 0,
    labor_markup_percentage numeric(5,2)  NOT NULL DEFAULT 0,
    parts_subtotal          numeric(12,2) NOT NULL DEFAULT 0,
    labor_subtotal          numeric(12,2) NOT NULL DEFAULT 0,
    subtotal                numeric(12,2) NOT NULL DEFAULT 0,
    discount                numeric(12,2) NOT NULL DEFAULT 0,
    discount_type           varchar(10)   NOT NULL DEFAULT 'FIXED',
    tax_1                   numeric(12,2) NOT NULL DEFAULT 0,
    tax_1_type              varchar(10)   NOT NULL DEFAULT 'PERCENTAGE',
    tax_1_percentage        numeric(5,2)  NOT NULL DEFAULT 0,
    tax_2                   numeric(12,2) NOT NULL DEFAULT 0,
    tax_2_type              varchar(10)   NOT NULL DEFAULT 'PERCENTAGE',
    tax_2_percentage        numeric(5,2)  NOT NULL DEFAULT 0,
    total_amount            numeric(12,2) NOT NULL DEFAULT 0,
    invoice_number          varchar(100)  NOT NULL DEFAULT '',
    purchase_order_number   varchar(100)  NOT NULL DEFAULT '',
    comments_count          integer       NOT NULL DEFAULT 0,
    images_count            integer       NOT NULL DEFAULT 0,
    documents_count         integer       NOT NULL DEFAULT 0,
    labels                  jsonb         NOT NULL DEFAULT '[]'::jsonb,
    custom_fields           jsonb         NOT NULL DEFAULT '{}'::jsonb,
    created_at              timestamptz   NOT NULL,
    updated_at              timestamptz   NOT NULL
);

CREATE TABLE work_order_line_item (
    id             bigserial PRIMARY KEY,
    work_order_id  bigint        NOT NULL,
    line_item_type varchar(20)   NOT NULL,
    title          varchar(255)  NOT NULL,
    description    text          NOT NULL DEFAULT '',
    position       integer       NOT NULL DEFAULT 0,
    service_task   varchar(255),
    parts_cost     numeric(12,2) NOT NULL DEFAULT 0,
    labor_cost     numeric(12,2) NOT NULL DEFAULT 0,
    subtotal       numeric(12,2) NOT NULL DEFAULT 0,
    created_at     timestamptz   NOT NULL,
    updated_at     timestamptz   NOT NULL
);

CREATE TABLE work_order_sub_line_item (
    id                      bigserial PRIMARY KEY,
    line_item_id            bigint        NOT NULL,
    item_type               varchar(10)   NOT NULL,
    description             varchar(255)  NOT NULL DEFAULT '',
    position                integer       NOT NULL DEFAULT 0,
    part_id                 bigint,
    part_location_detail_id bigint,
    technician_id           bigint,
    unit_cost               numeric(12,2) NOT NULL DEFAULT 0,
    quantity                numeric(10,2) NOT NULL DEFAULT 1,
    created_at              timestamptz   NOT NULL,
    updated_at              timestamptz   NOT NULL
);

CREATE TABLE labor_time_entry (
    id                 bigserial PRIMARY KEY,
    sub_line_item_id   bigint       NOT NULL,
    technician_id      bigint       NOT NULL,
    started_at         timestamptz  NOT NULL,
    ended_at           timestamptz,
    duration_seconds   integer,
    is_active          boolean      NOT NULL DEFAULT true,
    clock_in_latitude  numeric(10,7),
    clock_in_longitude numeric(10,7),
    clock_out_latitude numeric(10,7),
    clock_out_longitude numeric(10,7),
    created_at         timestamptz  NOT NULL
);

CREATE TABLE work_order_status_log (
    id            bigserial PRIMARY KEY,
    work_order_id bigint      NOT NULL,
    status_id     bigint      NOT NULL,
    changed_at    timestamptz NOT NULL
);

-- ---------------------------------------------------------------------------
-- Issues / faults
-- ---------------------------------------------------------------------------

CREATE TABLE issue_priority (
    id         bigserial PRIMARY KEY,
    company_id bigint      NOT NULL,
    name       varchar(50) NOT NULL,
    color      varchar(7)  NOT NULL DEFAULT '#FFC107',
    position   integer     NOT NULL DEFAULT 0,
    CONSTRAINT uq_issue_priority UNIQUE (company_id, name)
);

CREATE TABLE fault (
    id                     bigserial PRIMARY KEY,
    company_id             bigint         NOT NULL,
    family                 varchar(20)    NOT NULL DEFAULT 'OTRO',
    code                   varchar(50)    NOT NULL DEFAULT '',
    name                   varchar(255)   NOT NULL,
    description            text           NOT NULL DEFAULT '',
    applies_to_asset_types varchar(20)[]  NOT NULL DEFAULT '{}',
    CONSTRAINT uq_fault UNIQUE (company_id, code)
);

CREATE TABLE issue (
    id                       bigserial PRIMARY KEY,
    company_id               bigint       NOT NULL,
    number                   varchar(50)  NOT NULL DEFAULT '',
    asset_id                 bigint       NOT NULL,
    asset_type               varchar(20)  NOT NULL DEFAULT 'VEHICLE',
    name                     varchar(255) NOT NULL DEFAULT '',
    summary                  varchar(255) NOT NULL,
    description              text         NOT NULL DEFAULT '',
    state                    varchar(20)  NOT NULL DEFAULT 'OPEN',
    priority_id              bigint,
    fault_id                 bigint,
    source_type              varchar(20)  NOT NULL DEFAULT 'MANUAL',
    inspection_submission_id bigint,
    reported_at              timestamptz  NOT NULL,
    reported_by_id           bigint       NOT NULL,
    due_date                 date,
    due_meter_value          numeric(15,2),
    due_secondary_meter_value numeric(15,2),
    overdue                  boolean      NOT NULL DEFAULT false,
    resolved_at              timestamptz,
    resolved_by_id           bigint,
    resolution_note          text         NOT NULL DEFAULT '',
    reopened_at              timestamptz,
    reopened_by_id           bigint,
    resolvable_type          varchar(30)  NOT NULL DEFAULT '',
    resolvable_id            integer,
    closed_at                timestamptz,
    closed_by_id             bigint,
    closed_note              text         NOT NULL DEFAULT '',
    external_id              varchar(100) NOT NULL DEFAULT '',
    created_by_workflow      boolean      NOT NULL DEFAULT false,
    comments_count           integer      NOT NULL DEFAULT 0,
    images_count             integer      NOT NULL DEFAULT 0,
    documents_count          integer      NOT NULL DEFAULT 0,
    labels                   jsonb        NOT NULL DEFAULT '[]'::jsonb,
    custom_fields            jsonb        NOT NULL DEFAULT '{}'::jsonb,
    created_at               timestamptz  NOT NULL,
    updated_at               timestamptz  NOT NULL
);

-- ---------------------------------------------------------------------------
-- Purchase orders
-- ---------------------------------------------------------------------------

CREATE TABLE purchase_order (
    id                  bigserial PRIMARY KEY,
    company_id          bigint        NOT NULL,
    number              varchar(50)   NOT NULL,
    description         text          NOT NULL,
    state               varchar(20)   NOT NULL DEFAULT 'DRAFT',
    vendor_id           bigint        NOT NULL,
    destination_id      bigint        NOT NULL,
    discount_type       varchar(10)   NOT NULL DEFAULT 'FIXED',
    discount            numeric(12,2) NOT NULL DEFAULT 0,
    discount_percentage numeric(5,2)  NOT NULL DEFAULT 0,
    tax_1_type          varchar(10)   NOT NULL DEFAULT 'PERCENTAGE',
    tax_1               numeric(12,2) NOT NULL DEFAULT 0,
    tax_1_percentage    numeric(5,2)  NOT NULL DEFAULT 0,
    tax_2_type          varchar(10)   NOT NULL DEFAULT 'PERCENTAGE',
    tax_2               numeric(12,2) NOT NULL DEFAULT 0,
    tax_2_percentage    numeric(5,2)  NOT NULL DEFAULT 0,
    shipping            numeric(12,2) NOT NULL DEFAULT 0,
    subtotal            numeric(12,2) NOT NULL DEFAULT 0,
    total_amount        numeric(12,2) NOT NULL DEFAULT 0,
    created_by_id       bigint,
    submitted_at        timestamptz,
    submitted_by_id     bigint,
    rejected_at         timestamptz,
    rejected_by_id      bigint,
    approved_at         timestamptz,
    approved_by_id      bigint,
    purchased_at        timestamptz,
    received_partial_at timestamptz,
    received_full_at    timestamptz,
    closed_at           timestamptz,
    labels              jsonb         NOT NULL DEFAULT '[]'::jsonb,
    custom_fields       jsonb         NOT NULL DEFAULT '{}'::jsonb,
    created_at          timestamptz   NOT NULL,
    updated_at          timestamptz   NOT NULL
);

CREATE TABLE purchase_order_line_item (
    id                bigserial PRIMARY KEY,
    purchase_order_id bigint        NOT NULL,
    part_id           bigint        NOT NULL,
    quantity          numeric(10,2) NOT NULL,
    total_received    numeric(10,2) NOT NULL DEFAULT 0,
    unit_cost         numeric(12,2) NOT NULL,
    subtotal          numeric(12,2) NOT NULL DEFAULT 0,
    position          integer       NOT NULL DEFAULT 0,
    created_at        timestamptz   NOT NULL,
    updated_at        timestamptz   NOT NULL
);

-- ---------------------------------------------------------------------------
-- Service tasks / reminders / entries
-- ---------------------------------------------------------------------------

CREATE TABLE service_task (
    id                      bigserial PRIMARY KEY,
    company_id              bigint       NOT NULL,
    name                    varchar(255) NOT NULL,
    description             text         NOT NULL,
    expected_duration_seconds integer,
    parent_task_id          bigint,
    archived_at             timestamptz,
    created_at              timestamptz  NOT NULL,
    updated_at              timestamptz  NOT NULL
);

CREATE TABLE service_task_part (
    id              bigserial PRIMARY KEY,
    service_task_id bigint        NOT NULL,
    part_id         bigint        NOT NULL,
    quantity        numeric(10,2) NOT NULL DEFAULT 1,
    position        integer       NOT NULL DEFAULT 0
);

CREATE TABLE service_reminder (
    id                      bigserial PRIMARY KEY,
    company_id              bigint       NOT NULL,
    asset_id                bigint       NOT NULL,
    service_task_id         bigint,
    is_active               boolean      NOT NULL DEFAULT true,
    status                  varchar(20)  NOT NULL DEFAULT 'OK',
    time_interval           integer,
    time_frequency          varchar(10)  NOT NULL,
    next_due_at             timestamptz,
    due_soon_at             timestamptz,
    due_soon_time_threshold integer,
    meter_interval          numeric(15,2),
    next_due_meter_value    numeric(15,2),
    due_soon_meter_value    numeric(15,2),
    due_soon_meter_threshold numeric(15,2),
    snooze_until            timestamptz,
    last_service_entry_id   bigint,
    created_at              timestamptz  NOT NULL,
    updated_at              timestamptz  NOT NULL
);

CREATE TABLE service_entry (
    id                   bigserial PRIMARY KEY,
    company_id           bigint        NOT NULL,
    reference            varchar(100)  NOT NULL,
    status               varchar(20)   NOT NULL DEFAULT 'COMPLETED',
    asset_id             bigint        NOT NULL,
    vendor_id            bigint,
    work_order_id        bigint UNIQUE,
    started_at           timestamptz,
    completed_at         timestamptz,
    meter_value          numeric(15,2),
    parts_subtotal       numeric(12,2) NOT NULL DEFAULT 0,
    labor_subtotal       numeric(12,2) NOT NULL DEFAULT 0,
    subtotal             numeric(12,2) NOT NULL DEFAULT 0,
    discount             numeric(12,2) NOT NULL DEFAULT 0,
    discount_type        varchar(10)   NOT NULL DEFAULT 'FIXED',
    tax_1                numeric(12,2) NOT NULL DEFAULT 0,
    tax_1_type           varchar(10)   NOT NULL DEFAULT 'PERCENTAGE',
    tax_1_percentage     numeric(5,2)  NOT NULL DEFAULT 0,
    tax_2                numeric(12,2) NOT NULL DEFAULT 0,
    tax_2_type           varchar(10)   NOT NULL DEFAULT 'PERCENTAGE',
    tax_2_percentage     numeric(5,2)  NOT NULL DEFAULT 0,
    total_amount         numeric(12,2) NOT NULL DEFAULT 0,
    general_notes        text          NOT NULL,
    is_roadside_assistance boolean     NOT NULL DEFAULT false,
    labor_time_seconds   integer,
    labels               jsonb         NOT NULL DEFAULT '[]'::jsonb,
    custom_fields        jsonb         NOT NULL DEFAULT '{}'::jsonb,
    created_at           timestamptz   NOT NULL,
    updated_at           timestamptz   NOT NULL
);

CREATE TABLE service_entry_line_item (
    id                bigserial PRIMARY KEY,
    service_entry_id  bigint        NOT NULL,
    line_item_type    varchar(20)   NOT NULL,
    description       varchar(255)  NOT NULL,
    service_task_id   bigint,
    part_id           bigint,
    technician_id     bigint,
    tire_id           bigint,
    service_reminder_id bigint,
    unit_cost         numeric(12,2) NOT NULL DEFAULT 0,
    quantity          numeric(10,2) NOT NULL DEFAULT 1,
    parts_cost        numeric(12,2) NOT NULL DEFAULT 0,
    labor_cost        numeric(12,2) NOT NULL DEFAULT 0,
    subtotal          numeric(12,2) NOT NULL DEFAULT 0,
    position          integer       NOT NULL DEFAULT 0,
    created_at        timestamptz   NOT NULL,
    updated_at        timestamptz   NOT NULL
);

-- ---------------------------------------------------------------------------
-- Tires
-- ---------------------------------------------------------------------------

CREATE TABLE axle_template (
    id              bigserial PRIMARY KEY,
    company_id      bigint       NOT NULL,
    name            varchar(100) NOT NULL,
    description     text         NOT NULL,
    total_positions integer      NOT NULL DEFAULT 0,
    CONSTRAINT uq_axle_template UNIQUE (company_id, name)
);

CREATE TABLE axle_definition (
    id                 bigserial PRIMARY KEY,
    template_id        bigint      NOT NULL,
    position_index     integer     NOT NULL,
    label              varchar(50) NOT NULL,
    axle_role          varchar(10) NOT NULL,
    positions_per_side integer     NOT NULL DEFAULT 1,
    CONSTRAINT uq_axle_definition UNIQUE (template_id, position_index)
);

CREATE TABLE wheel_position_definition (
    id      bigserial PRIMARY KEY,
    axle_id bigint      NOT NULL,
    code    varchar(10) NOT NULL,
    side    varchar(1)  NOT NULL,
    slot    integer     NOT NULL DEFAULT 1,
    CONSTRAINT uq_wheel_position_definition UNIQUE (axle_id, code)
);

CREATE TABLE tire_model (
    id                        bigserial PRIMARY KEY,
    company_id                bigint       NOT NULL,
    brand                     varchar(100) NOT NULL,
    model_name                varchar(150) NOT NULL,
    size                      varchar(50)  NOT NULL,
    factory_tread_depth_32nds integer,
    minimum_tread_depth_32nds integer,
    life_expectancy_miles     integer,
    recommended_psi           numeric(5,1),
    CONSTRAINT uq_tire_model UNIQUE (company_id, brand, model_name, size)
);

CREATE TABLE tire (
    id                         bigserial PRIMARY KEY,
    company_id                 bigint       NOT NULL,
    tire_identification_number varchar(100) NOT NULL,
    tire_model_id              bigint,
    status                     varchar(15)  NOT NULL DEFAULT 'IN_STOCK',
    current_tread_depth_32nds  integer,
    current_psi                numeric(5,1),
    total_miles                integer      NOT NULL DEFAULT 0,
    current_vehicle_id         bigint,
    current_position_code      varchar(10)  NOT NULL DEFAULT '',
    purchase_date              date,
    purchase_cost              numeric(10,2),
    vendor_id                  bigint,
    created_at                 timestamptz  NOT NULL,
    CONSTRAINT uq_tire UNIQUE (company_id, tire_identification_number)
);

CREATE TABLE tire_installation (
    id                           bigserial PRIMARY KEY,
    vehicle_id                   bigint      NOT NULL,
    tire_id                      bigint      NOT NULL UNIQUE,
    position_code                varchar(10) NOT NULL,
    install_date                 date        NOT NULL,
    odometer_at_install          integer     NOT NULL DEFAULT 0,
    tread_depth_at_install_32nds integer,
    psi_at_install               numeric(5,1),
    installed_by_id              bigint,
    notes                        text        NOT NULL,
    CONSTRAINT uq_tire_installation UNIQUE (vehicle_id, position_code)
);

CREATE TABLE tire_mount_log (
    id                bigserial PRIMARY KEY,
    tire_id           bigint      NOT NULL,
    vehicle_id        bigint      NOT NULL,
    position_code     varchar(10) NOT NULL,
    event_type        varchar(10) NOT NULL,
    event_date        timestamptz NOT NULL,
    odometer          integer,
    tread_depth_32nds integer,
    psi               numeric(5,1),
    performed_by_id   bigint,
    reason            varchar(255) NOT NULL
);

CREATE TABLE tire_inspection (
    id                bigserial PRIMARY KEY,
    tire_id           bigint      NOT NULL,
    inspection_date   timestamptz NOT NULL,
    odometer          integer,
    tread_depth_32nds integer     NOT NULL,
    psi               numeric(5,1),
    measured_by_id    bigint,
    notes             text        NOT NULL
);

CREATE TABLE vehicle_axle_config (
    vehicle_id   bigint PRIMARY KEY,
    template_id  bigint       NOT NULL,
    display_name varchar(100) NOT NULL
);

CREATE TABLE tire_assignment_request (
    id               bigserial PRIMARY KEY,
    company_id       bigint      NOT NULL,
    tire_id          bigint      NOT NULL,
    vehicle_id       bigint      NOT NULL,
    position_code    varchar(10) NOT NULL,
    state            varchar(10) NOT NULL DEFAULT 'PENDING',
    requested_by_id  bigint,
    requested_at     timestamptz NOT NULL,
    approved_by_id   bigint,
    resolved_at      timestamptz,
    rejection_reason text        NOT NULL,
    notes            text        NOT NULL
);

-- ---------------------------------------------------------------------------
-- Fuel
-- ---------------------------------------------------------------------------

CREATE TABLE fuel_type (
    id         bigserial PRIMARY KEY,
    company_id bigint      NOT NULL,
    name       varchar(50) NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT uq_fuel_type UNIQUE (company_id, name)
);

CREATE TABLE fuel_entry (
    id             bigserial PRIMARY KEY,
    asset_id       bigint        NOT NULL,
    employee_id    bigint        NOT NULL,
    date           timestamptz   NOT NULL,
    fuel_type      varchar(20)   NOT NULL,
    quantity       numeric(10,3) NOT NULL,
    unit_cost      numeric(10,3) NOT NULL,
    total_cost     numeric(12,2) NOT NULL,
    odometer       numeric(15,2) NOT NULL,
    vendor_id      bigint        NOT NULL,
    full_tank      boolean       NOT NULL DEFAULT true,
    miles_traveled numeric(10,2),
    fuel_efficiency numeric(6,2),
    state          varchar(50)   NOT NULL,
    reference      varchar(100)  NOT NULL,
    personal       boolean       NOT NULL DEFAULT false,
    reset          boolean       NOT NULL DEFAULT false,
    latitude       numeric(10,7),
    longitude      numeric(10,7),
    external_id    varchar(100)  NOT NULL,
    updated_at     timestamptz   NOT NULL,
    no_semana      varchar(50),
    estado_prov    varchar(100),
    operator_name  varchar(200)
);

CREATE TABLE fuel_comment (
    id         bigserial PRIMARY KEY,
    entry_id   bigint      NOT NULL,
    user_id    bigint      NOT NULL,   -- auth_user (no FK)
    text       text        NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE fuel_photo (
    id             bigserial PRIMARY KEY,
    entry_id       bigint       NOT NULL,
    uploaded_by_id bigint       NOT NULL,  -- auth_user (no FK)
    file           varchar(100) NOT NULL,
    file_name      varchar(255) NOT NULL,
    file_size      bigint       NOT NULL,
    mime_type      varchar(100) NOT NULL,
    description    text,
    uploaded_at    timestamptz  NOT NULL,
    is_primary     boolean      NOT NULL DEFAULT false
);

-- ---------------------------------------------------------------------------
-- Inspections
-- ---------------------------------------------------------------------------

CREATE TABLE inspection_form (
    id                 bigserial PRIMARY KEY,
    company_id         bigint       NOT NULL,
    title              varchar(255) NOT NULL,
    description        text         NOT NULL,
    version            integer      NOT NULL DEFAULT 1,
    require_live_photo boolean      NOT NULL DEFAULT false,
    auto_create_issues boolean      NOT NULL DEFAULT true,
    color              varchar(7)   NOT NULL DEFAULT '#3B82F6',
    archived_at        timestamptz,
    created_at         timestamptz  NOT NULL,
    updated_at         timestamptz  NOT NULL
);

CREATE TABLE inspection_form_item (
    id                                     bigserial PRIMARY KEY,
    form_id                                bigint       NOT NULL,
    item_type                              varchar(20)  NOT NULL,
    label                                  varchar(255) NOT NULL,
    short_description                      varchar(500) NOT NULL,
    instructions                           text         NOT NULL,
    position                               integer      NOT NULL,
    is_required                            boolean      NOT NULL DEFAULT true,
    pass_label                             varchar(50)  NOT NULL DEFAULT 'Pass',
    fail_label                             varchar(50)  NOT NULL DEFAULT 'Fail',
    na_label                               varchar(50)  NOT NULL DEFAULT 'N/A',
    enable_na_option                       boolean      NOT NULL DEFAULT false,
    require_remark_on_fail                 boolean      NOT NULL DEFAULT true,
    require_remark_on_pass                 boolean      NOT NULL DEFAULT false,
    require_photo_on_fail                  boolean      NOT NULL DEFAULT false,
    require_meter_entry_photo_verification boolean      NOT NULL DEFAULT false,
    require_secondary_meter_if_one_exists  boolean      NOT NULL DEFAULT false,
    type_config                            jsonb        NOT NULL DEFAULT '{}'::jsonb,
    created_at                             timestamptz  NOT NULL,
    updated_at                             timestamptz  NOT NULL,
    CONSTRAINT uq_inspection_form_item UNIQUE (form_id, position)
);

CREATE TABLE inspection_submission (
    id                 bigserial PRIMARY KEY,
    company_id         bigint      NOT NULL,
    form_id            bigint      NOT NULL,
    asset_id           bigint      NOT NULL,
    submitted_by_id    bigint      NOT NULL,
    started_at         timestamptz NOT NULL,
    submitted_at       timestamptz NOT NULL,
    duration_seconds   integer,
    starting_latitude  numeric(10,7),
    starting_longitude numeric(10,7),
    submitted_latitude numeric(10,7),
    submitted_longitude numeric(10,7),
    signature          varchar(100),
    odometer           numeric(15,2),
    total_items        integer     NOT NULL DEFAULT 0,
    failed_items_count integer     NOT NULL DEFAULT 0,
    passed_items_count integer     NOT NULL DEFAULT 0,
    comments_count     integer     NOT NULL DEFAULT 0,
    images_count       integer     NOT NULL DEFAULT 0,
    general_notes      text        NOT NULL,
    created_at         timestamptz NOT NULL
);

CREATE TABLE inspection_submission_item (
    id                 bigserial PRIMARY KEY,
    submission_id      bigint      NOT NULL,
    form_item_id       bigint      NOT NULL,
    result_status      varchar(10) NOT NULL,
    result_value       jsonb       NOT NULL DEFAULT '{}'::jsonb,
    remark             text        NOT NULL,
    photo              varchar(100),
    latitude           numeric(10,7),
    longitude          numeric(10,7),
    generated_issue_id bigint,
    CONSTRAINT uq_inspection_submission_item UNIQUE (submission_id, form_item_id)
);

-- ---------------------------------------------------------------------------
-- Media / comments / documents / misc
-- ---------------------------------------------------------------------------

CREATE TABLE media (
    id             bigserial PRIMARY KEY,
    company_id     bigint       NOT NULL,
    asset_id       bigint       NOT NULL,
    file           varchar(100) NOT NULL,
    title          varchar(255) NOT NULL,
    description    text         NOT NULL,
    file_type      varchar(20)  NOT NULL,
    file_size      integer      NOT NULL,
    uploaded_by_id bigint,                 -- auth_user (no FK)
    created_at     timestamptz  NOT NULL,
    updated_at     timestamptz  NOT NULL
);

CREATE TABLE comment (
    id              bigserial PRIMARY KEY,
    company_id      bigint      NOT NULL,
    content_type_id bigint      NOT NULL,  -- django_content_type (no FK)
    object_id       integer     NOT NULL,
    body            text        NOT NULL,
    author_id       bigint,
    created_at      timestamptz NOT NULL,
    updated_at      timestamptz NOT NULL
);

CREATE TABLE warranty (
    id          bigserial PRIMARY KEY,
    company_id  bigint  NOT NULL,
    provider_id bigint  NOT NULL,
    asset_id    bigint,
    part_id     bigint,
    start_date  date    NOT NULL,
    end_date    date    NOT NULL,
    terms       text    NOT NULL,
    is_active   boolean NOT NULL DEFAULT true
);

CREATE TABLE weekly_mileage_goal (
    id                  bigserial PRIMARY KEY,
    company_id          bigint        NOT NULL,
    service_type        varchar(100)  NOT NULL,
    rate_per_mile       numeric(10,2) NOT NULL,
    weekly_mileage_goal integer       NOT NULL,
    mpg_goal            numeric(5,2),
    units_per_service   integer       NOT NULL DEFAULT 1,
    machines_in_workshop integer      NOT NULL DEFAULT 0,
    missing_miles       integer       NOT NULL DEFAULT 0,
    sort_order          integer       NOT NULL DEFAULT 0,
    is_active           boolean       NOT NULL DEFAULT true,
    created_at          timestamptz   NOT NULL,
    updated_at          timestamptz   NOT NULL,
    CONSTRAINT uq_weekly_mileage_goal UNIQUE (company_id, service_type)
);

-- ---------------------------------------------------------------------------
-- Many-to-many join tables
-- ---------------------------------------------------------------------------

CREATE TABLE employee_companies (
    id          bigserial PRIMARY KEY,
    employee_id bigint NOT NULL,
    company_id  bigint NOT NULL,
    CONSTRAINT uq_employee_companies UNIQUE (employee_id, company_id)
);

CREATE TABLE work_order_issues (
    id          bigserial PRIMARY KEY,
    work_order_id bigint NOT NULL,
    issue_id    bigint NOT NULL,
    CONSTRAINT uq_work_order_issues UNIQUE (work_order_id, issue_id)
);

CREATE TABLE work_order_faults (
    id          bigserial PRIMARY KEY,
    work_order_id bigint NOT NULL,
    fault_id    bigint NOT NULL,
    CONSTRAINT uq_work_order_faults UNIQUE (work_order_id, fault_id)
);

CREATE TABLE work_order_line_item_issues (
    id                  bigserial PRIMARY KEY,
    work_order_line_item_id bigint NOT NULL,
    issue_id            bigint NOT NULL,
    CONSTRAINT uq_work_order_line_item_issues UNIQUE (work_order_line_item_id, issue_id)
);

CREATE TABLE issue_assigned_to (
    id          bigserial PRIMARY KEY,
    issue_id    bigint NOT NULL,
    employee_id bigint NOT NULL,
    CONSTRAINT uq_issue_assigned_to UNIQUE (issue_id, employee_id)
);

CREATE TABLE issue_watchers (
    id          bigserial PRIMARY KEY,
    issue_id    bigint NOT NULL,
    employee_id bigint NOT NULL,
    CONSTRAINT uq_issue_watchers UNIQUE (issue_id, employee_id)
);

CREATE TABLE service_entry_line_item_issues (
    id                     bigserial PRIMARY KEY,
    service_entry_line_item_id bigint NOT NULL,
    issue_id               bigint NOT NULL,
    CONSTRAINT uq_service_entry_line_item_issues UNIQUE (service_entry_line_item_id, issue_id)
);

-- ---------------------------------------------------------------------------
-- Foreign keys (declared after all tables so creation order is irrelevant)
-- ---------------------------------------------------------------------------

ALTER TABLE role ADD CONSTRAINT fk_role_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE "group" ADD CONSTRAINT fk_group_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE "group" ADD CONSTRAINT fk_group_parent FOREIGN KEY (parent_id) REFERENCES "group"(id)            ON DELETE SET NULL;

ALTER TABLE employee ADD CONSTRAINT fk_employee_default_company FOREIGN KEY (default_company_id) REFERENCES company(id)          ON DELETE SET NULL;
ALTER TABLE employee ADD CONSTRAINT fk_employee_role FOREIGN KEY (role_id) REFERENCES role(id)             ON DELETE RESTRICT;
ALTER TABLE employee ADD CONSTRAINT fk_employee_group FOREIGN KEY (group_id) REFERENCES "group"(id)            ON DELETE SET NULL;

ALTER TABLE asset_type ADD CONSTRAINT fk_asset_type_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE asset_status ADD CONSTRAINT fk_asset_status_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE catalog_option ADD CONSTRAINT fk_catalog_option_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE asset ADD CONSTRAINT fk_asset_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE asset ADD CONSTRAINT fk_asset_asset_type FOREIGN KEY (asset_type_id) REFERENCES asset_type(id)        ON DELETE SET NULL;
ALTER TABLE asset ADD CONSTRAINT fk_asset_status FOREIGN KEY (status_id) REFERENCES asset_status(id)      ON DELETE RESTRICT;
ALTER TABLE asset ADD CONSTRAINT fk_asset_lease_vendor FOREIGN KEY (lease_vendor_id) REFERENCES vendor(id)           ON DELETE SET NULL;
ALTER TABLE asset ADD CONSTRAINT fk_asset_owner_company FOREIGN KEY (owner_company_id) REFERENCES company(id)          ON DELETE SET NULL;
ALTER TABLE asset ADD CONSTRAINT fk_asset_loan_vendor FOREIGN KEY (loan_vendor_id) REFERENCES vendor(id)           ON DELETE SET NULL;

ALTER TABLE vehicle ADD CONSTRAINT fk_vehicle_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE trailer ADD CONSTRAINT fk_trailer_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;

ALTER TABLE asset_trailer_assignment ADD CONSTRAINT fk_asset_trailer_assignment_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE asset_trailer_assignment ADD CONSTRAINT fk_asset_trailer_assignment_trailer FOREIGN KEY (trailer_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE asset_trailer_assignment ADD CONSTRAINT fk_asset_trailer_assignment_assigned_by FOREIGN KEY (assigned_by_id) REFERENCES employee(id)         ON DELETE SET NULL;

ALTER TABLE vehicle_make ADD CONSTRAINT fk_vehicle_make_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE vehicle_model ADD CONSTRAINT fk_vehicle_model_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE vehicle_model ADD CONSTRAINT fk_vehicle_model_make FOREIGN KEY (make_id) REFERENCES vehicle_make(id)      ON DELETE CASCADE;

ALTER TABLE part_category ADD CONSTRAINT fk_part_category_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE part_manufacturer ADD CONSTRAINT fk_part_manufacturer_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE measurement_unit ADD CONSTRAINT fk_measurement_unit_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE part ADD CONSTRAINT fk_part_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE part ADD CONSTRAINT fk_part_part_category FOREIGN KEY (part_category_id) REFERENCES part_category(id)     ON DELETE SET NULL;
ALTER TABLE part ADD CONSTRAINT fk_part_part_manufacturer FOREIGN KEY (part_manufacturer_id) REFERENCES part_manufacturer(id) ON DELETE SET NULL;
ALTER TABLE part ADD CONSTRAINT fk_part_measurement_unit FOREIGN KEY (measurement_unit_id) REFERENCES measurement_unit(id)  ON DELETE SET NULL;

ALTER TABLE location ADD CONSTRAINT fk_location_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE part_location ADD CONSTRAINT fk_part_location_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE part_location ADD CONSTRAINT fk_part_location_location FOREIGN KEY (location_id) REFERENCES location(id)         ON DELETE SET NULL;

ALTER TABLE part_inventory ADD CONSTRAINT fk_part_inventory_part FOREIGN KEY (part_id) REFERENCES part(id)             ON DELETE CASCADE;
ALTER TABLE part_inventory ADD CONSTRAINT fk_part_inventory_location FOREIGN KEY (location_id) REFERENCES part_location(id)     ON DELETE CASCADE;

ALTER TABLE inventory_adjustment_reason ADD CONSTRAINT fk_inventory_adjustment_reason_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE inventory_journal_entry ADD CONSTRAINT fk_inventory_journal_entry_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE inventory_journal_entry ADD CONSTRAINT fk_inventory_journal_entry_part FOREIGN KEY (part_id) REFERENCES part(id)             ON DELETE CASCADE;
ALTER TABLE inventory_journal_entry ADD CONSTRAINT fk_inventory_journal_entry_part_location_detail FOREIGN KEY (part_location_detail_id) REFERENCES part_inventory(id)    ON DELETE CASCADE;
ALTER TABLE inventory_journal_entry ADD CONSTRAINT fk_inventory_journal_entry_user FOREIGN KEY (user_id) REFERENCES employee(id)         ON DELETE SET NULL;
ALTER TABLE inventory_journal_entry ADD CONSTRAINT fk_inventory_journal_entry_reason FOREIGN KEY (reason_id) REFERENCES inventory_adjustment_reason(id) ON DELETE SET NULL;
ALTER TABLE inventory_journal_entry ADD CONSTRAINT fk_inventory_journal_entry_work_order FOREIGN KEY (work_order_id) REFERENCES work_order(id)        ON DELETE SET NULL;
ALTER TABLE inventory_journal_entry ADD CONSTRAINT fk_inventory_journal_entry_purchase_order_line FOREIGN KEY (purchase_order_line_id) REFERENCES purchase_order_line_item(id) ON DELETE SET NULL;
ALTER TABLE inventory_journal_entry ADD CONSTRAINT fk_inventory_journal_entry_vendor FOREIGN KEY (vendor_id) REFERENCES vendor(id)           ON DELETE SET NULL;
ALTER TABLE inventory_journal_entry ADD CONSTRAINT fk_inventory_journal_entry_transfer_part_location FOREIGN KEY (transfer_part_location_id) REFERENCES part_location(id)   ON DELETE SET NULL;

ALTER TABLE vendor ADD CONSTRAINT fk_vendor_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE work_order_status ADD CONSTRAINT fk_work_order_status_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE work_order ADD CONSTRAINT fk_work_order_location FOREIGN KEY (location_id) REFERENCES location(id)         ON DELETE SET NULL;
ALTER TABLE work_order ADD CONSTRAINT fk_work_order_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE work_order ADD CONSTRAINT fk_work_order_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE work_order ADD CONSTRAINT fk_work_order_status FOREIGN KEY (status_id) REFERENCES work_order_status(id)  ON DELETE RESTRICT;
ALTER TABLE work_order ADD CONSTRAINT fk_work_order_vendor FOREIGN KEY (vendor_id) REFERENCES vendor(id)           ON DELETE SET NULL;
ALTER TABLE work_order ADD CONSTRAINT fk_work_order_assigned_to FOREIGN KEY (assigned_to_id) REFERENCES employee(id)         ON DELETE SET NULL;
ALTER TABLE work_order ADD CONSTRAINT fk_work_order_issued_by FOREIGN KEY (issued_by_id) REFERENCES employee(id)         ON DELETE SET NULL;
ALTER TABLE work_order ADD CONSTRAINT fk_work_order_fault FOREIGN KEY (fault_id) REFERENCES fault(id)            ON DELETE SET NULL;

ALTER TABLE work_order_line_item ADD CONSTRAINT fk_work_order_line_item_work_order FOREIGN KEY (work_order_id) REFERENCES work_order(id)        ON DELETE CASCADE;

ALTER TABLE work_order_sub_line_item ADD CONSTRAINT fk_work_order_sub_line_item_line_item FOREIGN KEY (line_item_id) REFERENCES work_order_line_item(id) ON DELETE CASCADE;
ALTER TABLE work_order_sub_line_item ADD CONSTRAINT fk_work_order_sub_line_item_part FOREIGN KEY (part_id) REFERENCES part(id)             ON DELETE SET NULL;
ALTER TABLE work_order_sub_line_item ADD CONSTRAINT fk_work_order_sub_line_item_part_location_detail FOREIGN KEY (part_location_detail_id) REFERENCES part_inventory(id)    ON DELETE SET NULL;
ALTER TABLE work_order_sub_line_item ADD CONSTRAINT fk_work_order_sub_line_item_technician FOREIGN KEY (technician_id) REFERENCES employee(id)         ON DELETE SET NULL;

ALTER TABLE labor_time_entry ADD CONSTRAINT fk_labor_time_entry_sub_line_item FOREIGN KEY (sub_line_item_id) REFERENCES work_order_sub_line_item(id) ON DELETE CASCADE;
ALTER TABLE labor_time_entry ADD CONSTRAINT fk_labor_time_entry_technician FOREIGN KEY (technician_id) REFERENCES employee(id)         ON DELETE RESTRICT;

ALTER TABLE work_order_status_log ADD CONSTRAINT fk_work_order_status_log_work_order FOREIGN KEY (work_order_id) REFERENCES work_order(id)        ON DELETE CASCADE;
ALTER TABLE work_order_status_log ADD CONSTRAINT fk_work_order_status_log_status FOREIGN KEY (status_id) REFERENCES work_order_status(id)  ON DELETE CASCADE;

ALTER TABLE issue_priority ADD CONSTRAINT fk_issue_priority_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE fault ADD CONSTRAINT fk_fault_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE issue ADD CONSTRAINT fk_issue_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE issue ADD CONSTRAINT fk_issue_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE issue ADD CONSTRAINT fk_issue_priority FOREIGN KEY (priority_id) REFERENCES issue_priority(id)    ON DELETE SET NULL;
ALTER TABLE issue ADD CONSTRAINT fk_issue_fault FOREIGN KEY (fault_id) REFERENCES fault(id)            ON DELETE SET NULL;
ALTER TABLE issue ADD CONSTRAINT fk_issue_inspection_submission FOREIGN KEY (inspection_submission_id) REFERENCES inspection_submission(id) ON DELETE SET NULL;
ALTER TABLE issue ADD CONSTRAINT fk_issue_reported_by FOREIGN KEY (reported_by_id) REFERENCES employee(id)         ON DELETE RESTRICT;
ALTER TABLE issue ADD CONSTRAINT fk_issue_resolved_by FOREIGN KEY (resolved_by_id) REFERENCES employee(id)         ON DELETE SET NULL;
ALTER TABLE issue ADD CONSTRAINT fk_issue_reopened_by FOREIGN KEY (reopened_by_id) REFERENCES employee(id)         ON DELETE SET NULL;
ALTER TABLE issue ADD CONSTRAINT fk_issue_closed_by FOREIGN KEY (closed_by_id) REFERENCES employee(id)         ON DELETE SET NULL;

ALTER TABLE purchase_order ADD CONSTRAINT fk_purchase_order_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE purchase_order ADD CONSTRAINT fk_purchase_order_vendor FOREIGN KEY (vendor_id) REFERENCES vendor(id)           ON DELETE RESTRICT;
ALTER TABLE purchase_order ADD CONSTRAINT fk_purchase_order_destination FOREIGN KEY (destination_id) REFERENCES part_location(id)     ON DELETE RESTRICT;
ALTER TABLE purchase_order ADD CONSTRAINT fk_purchase_order_created_by FOREIGN KEY (created_by_id) REFERENCES employee(id)         ON DELETE SET NULL;
ALTER TABLE purchase_order ADD CONSTRAINT fk_purchase_order_submitted_by FOREIGN KEY (submitted_by_id) REFERENCES employee(id)         ON DELETE SET NULL;
ALTER TABLE purchase_order ADD CONSTRAINT fk_purchase_order_rejected_by FOREIGN KEY (rejected_by_id) REFERENCES employee(id)         ON DELETE SET NULL;
ALTER TABLE purchase_order ADD CONSTRAINT fk_purchase_order_approved_by FOREIGN KEY (approved_by_id) REFERENCES employee(id)         ON DELETE SET NULL;

ALTER TABLE purchase_order_line_item ADD CONSTRAINT fk_purchase_order_line_item_purchase_order FOREIGN KEY (purchase_order_id) REFERENCES purchase_order(id)    ON DELETE CASCADE;
ALTER TABLE purchase_order_line_item ADD CONSTRAINT fk_purchase_order_line_item_part FOREIGN KEY (part_id) REFERENCES part(id)             ON DELETE RESTRICT;

ALTER TABLE service_task ADD CONSTRAINT fk_service_task_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE service_task ADD CONSTRAINT fk_service_task_parent_task FOREIGN KEY (parent_task_id) REFERENCES service_task(id)      ON DELETE CASCADE;

ALTER TABLE service_task_part ADD CONSTRAINT fk_service_task_part_service_task FOREIGN KEY (service_task_id) REFERENCES service_task(id)      ON DELETE CASCADE;
ALTER TABLE service_task_part ADD CONSTRAINT fk_service_task_part_part FOREIGN KEY (part_id) REFERENCES part(id)             ON DELETE CASCADE;

ALTER TABLE service_reminder ADD CONSTRAINT fk_service_reminder_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE service_reminder ADD CONSTRAINT fk_service_reminder_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE service_reminder ADD CONSTRAINT fk_service_reminder_service_task FOREIGN KEY (service_task_id) REFERENCES service_task(id)      ON DELETE SET NULL;
ALTER TABLE service_reminder ADD CONSTRAINT fk_service_reminder_last_service_entry FOREIGN KEY (last_service_entry_id) REFERENCES service_entry(id)     ON DELETE SET NULL;

ALTER TABLE service_entry ADD CONSTRAINT fk_service_entry_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE service_entry ADD CONSTRAINT fk_service_entry_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE service_entry ADD CONSTRAINT fk_service_entry_vendor FOREIGN KEY (vendor_id) REFERENCES vendor(id)           ON DELETE SET NULL;
ALTER TABLE service_entry ADD CONSTRAINT fk_service_entry_work_order FOREIGN KEY (work_order_id) REFERENCES work_order(id)        ON DELETE SET NULL;

ALTER TABLE service_entry_line_item ADD CONSTRAINT fk_service_entry_line_item_service_entry FOREIGN KEY (service_entry_id) REFERENCES service_entry(id)     ON DELETE CASCADE;
ALTER TABLE service_entry_line_item ADD CONSTRAINT fk_service_entry_line_item_service_task FOREIGN KEY (service_task_id) REFERENCES service_task(id)      ON DELETE SET NULL;
ALTER TABLE service_entry_line_item ADD CONSTRAINT fk_service_entry_line_item_part FOREIGN KEY (part_id) REFERENCES part(id)             ON DELETE SET NULL;
ALTER TABLE service_entry_line_item ADD CONSTRAINT fk_service_entry_line_item_technician FOREIGN KEY (technician_id) REFERENCES employee(id)         ON DELETE SET NULL;
ALTER TABLE service_entry_line_item ADD CONSTRAINT fk_service_entry_line_item_tire FOREIGN KEY (tire_id) REFERENCES tire(id)             ON DELETE SET NULL;
ALTER TABLE service_entry_line_item ADD CONSTRAINT fk_service_entry_line_item_service_reminder FOREIGN KEY (service_reminder_id) REFERENCES service_reminder(id)  ON DELETE SET NULL;

ALTER TABLE axle_template ADD CONSTRAINT fk_axle_template_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE axle_definition ADD CONSTRAINT fk_axle_definition_template FOREIGN KEY (template_id) REFERENCES axle_template(id)     ON DELETE CASCADE;
ALTER TABLE wheel_position_definition ADD CONSTRAINT fk_wheel_position_definition_axle FOREIGN KEY (axle_id) REFERENCES axle_definition(id)   ON DELETE CASCADE;

ALTER TABLE tire_model ADD CONSTRAINT fk_tire_model_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE tire ADD CONSTRAINT fk_tire_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE tire ADD CONSTRAINT fk_tire_tire_model FOREIGN KEY (tire_model_id) REFERENCES tire_model(id)        ON DELETE RESTRICT;
ALTER TABLE tire ADD CONSTRAINT fk_tire_current_vehicle FOREIGN KEY (current_vehicle_id) REFERENCES asset(id)            ON DELETE SET NULL;
ALTER TABLE tire ADD CONSTRAINT fk_tire_vendor FOREIGN KEY (vendor_id) REFERENCES vendor(id)           ON DELETE SET NULL;

ALTER TABLE tire_installation ADD CONSTRAINT fk_tire_installation_vehicle FOREIGN KEY (vehicle_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE tire_installation ADD CONSTRAINT fk_tire_installation_tire FOREIGN KEY (tire_id) REFERENCES tire(id)             ON DELETE CASCADE;
ALTER TABLE tire_installation ADD CONSTRAINT fk_tire_installation_installed_by FOREIGN KEY (installed_by_id) REFERENCES employee(id)         ON DELETE SET NULL;

ALTER TABLE tire_mount_log ADD CONSTRAINT fk_tire_mount_log_tire FOREIGN KEY (tire_id) REFERENCES tire(id)             ON DELETE CASCADE;
ALTER TABLE tire_mount_log ADD CONSTRAINT fk_tire_mount_log_vehicle FOREIGN KEY (vehicle_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE tire_mount_log ADD CONSTRAINT fk_tire_mount_log_performed_by FOREIGN KEY (performed_by_id) REFERENCES employee(id)         ON DELETE SET NULL;

ALTER TABLE tire_inspection ADD CONSTRAINT fk_tire_inspection_tire FOREIGN KEY (tire_id) REFERENCES tire(id)             ON DELETE CASCADE;
ALTER TABLE tire_inspection ADD CONSTRAINT fk_tire_inspection_measured_by FOREIGN KEY (measured_by_id) REFERENCES employee(id)         ON DELETE SET NULL;

ALTER TABLE vehicle_axle_config ADD CONSTRAINT fk_vehicle_axle_config_vehicle FOREIGN KEY (vehicle_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE vehicle_axle_config ADD CONSTRAINT fk_vehicle_axle_config_template FOREIGN KEY (template_id) REFERENCES axle_template(id)     ON DELETE RESTRICT;

ALTER TABLE tire_assignment_request ADD CONSTRAINT fk_tire_assignment_request_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE tire_assignment_request ADD CONSTRAINT fk_tire_assignment_request_tire FOREIGN KEY (tire_id) REFERENCES tire(id)             ON DELETE CASCADE;
ALTER TABLE tire_assignment_request ADD CONSTRAINT fk_tire_assignment_request_vehicle FOREIGN KEY (vehicle_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE tire_assignment_request ADD CONSTRAINT fk_tire_assignment_request_requested_by FOREIGN KEY (requested_by_id) REFERENCES employee(id)         ON DELETE SET NULL;
ALTER TABLE tire_assignment_request ADD CONSTRAINT fk_tire_assignment_request_approved_by FOREIGN KEY (approved_by_id) REFERENCES employee(id)         ON DELETE SET NULL;

ALTER TABLE fuel_type ADD CONSTRAINT fk_fuel_type_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE fuel_entry ADD CONSTRAINT fk_fuel_entry_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE fuel_entry ADD CONSTRAINT fk_fuel_entry_employee FOREIGN KEY (employee_id) REFERENCES employee(id)         ON DELETE RESTRICT;
ALTER TABLE fuel_entry ADD CONSTRAINT fk_fuel_entry_vendor FOREIGN KEY (vendor_id) REFERENCES vendor(id)           ON DELETE RESTRICT;

ALTER TABLE fuel_comment ADD CONSTRAINT fk_fuel_comment_entry FOREIGN KEY (entry_id) REFERENCES fuel_entry(id)        ON DELETE CASCADE;
ALTER TABLE fuel_photo ADD CONSTRAINT fk_fuel_photo_entry FOREIGN KEY (entry_id) REFERENCES fuel_entry(id)        ON DELETE CASCADE;

ALTER TABLE inspection_form ADD CONSTRAINT fk_inspection_form_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE inspection_form_item ADD CONSTRAINT fk_inspection_form_item_form FOREIGN KEY (form_id) REFERENCES inspection_form(id)   ON DELETE CASCADE;

ALTER TABLE inspection_submission ADD CONSTRAINT fk_inspection_submission_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE inspection_submission ADD CONSTRAINT fk_inspection_submission_form FOREIGN KEY (form_id) REFERENCES inspection_form(id)   ON DELETE RESTRICT;
ALTER TABLE inspection_submission ADD CONSTRAINT fk_inspection_submission_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE inspection_submission ADD CONSTRAINT fk_inspection_submission_submitted_by FOREIGN KEY (submitted_by_id) REFERENCES employee(id)         ON DELETE RESTRICT;

ALTER TABLE inspection_submission_item ADD CONSTRAINT fk_inspection_submission_item_submission FOREIGN KEY (submission_id) REFERENCES inspection_submission(id) ON DELETE CASCADE;
ALTER TABLE inspection_submission_item ADD CONSTRAINT fk_inspection_submission_item_form_item FOREIGN KEY (form_item_id) REFERENCES inspection_form_item(id) ON DELETE RESTRICT;
ALTER TABLE inspection_submission_item ADD CONSTRAINT fk_inspection_submission_item_generated_issue FOREIGN KEY (generated_issue_id) REFERENCES issue(id)            ON DELETE SET NULL;

ALTER TABLE media ADD CONSTRAINT fk_media_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE media ADD CONSTRAINT fk_media_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;

ALTER TABLE comment ADD CONSTRAINT fk_comment_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE comment ADD CONSTRAINT fk_comment_author FOREIGN KEY (author_id) REFERENCES employee(id)         ON DELETE SET NULL;

ALTER TABLE warranty ADD CONSTRAINT fk_warranty_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;
ALTER TABLE warranty ADD CONSTRAINT fk_warranty_provider FOREIGN KEY (provider_id) REFERENCES vendor(id)           ON DELETE CASCADE;
ALTER TABLE warranty ADD CONSTRAINT fk_warranty_asset FOREIGN KEY (asset_id) REFERENCES asset(id)            ON DELETE CASCADE;
ALTER TABLE warranty ADD CONSTRAINT fk_warranty_part FOREIGN KEY (part_id) REFERENCES part(id)             ON DELETE CASCADE;

ALTER TABLE weekly_mileage_goal ADD CONSTRAINT fk_weekly_mileage_goal_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE employee_companies ADD CONSTRAINT fk_employee_companies_employee FOREIGN KEY (employee_id) REFERENCES employee(id)         ON DELETE CASCADE;
ALTER TABLE employee_companies ADD CONSTRAINT fk_employee_companies_company FOREIGN KEY (company_id) REFERENCES company(id)          ON DELETE CASCADE;

ALTER TABLE work_order_issues ADD CONSTRAINT fk_work_order_issues_work_order FOREIGN KEY (work_order_id) REFERENCES work_order(id)        ON DELETE CASCADE;
ALTER TABLE work_order_issues ADD CONSTRAINT fk_work_order_issues_issue FOREIGN KEY (issue_id) REFERENCES issue(id)            ON DELETE CASCADE;

ALTER TABLE work_order_faults ADD CONSTRAINT fk_work_order_faults_work_order FOREIGN KEY (work_order_id) REFERENCES work_order(id)        ON DELETE CASCADE;
ALTER TABLE work_order_faults ADD CONSTRAINT fk_work_order_faults_fault FOREIGN KEY (fault_id) REFERENCES fault(id)            ON DELETE CASCADE;

ALTER TABLE work_order_line_item_issues ADD CONSTRAINT fk_work_order_line_item_issues_work_order_line_item FOREIGN KEY (work_order_line_item_id) REFERENCES work_order_line_item(id) ON DELETE CASCADE;
ALTER TABLE work_order_line_item_issues ADD CONSTRAINT fk_work_order_line_item_issues_issue FOREIGN KEY (issue_id) REFERENCES issue(id)            ON DELETE CASCADE;

ALTER TABLE issue_assigned_to ADD CONSTRAINT fk_issue_assigned_to_issue FOREIGN KEY (issue_id) REFERENCES issue(id)            ON DELETE CASCADE;
ALTER TABLE issue_assigned_to ADD CONSTRAINT fk_issue_assigned_to_employee FOREIGN KEY (employee_id) REFERENCES employee(id)         ON DELETE CASCADE;

ALTER TABLE issue_watchers ADD CONSTRAINT fk_issue_watchers_issue FOREIGN KEY (issue_id) REFERENCES issue(id)            ON DELETE CASCADE;
ALTER TABLE issue_watchers ADD CONSTRAINT fk_issue_watchers_employee FOREIGN KEY (employee_id) REFERENCES employee(id)         ON DELETE CASCADE;

ALTER TABLE service_entry_line_item_issues ADD CONSTRAINT fk_service_entry_line_item_issues_service_entry_line_item FOREIGN KEY (service_entry_line_item_id) REFERENCES service_entry_line_item(id) ON DELETE CASCADE;
ALTER TABLE service_entry_line_item_issues ADD CONSTRAINT fk_service_entry_line_item_issues_issue FOREIGN KEY (issue_id) REFERENCES issue(id)            ON DELETE CASCADE;

-- ---------------------------------------------------------------------------
-- Indexes (btree unless noted) and partial unique indexes
-- ---------------------------------------------------------------------------

-- AssetTrailerAssignment: partial uniques (only while active)
CREATE UNIQUE INDEX uq_asset_trailer_assignment_asset_position_active ON asset_trailer_assignment (asset_id, position) WHERE is_active = true;
CREATE UNIQUE INDEX uq_asset_trailer_assignment_trailer_active ON asset_trailer_assignment (trailer_id)          WHERE is_active = true;

-- Group
CREATE INDEX idx_group_parent ON "group" (parent_id);

-- Fault
CREATE INDEX idx_fault_company_family ON fault (company_id, family);
CREATE INDEX idx_fault_applies_to_asset_types_gin ON fault USING gin (applies_to_asset_types);

-- Issue
CREATE INDEX idx_issue_company_state ON issue (company_id, state);
CREATE INDEX idx_issue_asset_created_at ON issue (asset_id, created_at DESC);
CREATE INDEX idx_issue_company_reported_at ON issue (company_id, reported_at DESC);

-- WorkOrder
CREATE INDEX idx_work_order_company_issued_at ON work_order (company_id, issued_at DESC);
CREATE INDEX idx_work_order_asset_issued_at ON work_order (asset_id, issued_at DESC);
CREATE INDEX idx_work_order_vendor_issued_at ON work_order (vendor_id, issued_at DESC);

-- ServiceEntry
CREATE INDEX idx_service_entry_company_completed_at ON service_entry (company_id, completed_at DESC);
CREATE INDEX idx_service_entry_asset_completed_at ON service_entry (asset_id, completed_at DESC);

-- ServiceReminder
CREATE INDEX idx_service_reminder_company_status ON service_reminder (company_id, status);
CREATE INDEX idx_service_reminder_asset_status ON service_reminder (asset_id, status);

-- InspectionSubmission
CREATE INDEX idx_inspection_submission_asset_submitted_at ON inspection_submission (asset_id, submitted_at DESC);
CREATE INDEX idx_inspection_submission_submitted_by_submitted_at ON inspection_submission (submitted_by_id, submitted_at DESC);

-- PurchaseOrder
CREATE INDEX idx_purchase_order_company_state ON purchase_order (company_id, state);
CREATE INDEX idx_purchase_order_vendor_created_at ON purchase_order (vendor_id, created_at DESC);

-- InventoryJournalEntry
CREATE INDEX idx_inventory_journal_entry_part_created_at ON inventory_journal_entry (part_id, created_at DESC);
CREATE INDEX idx_inventory_journal_entry_part_location_detail_created_at ON inventory_journal_entry (part_location_detail_id, created_at DESC);

-- TireMountLog
CREATE INDEX idx_tire_mount_log_tire_event_date ON tire_mount_log (tire_id, event_date DESC);
CREATE INDEX idx_tire_mount_log_vehicle_position_code_event_date ON tire_mount_log (vehicle_id, position_code, event_date DESC);

-- Media
CREATE INDEX idx_media_company_created_at ON media (company_id, created_at DESC);
CREATE INDEX idx_media_asset_created_at ON media (asset_id, created_at DESC);

-- Comment
CREATE INDEX idx_comment_content_type_object ON comment (content_type_id, object_id);
