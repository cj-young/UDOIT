package tenants

import (
	"context"
	"database/sql"

	"rewritetest/internal/tenants/internal/application"
	"rewritetest/internal/tenants/internal/infrastructure"
)

type Module struct {
	registerTenantUseCase *application.RegisterTenantUseCase
}

func New(db *sql.DB, lmsTypeValidator application.LMSTypeValidator) *Module {
	tenantRepository := infrastructure.NewMySQLTenantRepository(db)
	registerTenantUseCase := application.NewRegisterTenantUseCase(tenantRepository, lmsTypeValidator)

	return &Module{
		registerTenantUseCase: registerTenantUseCase,
	}
}

func (m *Module) RegisterTenant(ctx context.Context, lmsKey string) (int64, error) {
	return m.registerTenantUseCase.Execute(ctx, application.RegisterTenantCommand{LMSKey: lmsKey})
}
