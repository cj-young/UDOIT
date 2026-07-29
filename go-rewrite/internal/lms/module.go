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

	"rewritetest/internal/shared/auth"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Module struct {
	deleteFileUseCase          *application.DeleteFileUseCase
	beginAuthenticationUseCase *application.BeginAuthenticationUseCase
	providerRegistry           domain.LMSProviderRegistry
	providerConfigRepository   domain.LMSProviderConfigRepository
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

	deleteFileUseCase := application.NewDeleteFileUseCase(providerRegistry, providerConfigRepository, lmsObjectMappingRepository)
	beginAuthenticationUseCase := application.NewBeginAuthenticationUseCase(providerRegistry, providerConfigRepository)
	processOAuthRedirectUseCase := application.NewProcessOAuthRedirectUseCase(providerRegistry, providerConfigRepository, authAttemptRepository)

	handler := internal.NewHandler(processOAuthRedirectUseCase)

	return &Module{
		deleteFileUseCase:          deleteFileUseCase,
		beginAuthenticationUseCase: beginAuthenticationUseCase,
		providerRegistry:           providerRegistry,
		providerConfigRepository:   providerConfigRepository,
		handler:                    handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	m.handler.RegisterRoutes(rg)
}

// Redirect URL probably shouldn't be an explicit field because different kinds
// might not have redirect URLs
//
// This struct acts as a DTO for auth challenges and is functionally separate from
// the auth challenge defined in the domain
type AuthChallenge struct {
	Kind        AuthChallengeKind
	RedirectURL string
}

type AuthChallengeKind string

const (
	AuthChallengeKindRedirect AuthChallengeKind = "redirect"
	AuthChallengeKindNone     AuthChallengeKind = "none"
)

func (m *Module) DeleteFile(ctx context.Context, principal auth.Principal, fileID int64) error {
	return m.deleteFileUseCase.Execute(ctx, principal, fileID)
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

func (m *Module) BeginAuthentication(ctx context.Context, userID int64, tenantID int64, targetLinkURI string) (AuthChallenge, error) {
	authChallenge, err := m.beginAuthenticationUseCase.Execute(ctx, userID, tenantID, targetLinkURI)
	if err != nil {
		return AuthChallenge{}, err
	}

	return AuthChallenge{
		Kind:        AuthChallengeKind(authChallenge.Kind),
		RedirectURL: authChallenge.RedirectURL,
	}, nil
}

func (m *Module) GetCourseContent(ctx context.Context, principal auth.Principal, courseID int64) error {
	return nil
}
