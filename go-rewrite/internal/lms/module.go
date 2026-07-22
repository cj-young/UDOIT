package lms

import (
	"context"
	"database/sql"

	"rewritetest/internal/lms/internal/application"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/lms/internal/infrastructure"

	"rewritetest/internal/shared/auth"
)

type Module struct {
	deleteFileUseCase *application.DeleteFileUseCase
}

func New(db *sql.DB) *Module {
	credentialRepository := infrastructure.NewMySQLLMSCredentialRepository(db)
	providerConfigRepository := infrastructure.NewMySQLLMSProviderConfigRepository(db)
	lmsObjectMappingRepository := infrastructure.NewMySQLLMSObjectMappingRepository(db)

	providerRegistry := infrastructure.NewMapLMSProviderRegistry()
	providerRegistry.RegisterProvider(domain.LMSTypeCanvas, infrastructure.NewCanvasLMSProvider(credentialRepository))

	deleteFileUseCase := application.NewDeleteFileUseCase(providerRegistry, providerConfigRepository, lmsObjectMappingRepository)

	return &Module{
		deleteFileUseCase: deleteFileUseCase,
	}
}

func (m *Module) DeleteFile(ctx context.Context, principal auth.Principal, fileID int64) error {
	return m.deleteFileUseCase.Execute(ctx, principal, fileID)
}

func (m *Module) IsValidLMSType(lmsKey string) bool {
	lmsType := domain.LMSType(lmsKey)
	return lmsType.IsValid()
}