package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"rewritetest/internal/lms/internal/domain"
	lmssqlc "rewritetest/internal/lms/internal/infrastructure/sqlc"
	"rewritetest/internal/shared/apperr"
)

type MySQLLMSProviderConfigRepository struct {
	queries *lmssqlc.Queries
}

func NewMySQLLMSProviderConfigRepository(db *sql.DB) *MySQLLMSProviderConfigRepository {
	return &MySQLLMSProviderConfigRepository{
		queries: lmssqlc.New(db),
	}
}

func (r *MySQLLMSProviderConfigRepository) GetByTenant(ctx context.Context, tenantID int64) (*domain.LMSProviderConfig, error) {
	result, err := r.queries.GetLMSProviderConfigByTenant(ctx, uint64(tenantID))
	if err != nil {
		return nil, err
	}

	var config map[string]any
	if err := json.Unmarshal(result.ConfigJson, &config); err != nil {
		return nil, err
	}

	lmsType := domain.LMSType(result.LmsType)
	if !lmsType.IsValid() {
		return nil, apperr.Internal("An invalid LMS type was found in the requested LMS config.")
	}

	return domain.NewLMSProviderConfig(
		int64(result.TenantID),
		lmsType,
		config,
	), nil
}

func (r *MySQLLMSProviderConfigRepository) UpsertByTenant(ctx context.Context, tenantID int64, lmsKey domain.LMSType, data map[string]any) error {
	configJSON, err := json.Marshal(data)
	if err != nil {
		return apperr.Internal("Failed to marshal LMS provider config")
	}

	return r.queries.UpsertLMSProviderConfigByTenant(ctx, lmssqlc.UpsertLMSProviderConfigByTenantParams{
		TenantID:   uint64(tenantID),
		LmsType:    string(lmsKey),
		ConfigJson: configJSON,
	})
}
