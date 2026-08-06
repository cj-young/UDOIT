-- name: CreateCourse :execresult
INSERT INTO course (title, tenant_id, is_active, is_dirty, external_id, external_data, updated_at)
VALUES (
  sqlc.arg(title),
  sqlc.arg(tenant_id),
  sqlc.arg(is_active),
  sqlc.arg(is_dirty),
  sqlc.arg(external_id),
  sqlc.arg(external_data),
  sqlc.arg(updated_at)
);
