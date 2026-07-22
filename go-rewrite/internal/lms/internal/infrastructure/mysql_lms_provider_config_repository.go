package infrastructure

import (
	"context"
	"database/sql"
	"rewritetest/internal/lms/internal/domain"
)

type MySQLLMSProviderConfigRepository struct {
	db *sql.DB
}

func NewMySQLLMSProviderConfigRepository(db *sql.DB) *MySQLLMSProviderConfigRepository {
	return &MySQLLMSProviderConfigRepository{
		db: db,
	}
}

func (r *MySQLLMSProviderConfigRepository) GetByTenant(ctx context.Context, tenantID int64) (*domain.LMSProviderConfig, error) {
	
	return domain.NewLMSProviderConfig(tenantID, domain.LMSTypeCanvas, map[string]any{
		"baseUrl": "https://devhub.cdl.ucf.edu",
	}), nil
}