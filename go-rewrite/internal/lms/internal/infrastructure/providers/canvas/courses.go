package canvas

import (
	"rewritetest/internal/lms/internal/domain"
	"rewritetest/internal/shared/apperr"
)


type canvasCourse struct {
	courseID string
}


func (p *CanvasLMSProvider) asCanvasCourse(course domain.LMSCourse) (canvasCourse, error) {

	courseID, ok := course.ExternalData["course_id"].(string)
	if !ok {
		return canvasCourse{}, apperr.New(apperr.CodeInternal, "invalid_canvas_course", "Missing or invalid 'course_id' in course mapping data")
	}

	return canvasCourse {
		courseID:  courseID,
	}, nil
}