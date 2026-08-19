package canvas

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
)

type canvasContentType string

const (
	canvasContentTypePage            canvasContentType = "page"
	canvasContentTypeAssignment      canvasContentType = "assignment"
	canvasContentTypeAnnouncement    canvasContentType = "announcement"
	canvasContentTypeDiscussionTopic canvasContentType = "discussion_topic"
	canvasContentTypeSyllabus        canvasContentType = "syllabus"
)

func ParseCanvasContentType(s string) (canvasContentType, error) {
	switch s {
	case string(canvasContentTypePage):
		return canvasContentTypePage, nil
	case string(canvasContentTypeAssignment):
		return canvasContentTypeAssignment, nil
	case string(canvasContentTypeAnnouncement):
		return canvasContentTypeAnnouncement, nil
	case string(canvasContentTypeDiscussionTopic):
		return canvasContentTypeDiscussionTopic, nil
	case string(canvasContentTypeSyllabus):
		return canvasContentTypeSyllabus, nil
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
	assignments, err := p.getAssignments(ctx, canvasCourse.courseID, userID)
	if err != nil {
		return nil, err
	}
 
	discussionTopics, err := p.getDiscussionTopics(ctx, canvasCourse.courseID, userID)
	if err != nil {
		return nil, err
	}
	announcements, err := p.getAnnouncements(ctx, canvasCourse.courseID, userID)
	if err != nil {
		return nil, err
	}

	syllabus, err := p.getSyllabus(ctx, canvasCourse.courseID, userID)
	if err != nil {
		return nil, err
	}

	totalContent := len(pages) + len(assignments) + len(discussionTopics) + len(announcements)
	if syllabus != nil {
		totalContent++
	}

	courseContents := make([]domain.CourseContent, 0, totalContent)
	for _, page := range pages {
		courseContents = append(courseContents, domain.CourseContent{
			ExternalID: "page:" + page.URL, // TODO: evaluate whether this is the best external ID
			ExternalData: map[string]any{
				"content_id":   page.PageID,
				"updated_at":   page.UpdatedAt,
				"content_type": string(canvasContentTypePage),
			},
			HTML: page.Body,
			Type: domain.CourseContentTypePage,
		})
	}

	for _, assignment := range assignments {
		assignmentIDStr := strconv.FormatInt(assignment.ID, 10)
		courseContents = append(courseContents, domain.CourseContent{
			ExternalID: "assignment:" + assignmentIDStr,
			ExternalData: map[string]any{
				"content_id":   assignment.ID,
				"updated_at":   assignment.UpdatedAt,
				"content_type": string(canvasContentTypeAssignment),
			},
			HTML: assignment.Description,
			Type: domain.CourseContentTypeAssignment,
		})
	}

	for _, discussionTopic := range discussionTopics {
		discussionTopicIDStr := strconv.FormatInt(discussionTopic.ID, 10)
		courseContents = append(courseContents, domain.CourseContent{
			ExternalID: "discussion_topic:" + discussionTopicIDStr,
			ExternalData: map[string]any{
				"content_id":   discussionTopic.ID,
				"updated_at":   discussionTopic.UpdatedAt,
				"content_type": string(canvasContentTypeDiscussionTopic),
			},
			HTML: discussionTopic.Message,
			Type: domain.CourseContentTypeDiscussionTopic,
		})
	}

	for _, announcement := range announcements {
		announcementIDStr := strconv.FormatInt(announcement.ID, 10)
		courseContents = append(courseContents, domain.CourseContent{
			ExternalID: "announcement:" + announcementIDStr,
			ExternalData: map[string]any{
				"content_id":   announcement.ID,
				"updated_at":   announcement.UpdatedAt,
				"content_type": string(canvasContentTypeAnnouncement),
			},
			HTML: announcement.Message,
			Type: domain.CourseContentTypeAnnouncement,
		})
	}

	if syllabus != nil {
		syllabusIDStr := strconv.FormatInt(syllabus.ID, 10)
		courseContents = append(courseContents, domain.CourseContent{
			ExternalID: "syllabus:" + syllabusIDStr,
			ExternalData: map[string]any{
				"content_id":   syllabus.ID,
				"updated_at":   time.Time{}.Format(time.RFC3339),
				"content_type": string(canvasContentTypeSyllabus),
			},
			HTML: syllabus.SyllabusBody,
			Type: domain.CourseContentTypeSyllabus,
		})
	}

	return courseContents, nil
}

type canvasContent struct {
	contentID   string
	contentType canvasContentType
	updatedAt   time.Time
}

func (p *CanvasLMSProvider) asCanvasContent(content domain.LMSContent) (canvasContent, error) {
	var contentID string
	switch v := content.ExternalData["content_id"].(type) {
	case string:
		contentID = v
	case int:
		contentID = strconv.Itoa(v)
	case int32:
		contentID = strconv.FormatInt(int64(v), 10)
	case int64:
		contentID = strconv.FormatInt(v, 10)
	case float64:
		if v != float64(int64(v)) {
			return canvasContent{}, apperr.Internal("Missing or invalid content ID in content mapping data")
		}
		contentID = strconv.FormatInt(int64(v), 10)
	default:
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

	updatedAtStr, ok := content.ExternalData["updated_at"].(string)
	if !ok {
		return canvasContent{}, apperr.Internal("Missing or invalid updated timestamp in content mapping data")
	}

	var updatedAt time.Time
	updatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return canvasContent{}, apperr.Internal("Invalid updated timestamp in content mapping data")
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

type AssignmentResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updated_at"`
}

type SyllabusResponse struct {
	ID           int64  `json:"id"`
	SyllabusBody string `json:"syllabus_body"`
}

type DiscussionTopicResponse struct {
	ID             int64  `json:"id"`
	Message        string `json:"message"`
	UpdatedAt      string
	LastReplyAt    string `json:"last_reply_at"`
	PostedAt       string `json:"posted_at"`
	IsAnnouncement bool   `json:"is_announcement"`
}

type AnnouncementResponse struct {
	ID          int64  `json:"id"`
	Message     string `json:"message"`
	UpdatedAt   string
	LastReplyAt string `json:"last_reply_at"`
	PostedAt    string `json:"posted_at"`
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

func (p *CanvasLMSProvider) getAssignments(ctx context.Context, canvasCourseID string, userID int64) ([]AssignmentResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "/assignments?per_page=100"
	url := ""

	var assignments []AssignmentResponse

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

			var newAssignments []AssignmentResponse
			if err := json.NewDecoder(resp.Body).Decode(&newAssignments); err != nil {
				return apperr.Internal("Failed to decode response body")
			}
			assignments = append(assignments, newAssignments...)

			url = p.getNextPageLink(resp)
			path = ""
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return assignments, nil
}

func (p *CanvasLMSProvider) getDiscussionTopics(ctx context.Context, canvasCourseID string, userID int64) ([]DiscussionTopicResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "/discussion_topics?per_page=100"
	url := ""

	var discussionTopics []DiscussionTopicResponse

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

			var pageTopics []DiscussionTopicResponse
			if err := json.NewDecoder(resp.Body).Decode(&pageTopics); err != nil {
				return apperr.Internal("Failed to decode response body")
			}

			for _, topic := range pageTopics {
				if topic.IsAnnouncement {
					continue
				}

				updatedAt := topic.LastReplyAt
				if updatedAt == "" {
					updatedAt = topic.PostedAt
				}
				if updatedAt == "" {
					updatedAt = time.Time{}.Format(time.RFC3339)
				}

				topic.UpdatedAt = updatedAt
				discussionTopics = append(discussionTopics, topic)
			}

			url = p.getNextPageLink(resp)
			path = ""
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return discussionTopics, nil
}

func (p *CanvasLMSProvider) getAnnouncements(ctx context.Context, canvasCourseID string, userID int64) ([]AnnouncementResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "/discussion_topics?only_announcements=true&per_page=100"
	url := ""

	var announcements []AnnouncementResponse

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

			var pageAnnouncements []AnnouncementResponse
			if err := json.NewDecoder(resp.Body).Decode(&pageAnnouncements); err != nil {
				return apperr.Internal("Failed to decode response body")
			}

			for _, announcement := range pageAnnouncements {
				updatedAt := announcement.LastReplyAt
				if updatedAt == "" {
					updatedAt = announcement.PostedAt
				}
				if updatedAt == "" {
					updatedAt = time.Time{}.Format(time.RFC3339)
				}

				announcement.UpdatedAt = updatedAt
				announcements = append(announcements, announcement)
			}

			url = p.getNextPageLink(resp)
			path = ""
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return announcements, nil
}

func (p *CanvasLMSProvider) getSyllabus(ctx context.Context, canvasCourseID string, userID int64) (*SyllabusResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "?include[]=syllabus_body"

	resp, err := p.doAuthenticatedRequest(ctx, CanvasRequest{
		Path:   path,
		Body:   nil,
		Method: http.MethodGet,
		Config: p.config,
		UserID: userID,
	})
	if err != nil {
		slog.Info("failed to send request", "path", path, "error", err)
		return nil, apperr.Internal("Failed to send HTTP request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apperr.Internal("Unexpected status code: " + resp.Status)
	}

	var syllabus SyllabusResponse
	if err := json.NewDecoder(resp.Body).Decode(&syllabus); err != nil {
		return nil, apperr.Internal("Failed to decode response body")
	}

	if strings.TrimSpace(syllabus.SyllabusBody) == "" {
		return nil, nil
	}

	return &syllabus, nil
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
