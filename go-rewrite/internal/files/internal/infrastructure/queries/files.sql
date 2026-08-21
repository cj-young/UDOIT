-- name: GetFileByID :one
SELECT id, course_id, file_name, file_type, updated_at_lms, active, is_available, is_hidden, file_size, download_url, reviewed_by_id, reviewed_on, external_data, external_id, reviewed
FROM file_item
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: UpdateFile :exec
UPDATE file_item
SET
  course_id = sqlc.arg(course_id),
  file_name = sqlc.arg(file_name),
  file_type = sqlc.arg(file_type),
  updated_at_lms = sqlc.arg(updated_at_lms),
  active = sqlc.arg(active),
  is_available = sqlc.arg(is_available),
  is_hidden = sqlc.arg(is_hidden),
  file_size = sqlc.arg(file_size),
  download_url = sqlc.arg(download_url),
  reviewed_by_id = sqlc.arg(reviewed_by_id),
  reviewed_on = sqlc.arg(reviewed_on),
  reviewed = sqlc.arg(reviewed),
  external_data = sqlc.arg(external_data),
  external_id = sqlc.arg(external_id)
WHERE id = sqlc.arg(id);

-- name: UpsertFileByCourseExternalID :exec
INSERT INTO file_item (
  course_id,
  file_name,
  file_type,
  updated_at_lms,
  active,
  is_available,
  is_hidden,
  file_size,
  download_url,
  external_data,
  external_id
)
VALUES (
  sqlc.arg(course_id),
  sqlc.arg(file_name),
  sqlc.arg(file_type),
  sqlc.arg(updated_at_lms),
  sqlc.arg(active),
  sqlc.arg(is_available),
  sqlc.arg(is_hidden),
  sqlc.arg(file_size),
  sqlc.arg(download_url),
  sqlc.arg(external_data),
  sqlc.arg(external_id)
)
ON DUPLICATE KEY UPDATE
  file_name = VALUES(file_name),
  file_type = VALUES(file_type),
  updated_at_lms = VALUES(updated_at_lms),
  active = VALUES(active),
  is_available = VALUES(is_available),
  is_hidden = VALUES(is_hidden),
  file_size = VALUES(file_size),
  download_url = VALUES(download_url),
  external_data = VALUES(external_data),
  external_id = VALUES(external_id);

-- name: DeleteFile :exec
DELETE FROM file_item
WHERE id = sqlc.arg(id);
