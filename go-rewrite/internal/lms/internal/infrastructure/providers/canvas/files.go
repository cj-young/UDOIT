package canvas

import (
	"context"
	"fmt"
	"net/http"

	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
	"rewritetest/internal/shared/auth"
)


type canvasFile struct {
	fileID string
	contextType string
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

// Deletes a file from Canvas
// https://developerdocs.instructure.com/services/canvas/resources/files#method.files.destroy
func (p *CanvasLMSProvider) DeleteFile(ctx context.Context, principal auth.Principal, file domain.LMSFile) error {
	canvasFile, err := p.asCanvasFile(file)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/files/%s", p.config.baseURL, canvasFile.fileID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return apperr.New(
			apperr.CodeInternal, "Failed to create HTTP request", err.Error(),
			apperr.WithOp("lms.infrastructure.canvas_lms_provider.DeleteFile"),
			apperr.WithCause(err),
		)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.doAuthenticatedRequest(ctx, CanvasRequest{
		Path:   url,
		Body:   nil,
		Method: http.MethodDelete,
		Config: p.config,
		UserID: principal.AgentID,
	})
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
