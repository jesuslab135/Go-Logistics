-- 000017_custom_field_definitions.up.sql
-- custom_fields is a jsonb column on eight resources, and until now nothing
-- described what could go in it. Every client wrote whatever keys it liked, no
-- value was ever checked, and the only way to build a form over it was to guess
-- - or to ship a raw JSON editor, which the original audit explicitly advised
-- against putting in front of a business user.
--
-- A definition is per company and per resource. A field meaningful on an asset
-- is rarely meaningful on a purchase order, and one global set would put every
-- field on every form.
CREATE TABLE custom_field_definition (
    id         bigserial    PRIMARY KEY,
    company_id bigint       NOT NULL,
    -- resource is the API's own plural name: "assets", "parts", "work-orders".
    resource   varchar(50)  NOT NULL,
    -- key is what appears in the jsonb document; label is what a person reads.
    key        varchar(50)  NOT NULL,
    label      varchar(100) NOT NULL,
    -- field_type is one of text, number, date, boolean, select.
    field_type varchar(20)  NOT NULL,
    required   boolean      NOT NULL DEFAULT false,
    -- options lists the permitted values for a select, and is empty otherwise.
    options    jsonb        NOT NULL DEFAULT '[]'::jsonb,
    position   integer      NOT NULL DEFAULT 0,
    created_at timestamptz  NOT NULL DEFAULT now(),
    updated_at timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT uq_custom_field_definition UNIQUE (company_id, resource, key)
);

ALTER TABLE custom_field_definition ADD CONSTRAINT fk_custom_field_definition_company
    FOREIGN KEY (company_id) REFERENCES company(id) ON DELETE CASCADE;

-- Forms are drawn per resource, in order.
CREATE INDEX idx_custom_field_definition_resource ON custom_field_definition (company_id, resource, position);

-- Filtering by a custom field is the main thing people want from one, and a
-- jsonb containment query without a GIN index is a sequential scan of the whole
-- table.
CREATE INDEX idx_employee_custom_fields       ON employee       USING gin (custom_fields);
CREATE INDEX idx_asset_custom_fields          ON asset          USING gin (custom_fields);
CREATE INDEX idx_part_custom_fields           ON part           USING gin (custom_fields);
CREATE INDEX idx_vendor_custom_fields         ON vendor         USING gin (custom_fields);
CREATE INDEX idx_work_order_custom_fields     ON work_order     USING gin (custom_fields);
CREATE INDEX idx_issue_custom_fields          ON issue          USING gin (custom_fields);
CREATE INDEX idx_purchase_order_custom_fields ON purchase_order USING gin (custom_fields);
CREATE INDEX idx_service_entry_custom_fields  ON service_entry  USING gin (custom_fields);

-- Nothing is seeded, and existing documents are left exactly as they are. A
-- company that has been writing keys no definition describes keeps them: they
-- are that tenant's data, and rejecting them on the next write would make
-- previously-saved records uneditable through a form that no longer shows the
-- fields holding them.
