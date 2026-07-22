-- name: CreateCourse :execresult
INSERT INTO course (title, is_active, is_dirty, updated_at)
VALUES (
  sqlc.arg(title),
  sqlc.arg(is_active),
  sqlc.arg(is_dirty),
  sqlc.arg(updated_at)
);
