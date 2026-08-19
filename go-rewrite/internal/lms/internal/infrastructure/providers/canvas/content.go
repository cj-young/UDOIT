package canvas

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
)

type canvasContentType string

const (
	canvasContentTypePage canvasContentType = "page"
)

func ParseCanvasContentType(s string) (canvasContentType, error) {
	switch s {
	case string(canvasContentTypePage):
		return canvasContentTypePage, nil
	default:
		return "", apperr.Internal("Invalid content type")
	}
}

func (p canvasContentType) String() string {
	return string(p)
}

func (p *CanvasLMSProvider) GetContent(
	ctx context.Context,
	course domain.LMSCourse,
	currentContent []domain.LMSContent,
	userID int64,
) ([]domain.CourseContent, error) {
	canvasCourse, err := p.asCanvasCourse(course)
	if err != nil {
		return nil, err
	}

	// TODO: use current content mappings to skip fetching content that is already up to date.
	pages, err := p.getPages(ctx, canvasCourse.courseID, userID)
	if err != nil {
		return nil, err
	}

	courseContents := make([]domain.CourseContent, len(pages))
	for i, page := range pages {
		courseContents[i] = domain.CourseContent{
			ExternalID: page.URL, // TODO: evaluate whether this is the best external ID
			ExternalData: map[string]any{
				"content_id":   page.PageID,
				"updated_at":   page.UpdatedAt,
				"content_type": string(canvasContentTypePage),
			},
			HTML: page.Body,
			Type: domain.CourseContentTypePage,
		}
	}

	return courseContents, nil
}

type canvasContent struct {
	contentID   string
	contentType canvasContentType
	updatedAt   time.Time
}

func (p *CanvasLMSProvider) asCanvasContent(content domain.LMSContent) (canvasContent, error) {
	contentID, ok := content.ExternalData["content_id"].(string)
	if !ok {
		return canvasContent{}, apperr.Internal("Missing or invalid content ID in content mapping data")
	}

	contentTypeStr, ok := content.ExternalData["content_type"].(string)
	if !ok {
		return canvasContent{}, apperr.Internal("Missing or invalid content type in content mapping data")
	}
	contentType, err := ParseCanvasContentType(contentTypeStr)
	if err != nil {
		return canvasContent{}, apperr.Internal("Invalid content type in content mapping data")
	}

	updatedAt, ok := content.ExternalData["updated_at"].(time.Time)
	if !ok {
		return canvasContent{}, apperr.Internal("Missing or invalid updated timestamp in content mapping data")
	}

	return canvasContent{
		contentID:   contentID,
		contentType: contentType,
		updatedAt:   updatedAt,
	}, nil
}

type PageResponse struct {
	PageID    int    `json:"page_id"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
}

func (p *CanvasLMSProvider) getPages(ctx context.Context, canvasCourseID string, userID int64) ([]PageResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "/pages?include[]=body&per_page=100"
	url := ""

	var pages []PageResponse

	for url != "" || path != "" {
		err := func() error {
			resp, err := p.doAuthenticatedRequest(ctx, CanvasRequest{
				Path:   path,
				URL:    url,
				Body:   nil,
				Method: http.MethodGet,
				Config: p.config,
				UserID: userID,
			})
			if err != nil {
				slog.Info("failed to send request", "url", url, "error", err)
				return apperr.Internal("Failed to send HTTP request")
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return apperr.Internal("Unexpected status code: " + resp.Status)
			}

			var newPages []PageResponse
			if err := json.NewDecoder(resp.Body).Decode(&newPages); err != nil {
				return apperr.Internal("Failed to decode response body")
			}
			pages = append(pages, newPages...)

			url = p.getNextPageLink(resp)
			path = "" // Clear the path since we are now using the URL for the next request
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return pages, nil
}

// Pagination in Canvas works through a "Link" header.
// https://developerdocs.instructure.com/services/canvas/basics/file.pagination
func (p *CanvasLMSProvider) getNextPageLink(resp *http.Response) string {
	link := resp.Header.Get("Link")

	for part := range strings.SplitSeq(link, ",") {
		if strings.Contains(part, `rel="next"`) {
			start := strings.Index(part, "<")
			end := strings.Index(part, ">")
			if start != -1 && end != -1 && start < end {
				return part[start+1 : end]
			}
		}
	}
	return ""
}
