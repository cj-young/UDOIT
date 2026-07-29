package lms

import (
	"context"
	"database/sql"
	"fmt"

	"rewritetest/internal/lms/internal"
	"rewritetest/internal/lms/internal/application"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/lms/internal/infrastructure"
	"rewritetest/internal/lms/internal/infrastructure/providers/canvas"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Module struct {
	providerRegistry           domain.LMSProviderRegistry
	providerConfigRepository   domain.LMSProviderConfigRepository
	objectMappingRepository domain.LMSObjectMappingRepository
	handler                    *internal.Handler
}

func New(db *sql.DB, client *redis.Client, baseURL string) *Module {

	authAttemptTTL := time.Hour

	credentialRepository := infrastructure.NewMySQLLMSCredentialRepository(db)
	providerConfigRepository := infrastructure.NewMySQLLMSProviderConfigRepository(db)
	lmsObjectMappingRepository := infrastructure.NewMySQLLMSObjectMappingRepository(db)
	authAttemptRepository := infrastructure.NewRedisAuthAttemptRepository(client, authAttemptTTL, "auth_attempt:")

	// LMS providers

	oauthRedirectURI := fmt.Sprintf("%s/oauth/callback", baseURL)
	canvasProvider := canvas.NewCanvasLMSProvider(credentialRepository, authAttemptRepository, oauthRedirectURI)

	providerRegistry := infrastructure.NewMapLMSProviderRegistry()
	providerRegistry.RegisterProvider(domain.LMSTypeCanvas, canvasProvider)

	processOAuthRedirectUseCase := application.NewProcessOAuthRedirectUseCase(providerRegistry, providerConfigRepository, authAttemptRepository)

	handler := internal.NewHandler(processOAuthRedirectUseCase)

	return &Module{
		providerRegistry:           providerRegistry,
		providerConfigRepository:   providerConfigRepository,
		objectMappingRepository:    lmsObjectMappingRepository,
		handler:                    handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	m.handler.RegisterRoutes(rg)
}

func (m *Module) IsValidLMSType(lmsKey string) bool {
	lmsType := domain.LMSType(lmsKey)
	return lmsType.IsValid()
}

func (m *Module) SaveProviderConfig(ctx context.Context, tenantID int64, lmsKey string, configData map[string]any) error {
	return m.providerConfigRepository.UpsertByTenant(ctx, tenantID, domain.LMSType(lmsKey), configData)
}

func (m *Module) ValidateProviderConfig(ctx context.Context, lmsKey string, configData map[string]any) error {
	provider, err := m.providerRegistry.Get(ctx, domain.LMSType(lmsKey))
	if err != nil {
		return err
	}
	return provider.ValidateConfig(configData)
}