-- name: CreateTenant :execresult
INSERT INTO tenant (lms_key)
VALUES (sqlc.arg(lms_key));
