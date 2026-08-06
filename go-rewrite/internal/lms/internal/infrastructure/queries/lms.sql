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

-- name: GetLMSProviderConfigByTenant :one
SELECT tenant_id, lms_type, config_json, created_at, updated_at
FROM lms_provider_config
WHERE tenant_id = sqlc.arg(tenant_id)
LIMIT 1;

-- name: UpsertLMSProviderConfigByTenant :exec
INSERT INTO lms_provider_config (tenant_id, lms_type, config_json)
VALUES (sqlc.arg(tenant_id), sqlc.arg(lms_type), sqlc.arg(config_json))
ON DUPLICATE KEY UPDATE
  lms_type = VALUES(lms_type),
  config_json = VALUES(config_json),
  updated_at = NOW();