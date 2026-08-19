package domain

type CourseContent struct {
	ExternalID   string
	ExternalData map[string]any
	HTML         string
	Type         CourseContentType
}

type CourseContentType string

const (
	CourseContentTypePage       CourseContentType = "page"
	CourseContentTypeAssignment CourseContentType = "assignment"
	CourseContentTypeAnnouncement CourseContentType = "announcement"
	CourseContentTypeDiscussionTopic CourseContentType = "discussion_topic"
	CourseContentTypeSyllabus   CourseContentType = "syllabus"
)
