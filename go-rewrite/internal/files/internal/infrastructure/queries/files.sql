-- name: GetFileByID :one
SELECT id, course_id, reviewed_by_id, reviewed_on, external_data, external_id, reviewed
FROM file_item
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: UpdateFile :exec
UPDATE file_item
SET
  course_id = sqlc.arg(course_id),
  reviewed_by_id = sqlc.arg(reviewed_by_id),
  reviewed_on = sqlc.arg(reviewed_on),
  reviewed = sqlc.arg(reviewed),
  external_data = sqlc.arg(external_data),
  external_id = sqlc.arg(external_id)
WHERE id = sqlc.arg(id);

-- name: DeleteFile :exec
DELETE FROM file_item
WHERE id = sqlc.arg(id);
