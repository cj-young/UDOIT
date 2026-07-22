package infrastructure

import (
	"context"
	"database/sql"

	"rewritetest/internal/tenants/internal/domain"
	tenantssqlc "rewritetest/internal/tenants/internal/infrastructure/sqlc"
)

type MySQLTenantRepository struct {
	queries *tenantssqlc.Queries
}

func NewMySQLTenantRepository(db *sql.DB) *MySQLTenantRepository {
	return &MySQLTenantRepository{queries: tenantssqlc.New(db)}
}

func (r *MySQLTenantRepository) Create(ctx context.Context, lmsKey string) (int64, error) {
	result, err := r.queries.CreateTenant(ctx, lmsKey)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

var _ domain.TenantRepository = (*MySQLTenantRepository)(nil)
