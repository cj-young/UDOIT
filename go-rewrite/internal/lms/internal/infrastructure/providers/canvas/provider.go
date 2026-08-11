package canvas

import (
	"net/http"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
	"time"
)

type CanvasLMSProvider struct {
	lmsCredentialRepository domain.LMSCredentialRepository
	authAttemptRepository   domain.AuthAttemptRepository
	httpClient              *http.Client
	oauthRedirectURI        string
}

func NewCanvasLMSProvider(
	lmsCredentialRepository domain.LMSCredentialRepository,
	authAttemptRepository domain.AuthAttemptRepository,
	oauthRedirectURI string,
) *CanvasLMSProvider {
	return &CanvasLMSProvider{
		lmsCredentialRepository: lmsCredentialRepository,
		authAttemptRepository:   authAttemptRepository,
		httpClient:              http.DefaultClient,
		oauthRedirectURI:        oauthRedirectURI,
	}
}

type canvasContent struct {
	contentID string
	updatedAt time.Time
}

type canvasConfig struct {
	baseURL      string
	clientID     string
	clientSecret string
}

func (p *CanvasLMSProvider) asCanvasContent(content domain.LMSContent) (canvasContent, error) {
	contentID, ok := content.ExternalData["content_id"].(string)
	if !ok {
		return canvasContent{}, apperr.New(apperr.CodeInternal, "invalid_canvas_content", "Missing or invalid 'content_id' in content mapping data")
	}

	updatedAt, ok := content.ExternalData["updated_at"].(time.Time)
	if !ok {
		return canvasContent{}, apperr.New(apperr.CodeInternal, "invalid_canvas_content", "Missing or invalid 'updated_at' in content mapping data")
	}

	return canvasContent{
		contentID: contentID,
		updatedAt: updatedAt,
	}, nil
}

func (p *CanvasLMSProvider) ValidateConfig(configData map[string]any) error {
	if _, ok := configData["base_url"].(string); !ok {
		return apperr.New(
			apperr.CodeValidation, "invalid_canvas_config", "Missing or non-string 'base_url' in Canvas provider config",
			apperr.WithOp("lms.infrastructure.canvas.ValidateConfig"),
		)
	}

	if _, ok := configData["client_id"].(string); !ok {
		return apperr.New(
			apperr.CodeValidation, "invalid_canvas_config", "Missing or non-string 'client_id' in Canvas provider config",
			apperr.WithOp("lms.infrastructure.canvas.ValidateConfig"),
		)
	}

	if _, ok := configData["client_secret"].(string); !ok {
		return apperr.New(
			apperr.CodeValidation, "invalid_canvas_config", "Missing or non-string 'client_secret' in Canvas provider config",
			apperr.WithOp("lms.infrastructure.canvas.ValidateConfig"),
		)
	}

	return nil
}

func (p *CanvasLMSProvider) asCanvasConfig(config domain.LMSProviderConfig) (canvasConfig, error) {
	baseURL, ok := config.Data()["base_url"].(string)
	if !ok {
		return canvasConfig{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Config", "Missing or invalid 'base_url' in config data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasConfig"),
		)
	}

	clientID, ok := config.Data()["client_id"].(string)
	if !ok {
		return canvasConfig{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Config", "Missing or invalid 'client_id' in config data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasConfig"),
		)
	}

	clientSecret, ok := config.Data()["client_secret"].(string)
	if !ok {
		return canvasConfig{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Config", "Missing or invalid 'client_secret' in config data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasConfig"),
		)
	}

	return canvasConfig{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}, nil
}

func (p *CanvasLMSProvider) asCanvasFile(file domain.LMSFile) (canvasFile, error) {
	fileID, ok := file.ExternalData["file_id"].(string)
	if !ok {
		return canvasFile{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS File", "Missing or invalid 'fileId' in mapping data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasFile"),
		)
	}

	contextType, ok := file.ExternalData["context_type"].(string)
	if !ok {
		return canvasFile{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS File", "Missing or invalid 'contextType' in mapping data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasFile"),
		)
	}

	return canvasFile{
		fileID:      fileID,
		contextType: contextType,
	}, nil
}

var _ domain.FullLMSProvider = (*CanvasLMSProvider)(nil)
var _ domain.OAuthRedirectProcessor = (*CanvasLMSProvider)(nil)
