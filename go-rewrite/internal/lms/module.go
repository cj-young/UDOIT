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
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type Module struct {
	providerRegistry           domain.LMSProviderRegistry
	providerConfigRepository   domain.LMSProviderConfigRepository
	handler                    *internal.Handler
}

func New(db *sql.DB, client *redis.Client, baseURL string) *Module {

	authAttemptTTL := time.Hour

	credentialRepository := infrastructure.NewMySQLLMSCredentialRepository(db)
	providerConfigRepository := infrastructure.NewMySQLLMSProviderConfigRepository(db)
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

type ContentItemDTO struct {
	ExternalID 	string
	HTML 				string
	Type        string
}

type GetContentRequest struct {
	CourseID 						int64
	ExternalCourseID 		string
	ExternalCourseData	map[string]any
	TenantID						int64
	UserID							int64
}

func (m *Module) GetContent(ctx context.Context, req GetContentRequest) ([]ContentItemDTO, error) {
	lmsProvider, tenantConfig, err := application.GetLMSProviderAndConfig(
		ctx,
		m.providerRegistry,
		m.providerConfigRepository,
		req.TenantID,
	)
	if err != nil {
		return nil, err
	}
	
	contentItems, err := lmsProvider.GetContent(ctx, tenantConfig, domain.LMSCourse{
		ID:          	req.CourseID,
		ExternalID:   req.ExternalCourseID,
		ExternalData: req.ExternalCourseData,
	}, []domain.LMSContent{}, req.UserID)
	if err != nil {
		return nil, err
	}
	
	contentItemDTOs := make([]ContentItemDTO, len(contentItems))
	for i, item := range contentItems {
		contentItemDTOs[i] = ContentItemDTO{
			ExternalID: item.ExternalID,
			HTML:       item.HTML,
			Type:       string(item.Type),
		}
	}
	return contentItemDTOs, nil
}

func (m *Module) GetCourseInfoFromLTILaunch(ctx context.Context, tenantID int64, claims jwt.MapClaims) (string, map[string]any, error) {
	lmsProvider, tenantConfig, err := application.GetLMSProviderAndConfig(
		ctx,
		m.providerRegistry,
		m.providerConfigRepository,
		tenantID,
	)
	if err != nil {
		return "", nil, err
	}

	externalID, externalData, err := lmsProvider.GetCourseInfoFromLTILaunch(ctx, tenantConfig, claims)
	if err != nil {
		return "", nil, err
	}

	return externalID, externalData, nil
}