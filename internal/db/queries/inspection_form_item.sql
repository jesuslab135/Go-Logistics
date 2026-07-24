-- name: ListInspectionFormItems :many
SELECT c.* FROM inspection_form_item c JOIN inspection_form p ON p.id = c.form_id
WHERE c.form_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.position LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountInspectionFormItems :one
SELECT count(*) FROM inspection_form_item c JOIN inspection_form p ON p.id = c.form_id WHERE c.form_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetInspectionFormItem :one
SELECT c.* FROM inspection_form_item c JOIN inspection_form p ON p.id = c.form_id
WHERE c.id = sqlc.arg(id) AND c.form_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateInspectionFormItem :one
INSERT INTO inspection_form_item (
    form_id, item_type, label, short_description, instructions, position, is_required, pass_label, fail_label, na_label, enable_na_option, require_remark_on_fail, require_remark_on_pass, require_photo_on_fail, require_meter_entry_photo_verification, require_secondary_meter_if_one_exists, type_config, created_at, updated_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(item_type), sqlc.arg(label), sqlc.arg(short_description), sqlc.arg(instructions), sqlc.arg(position), sqlc.arg(is_required), sqlc.arg(pass_label), sqlc.arg(fail_label), sqlc.arg(na_label), sqlc.arg(enable_na_option), sqlc.arg(require_remark_on_fail), sqlc.arg(require_remark_on_pass), sqlc.arg(require_photo_on_fail), sqlc.arg(require_meter_entry_photo_verification), sqlc.arg(require_secondary_meter_if_one_exists), sqlc.arg(type_config), sqlc.arg(created_at), sqlc.arg(updated_at)
WHERE EXISTS (SELECT 1 FROM inspection_form WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateInspectionFormItem :one
UPDATE inspection_form_item AS c SET item_type = sqlc.arg(item_type), label = sqlc.arg(label), short_description = sqlc.arg(short_description), instructions = sqlc.arg(instructions), position = sqlc.arg(position), is_required = sqlc.arg(is_required), pass_label = sqlc.arg(pass_label), fail_label = sqlc.arg(fail_label), na_label = sqlc.arg(na_label), enable_na_option = sqlc.arg(enable_na_option), require_remark_on_fail = sqlc.arg(require_remark_on_fail), require_remark_on_pass = sqlc.arg(require_remark_on_pass), require_photo_on_fail = sqlc.arg(require_photo_on_fail), require_meter_entry_photo_verification = sqlc.arg(require_meter_entry_photo_verification), require_secondary_meter_if_one_exists = sqlc.arg(require_secondary_meter_if_one_exists), type_config = sqlc.arg(type_config), updated_at = sqlc.arg(updated_at)
FROM inspection_form p
WHERE c.id = sqlc.arg(id) AND c.form_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.form_id
RETURNING c.*;

-- name: DeleteInspectionFormItem :exec
DELETE FROM inspection_form_item AS c USING inspection_form p
WHERE c.id = sqlc.arg(id) AND c.form_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.form_id;
