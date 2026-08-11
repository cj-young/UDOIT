package canvas

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
	"strings"
)

func (p *CanvasLMSProvider) GetContent(
	ctx context.Context,
	tenantConfig domain.LMSProviderConfig,
	course domain.LMSCourse,
	currentContent []domain.LMSContent,
	userID int64,
) ([]domain.CourseContent, error) {
	
	canvasConfig, err := p.asCanvasConfig(tenantConfig)
	if err != nil {
		return nil, err
	}

	canvasCourse, err := p.asCanvasCourse(course)
	if err != nil {
		return nil, err
	}

	// TODO: use current content mappings to skip fetching content that is already up to date.
	pages, err := p.getPages(ctx, canvasConfig, canvasCourse.courseID, userID)
	if err != nil {
		return nil, err
	}

	courseContents := make([]domain.CourseContent, len(pages))
	for i, page := range pages {
		courseContents[i] = domain.CourseContent{
			ExternalID:	page.URL, // TODO: evaluate whether this is the best external ID
			HTML:				page.Body,
			Type:				domain.CourseContentTypePage,	
		}
	}

	return courseContents, nil
}



type PageResponse struct {
	PageID 			int `json:"page_id"`
	URL 				string `json:"url"`
	CreatedAt 	string `json:"created_at"`
	UpdatedAt 	string `json:"updated_at"`
	Body 				string `json:"body"`
}

func (p *CanvasLMSProvider) getPages(ctx context.Context, config canvasConfig, canvasCourseID string, userID int64) ([]PageResponse, error) {

	path := "/api/v1/courses/"+canvasCourseID+"/pages?include[]=body&per_page=100"
	url := ""

	var pages []PageResponse


	for url != "" || path != "" {
		err := func () error {
			resp, err := p.doAuthenticatedRequest(ctx, CanvasRequest{
					Path:   path,
					URL:    url,
					Body:   nil,
					Method: http.MethodGet,
					Config: config,
					UserID: userID,
			})
			if err != nil {
				slog.Info("failed to send request", "url", url, "error", err)
				return apperr.New(apperr.CodeInternal, "failed_to_send_request", "Failed to send HTTP request")
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return apperr.New(apperr.CodeInternal, "unexpected_status_code", "Unexpected status code: "+resp.Status)
			}

			var newPages []PageResponse
			if err := json.NewDecoder(resp.Body).Decode(&newPages); err != nil {
				return apperr.New(apperr.CodeInternal, "failed_to_decode_response", "Failed to decode response body")
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