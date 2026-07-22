package lms

import (
	"context"
	"database/sql"

	"rewritetest/internal/lms/internal"
	"rewritetest/internal/lms/internal/application"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/lms/internal/infrastructure"

	"rewritetest/internal/shared/auth"

	"github.com/gin-gonic/gin"
)

type Module struct {
	handler           *internal.Handler
	deleteFileUseCase *application.DeleteFileUseCase
}

func New(db *sql.DB, internalSyncSecret string) *Module {
	mappingRepository := infrastructure.NewMySQLUserLMSMappingRepository(db)
	credentialRepository := infrastructure.NewMySQLLMSCredentialRepository(db)
	providerConfigRepository := infrastructure.NewMySQLLMSProviderConfigRepository(db)
	lmsObjectMappingRepository := infrastructure.NewMySQLLMSObjectMappingRepository(db)

	providerRegistry := infrastructure.NewMapLMSProviderRegistry()
	providerRegistry.RegisterProvider(domain.LMSTypeCanvas, infrastructure.NewCanvasLMSProvider(credentialRepository))

	syncSessionUseCase := application.NewSyncSessionUseCase(mappingRepository, credentialRepository)
	deleteFileUseCase := application.NewDeleteFileUseCase(providerRegistry, providerConfigRepository, lmsObjectMappingRepository)
	handler := internal.NewHandler(syncSessionUseCase, internalSyncSecret)

	return &Module{
		handler:           handler,
		deleteFileUseCase: deleteFileUseCase,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	m.handler.RegisterRoutes(rg)
}

func (m *Module) DeleteFile(ctx context.Context, principal auth.Principal, fileID int64) error {
	return m.deleteFileUseCase.Execute(ctx, principal, fileID)
}
