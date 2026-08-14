package domain

type CourseContent struct {
	ExternalID string
	HTML       string
	Type       CourseContentType
}

type CourseContentType string

const (
	CourseContentTypePage CourseContentType = "page"
)
