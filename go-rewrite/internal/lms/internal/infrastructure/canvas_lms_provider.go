package infrastructure

import (
	"context"
	"fmt"
	"net/http"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
	"rewritetest/internal/shared/auth"
)

type CanvasLMSProvider struct {
	lmsCredentialRepository domain.LMSCredentialRepository
}

func NewCanvasLMSProvider(
	lmsCredentialRepository domain.LMSCredentialRepository,
) *CanvasLMSProvider {
	return &CanvasLMSProvider{
		lmsCredentialRepository: lmsCredentialRepository,
	}
}

type CanvasCredential struct {
	apiToken     string
	refreshToken string
}

type CanvasConfig struct {
	baseURL string
}

type CanvasFile struct {
	fileID string
}

func (p *CanvasLMSProvider) DeleteFile(ctx context.Context, principal auth.Principal, config domain.LMSProviderConfig, fileMapping domain.LMSObjectMapping) error {

	cred, err := p.lmsCredentialRepository.GetActiveByUser(ctx, principal.AgentID)
	if err != nil {
		return err
	}
	if cred == nil {
		return apperr.New(
			apperr.CodeNotFound, "Canvas LMS Credential not found", "",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.DeleteFile"),
		)
	}

	validatedConfig, err := p.asCanvasConfig(config)
	if err != nil {
		return err
	}

	credential, err := p.asCanvasCredential(*cred)
	if err != nil {
		return err
	}

	canvasFile, err := p.asCanvasFile(fileMapping)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/files/%s", validatedConfig.baseURL, canvasFile.fileID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return apperr.New(
			apperr.CodeInternal, "Failed to create HTTP request", err.Error(),
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.DeleteFile"),
			apperr.WithCause(err),
		)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", credential.apiToken))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return apperr.New(
			apperr.CodeInternal, "Failed to send HTTP request", err.Error(),
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.DeleteFile"),
			apperr.WithCause(err),
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return apperr.New(
			apperr.CodeInternal, "Failed to delete file", fmt.Sprintf("Unexpected status code: %d", resp.StatusCode),
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.DeleteFile"),
		)
	}

	return nil
}

func (p *CanvasLMSProvider) asCanvasCredential(cred domain.LMSCredential) (CanvasCredential, error) {
	if cred.LMSKey() != "canvas" {
		return CanvasCredential{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Credential", fmt.Sprintf("Expected LMSKey 'canvas', got '%s'", cred.LMSKey()),
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasCredential"),
		)
	}

	apiToken, ok := cred.Payload()["apiToken"].(string)
	if !ok {
		return CanvasCredential{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Credential", "Missing or invalid 'apiToken' in credential data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasCredential"),
		)
	}

	refreshToken, ok := cred.Payload()["refreshToken"].(string)
	if !ok {
		return CanvasCredential{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Credential", "Missing or invalid 'refreshToken' in credential data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasCredential"),
		)
	}

	return CanvasCredential{
		apiToken:     apiToken,
		refreshToken: refreshToken,
	}, nil
}

func (p *CanvasLMSProvider) asCanvasConfig(config domain.LMSProviderConfig) (CanvasConfig, error) {
	baseURL, ok := config.Data()["baseUrl"].(string)
	if !ok {
		return CanvasConfig{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS Config", "Missing or invalid 'baseUrl' in config data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasConfig"),
		)
	}

	return CanvasConfig{
		baseURL: baseURL,
	}, nil
}

func (p *CanvasLMSProvider) asCanvasFile(mapping domain.LMSObjectMapping) (CanvasFile, error) {
	if mapping.ObjectType() != domain.LMSObjectTypeFile {
		return CanvasFile{}, apperr.New(
			apperr.CodeInternal, "Invalid LMS object mapping", fmt.Sprintf("Expected object type '%s', got '%s'", domain.LMSObjectTypeFile, mapping.ObjectType()),
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasFile"),
		)
	}

	fileID, ok := mapping.Data()["fileId"].(string)
	if !ok {
		return CanvasFile{}, apperr.New(
			apperr.CodeInternal, "Invalid Canvas LMS File", "Missing or invalid 'fileId' in mapping data",
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.asCanvasFile"),
		)
	}

	return CanvasFile{
		fileID: fileID,
	}, nil
}

var _ domain.LMSProvider = (*CanvasLMSProvider)(nil)
