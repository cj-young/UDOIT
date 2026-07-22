-- name: GetFileByID :one
SELECT id, course_id, reviewed_by_id, reviewed_on, reviewed
FROM file_item
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: UpdateFile :exec
UPDATE file_item
SET
  course_id = sqlc.arg(course_id),
  reviewed_by_id = sqlc.arg(reviewed_by_id),
  reviewed_on = sqlc.arg(reviewed_on),
  reviewed = sqlc.arg(reviewed)
WHERE id = sqlc.arg(id);

-- name: DeleteFile :exec
DELETE FROM file_item
WHERE id = sqlc.arg(id);
