-- name: UpsertLMSUserMapping :exec
INSERT INTO lms_user_mapping (
  user_id,
  lms_key,
  external_user_id,
  api_domain,
  metadata,
  created_at,
  updated_at
)
VALUES (
  sqlc.arg(user_id),
  sqlc.arg(lms_key),
  sqlc.arg(external_user_id),
  sqlc.arg(api_domain),
  sqlc.arg(metadata),
  NOW(),
  NOW()
)
ON DUPLICATE KEY UPDATE
  lms_key = VALUES(lms_key),
  external_user_id = VALUES(external_user_id),
  api_domain = VALUES(api_domain),
  metadata = VALUES(metadata),
  updated_at = NOW();

-- name: GetLMSUserMappingByUserID :one
SELECT user_id, lms_key, COALESCE(external_user_id, ''), COALESCE(api_domain, ''), COALESCE(metadata, '{}'), created_at, updated_at
FROM lms_user_mapping
WHERE user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: UpsertLMSUserCredential :exec
INSERT INTO lms_user_credential (
  user_id,
  lms_key,
  schema_name,
  credential_json,
  expires_at,
  is_active,
  created_at,
  updated_at
)
VALUES (
  sqlc.arg(user_id),
  sqlc.arg(lms_key),
  sqlc.arg(schema_name),
  sqlc.arg(credential_json),
  sqlc.arg(expires_at),
  1,
  NOW(),
  NOW()
)
ON DUPLICATE KEY UPDATE
  schema_name = VALUES(schema_name),
  credential_json = VALUES(credential_json),
  expires_at = VALUES(expires_at),
  is_active = 1,
  updated_at = NOW();

-- name: GetActiveLMSUserCredentialByUserID :one
SELECT user_id, lms_key, credential_json, expires_at, is_active, created_at, updated_at
FROM lms_user_credential
WHERE user_id = sqlc.arg(user_id)
  AND is_active = 1
LIMIT 1;

-- name: GetLMSObjectMappingByTypeAndInternalID :one
SELECT lms_key, mapping_json
FROM lms_object_mapping
WHERE object_type = sqlc.arg(object_type)
  AND internal_id = sqlc.arg(internal_id)
LIMIT 1;
