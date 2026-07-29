package canvas

import (
	"fmt"
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

type CanvasCredential struct {
	apiToken     string
	refreshToken string
	expiresAt    *time.Time
}

type CanvasConfig struct {
	baseURL      string
	clientID     string
	clientSecret string
}

func (p *CanvasLMSProvider) asCanvasCredential(cred domain.LMSCredential) (CanvasCredential, error) {
	if cred.LMSKey() != "canvas" {
		return CanvasCredential{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Credential", fmt.Sprintf("Expected LMSKey 'canvas', got '%s'", cred.LMSKey()),
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasCredential"),
		)
	}

	apiToken, ok := cred.Payload()["api_token"].(string)
	if !ok {
		return CanvasCredential{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Credential", "Missing or invalid 'apiToken' in credential data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasCredential"),
		)
	}

	refreshToken, ok := cred.Payload()["refresh_token"].(string)
	if !ok {
		return CanvasCredential{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Credential", "Missing or invalid 'refreshToken' in credential data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasCredential"),
		)
	}

	return CanvasCredential{
		apiToken:     apiToken,
		refreshToken: refreshToken,
		expiresAt:    cred.ExpiresAt(),
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

func (p *CanvasLMSProvider) asCanvasConfig(config domain.LMSProviderConfig) (CanvasConfig, error) {
	baseURL, ok := config.Data()["base_url"].(string)
	if !ok {
		return CanvasConfig{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Config", "Missing or invalid 'base_url' in config data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasConfig"),
		)
	}

	clientID, ok := config.Data()["client_id"].(string)
	if !ok {
		return CanvasConfig{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Config", "Missing or invalid 'client_id' in config data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasConfig"),
		)
	}

	clientSecret, ok := config.Data()["client_secret"].(string)
	if !ok {
		return CanvasConfig{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Config", "Missing or invalid 'client_secret' in config data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasConfig"),
		)
	}

	return CanvasConfig{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}, nil
}

func (p *CanvasLMSProvider) asCanvasFile(mapping domain.LMSObjectMapping) (CanvasFile, error) {
	if mapping.ObjectType() != domain.LMSObjectTypeFile {
		return CanvasFile{}, apperr.New(
			apperr.CodeInternal, "Invalid LMS object mapping", fmt.Sprintf("Expected object type '%s', got '%s'", domain.LMSObjectTypeFile, mapping.ObjectType()),
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasFile"),
		)
	}

	fileID, ok := mapping.Data()["file_id"].(string)
	if !ok {
		return CanvasFile{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS File", "Missing or invalid 'fileId' in mapping data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasFile"),
		)
	}

	contextType, ok := mapping.Data()["context_type"].(string)
	if !ok {
		return CanvasFile{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS File", "Missing or invalid 'contextType' in mapping data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasFile"),
		)
	}

	return CanvasFile{
		fileID:      fileID,
		contextType: contextType,
	}, nil
}

var _ domain.FullLMSProvider = (*CanvasLMSProvider)(nil)
var _ domain.OAuthRedirectProcessor = (*CanvasLMSProvider)(nil)
