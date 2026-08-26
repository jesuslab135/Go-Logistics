-- 000017_custom_field_definitions.down.sql
-- The stored documents are untouched by this rollback; only the descriptions of
-- what may go in them, and the indexes for querying them, are removed.
DROP INDEX IF EXISTS idx_service_entry_custom_fields;
DROP INDEX IF EXISTS idx_purchase_order_custom_fields;
DROP INDEX IF EXISTS idx_issue_custom_fields;
DROP INDEX IF EXISTS idx_work_order_custom_fields;
DROP INDEX IF EXISTS idx_vendor_custom_fields;
DROP INDEX IF EXISTS idx_part_custom_fields;
DROP INDEX IF EXISTS idx_asset_custom_fields;
DROP INDEX IF EXISTS idx_employee_custom_fields;
DROP TABLE IF EXISTS custom_field_definition;
