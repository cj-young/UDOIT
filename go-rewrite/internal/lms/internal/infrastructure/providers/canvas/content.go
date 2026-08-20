package canvas

import (
	"context"
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
	canvasContentTypeQuiz            canvasContentType = "quiz"
	canvasContentTypeQuizQuestion    canvasContentType = "quiz_question"
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
	case string(canvasContentTypeQuiz):
		return canvasContentTypeQuiz, nil
	case string(canvasContentTypeQuizQuestion):
		return canvasContentTypeQuizQuestion, nil
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
	quizzes, err := p.getQuizzes(ctx, canvasCourse.courseID, userID)
	if err != nil {
		return nil, err
	}

	syllabus, err := p.getSyllabus(ctx, canvasCourse.courseID, userID)
	if err != nil {
		return nil, err
	}

	quizQuestionsCount := 0
	for _, quiz := range quizzes {
		quizQuestionsCount += len(quiz.Questions)
	}

	totalContent := len(pages) + len(assignments) + len(discussionTopics) + len(announcements) + len(quizzes) + quizQuestionsCount
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

	for _, quiz := range quizzes {
		quizIDStr := strconv.FormatInt(quiz.ID, 10)
		courseContents = append(courseContents, domain.CourseContent{
			ExternalID: "quiz:" + quizIDStr,
			ExternalData: map[string]any{
				"content_id":   quiz.ID,
				"updated_at":   quiz.UpdatedAt,
				"content_type": string(canvasContentTypeQuiz),
			},
			HTML: quiz.Description,
			Type: domain.CourseContentTypeQuiz,
		})

		for _, question := range quiz.Questions {
			// Include quiz ID because Canvas question IDs may only be unique within a quiz.
			questionIDStr := strconv.FormatInt(question.ID, 10)
			quizIDStr := strconv.FormatInt(quiz.ID, 10)
			courseContents = append(courseContents, domain.CourseContent{
				ExternalID: "quiz_question:" + quizIDStr + ":" + questionIDStr,
				ExternalData: map[string]any{
					"content_id":   question.ID,
					"updated_at":   question.UpdatedAt,
					"content_type": string(canvasContentTypeQuizQuestion),
					"quiz_id":      quiz.ID,
				},
				HTML: question.QuestionText,
				Type: domain.CourseContentTypeQuizQuestion,
			})
		}
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
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	UpdatedAt       string `json:"updated_at"`
	QuizID          *int64 `json:"quiz_id"`
	DiscussionTopic any    `json:"discussion_topic"`
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

type QuizResponse struct {
	ID          int64  `json:"id"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updated_at"`
	Questions   []QuizQuestionResponse
}

type QuizQuestionResponse struct {
	ID           int64  `json:"id"`
	QuestionText string `json:"question_text"`
	UpdatedAt    string
}

func (p *CanvasLMSProvider) getPages(ctx context.Context, canvasCourseID string, userID int64) ([]PageResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "/pages?include[]=body&per_page=100"
	return fetchPaginated[PageResponse](p, ctx, userID, path, nil)
}

func (p *CanvasLMSProvider) getAssignments(ctx context.Context, canvasCourseID string, userID int64) ([]AssignmentResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "/assignments?per_page=100"
	return fetchPaginated[AssignmentResponse](p, ctx, userID, path, func(assignment AssignmentResponse) (AssignmentResponse, bool) {
		// Skip quizzes and discussion topics, since they will be fetched elsewhere.
		
		if assignment.QuizID != nil {
			return AssignmentResponse{}, false
		}

		if assignment.DiscussionTopic != nil {
			return AssignmentResponse{}, false
		}

		return assignment, true
	})
}

func (p *CanvasLMSProvider) getDiscussionTopics(ctx context.Context, canvasCourseID string, userID int64) ([]DiscussionTopicResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "/discussion_topics?per_page=100"
	return fetchPaginated[DiscussionTopicResponse](p, ctx, userID, path, func(topic DiscussionTopicResponse) (DiscussionTopicResponse, bool) {
		if topic.IsAnnouncement {
			return DiscussionTopicResponse{}, false
		}

		topic.UpdatedAt = normalizeCanvasUpdatedAt(topic.LastReplyAt, topic.PostedAt)
		return topic, true
	})
}

func (p *CanvasLMSProvider) getAnnouncements(ctx context.Context, canvasCourseID string, userID int64) ([]AnnouncementResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "/discussion_topics?only_announcements=true&per_page=100"
	return fetchPaginated[AnnouncementResponse](p, ctx, userID, path, func(announcement AnnouncementResponse) (AnnouncementResponse, bool) {
		announcement.UpdatedAt = normalizeCanvasUpdatedAt(announcement.LastReplyAt, announcement.PostedAt)
		return announcement, true
	})
}

func (p *CanvasLMSProvider) getQuizzes(ctx context.Context, canvasCourseID string, userID int64) ([]QuizResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "/quizzes?per_page=100"

	quizzes, err := fetchPaginated[QuizResponse](p, ctx, userID, path, func(quiz QuizResponse) (QuizResponse, bool) {
		quiz.UpdatedAt = normalizeCanvasUpdatedAt(quiz.UpdatedAt, "")
		return quiz, true
	})
	if err != nil {
		return nil, err
	}

	for i, quiz := range quizzes {
		questions, err := p.getQuizQuestions(ctx, canvasCourseID, quiz.ID, userID)
		if err != nil {
			return nil, err
		}
		quizzes[i].Questions = questions
	}

	return quizzes, nil
}

func (p *CanvasLMSProvider) getQuizQuestions(ctx context.Context, canvasCourseID string, quizID int64, userID int64) ([]QuizQuestionResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "/quizzes/" + strconv.FormatInt(quizID, 10) + "/questions?per_page=100"

	questions, err := fetchPaginated[QuizQuestionResponse](p, ctx, userID, path, func(question QuizQuestionResponse) (QuizQuestionResponse, bool) {
		question.UpdatedAt = time.Time{}.Format(time.RFC3339)
		return question, true
	})
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (p *CanvasLMSProvider) getSyllabus(ctx context.Context, canvasCourseID string, userID int64) (*SyllabusResponse, error) {
	path := "/api/v1/courses/" + canvasCourseID + "?include[]=syllabus_body"

	syllabus, err := fetchOne[SyllabusResponse](p, ctx, userID, path)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(syllabus.SyllabusBody) == "" {
		return nil, nil
	}

	return &syllabus, nil
}

func normalizeCanvasUpdatedAt(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	if fallback != "" {
		return fallback
	}
	return time.Time{}.Format(time.RFC3339)
}
