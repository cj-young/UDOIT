package canvas

import (
	"context"
	"fmt"
	"net/http"

	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
	"rewritetest/internal/shared/auth"
)


type CanvasFile struct {
	fileID string
	contextType string
}


// Deletes a file from Canvas
// https://developerdocs.instructure.com/services/canvas/resources/files#method.files.destroy
func (p *CanvasLMSProvider) DeleteFile(ctx context.Context, principal auth.Principal, config domain.LMSProviderConfig, fileMapping domain.LMSObjectMapping) error {

	canvasConfig, err := p.asCanvasConfig(config)
	if err != nil {
		return err
	}

	canvasFile, err := p.asCanvasFile(fileMapping)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/files/%s", canvasConfig.baseURL, canvasFile.fileID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return apperr.New(
			apperr.CodeInternal, "Failed to create HTTP request", err.Error(),
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.DeleteFile"),
			apperr.WithCause(err),
		)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.doAuthenticatedRequest(ctx, req, canvasConfig, principal.AgentID)
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
