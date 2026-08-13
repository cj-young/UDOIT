package canvas

import (
	"net/http"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
)

type CanvasLMSProvider struct {
	lmsCredentialRepository domain.LMSCredentialRepository
	authAttemptRepository   domain.AuthAttemptRepository
	httpClient              *http.Client
	oauthRedirectURI        string
	config                  canvasConfig
}

func NewCanvasLMSProvider(
	lmsCredentialRepository 	domain.LMSCredentialRepository,
	authAttemptRepository 		domain.AuthAttemptRepository,
	oauthRedirectURI 					string,
	config 										domain.LMSProviderConfig,									
) (*CanvasLMSProvider, error) {

	canvasConfig, err := asCanvasConfig(config)
	if err != nil {
		return nil, err
	}

	return &CanvasLMSProvider{
		lmsCredentialRepository: lmsCredentialRepository,
		authAttemptRepository:   authAttemptRepository,
		httpClient:              http.DefaultClient,
		oauthRedirectURI:        oauthRedirectURI,
		config:                  canvasConfig,
	}, nil
}

type canvasConfig struct {
	baseURL      	string
	clientID     	string
	clientSecret 	string
	tenantID			int64
}

func ValidateConfig(configData map[string]any) error {
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

func asCanvasConfig(config domain.LMSProviderConfig) (canvasConfig, error) {
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
		tenantID:     config.TenantID(),
	}, nil
}

var _ domain.FullLMSProvider = (*CanvasLMSProvider)(nil)
var _ domain.OAuthRedirectProcessor = (*CanvasLMSProvider)(nil)
