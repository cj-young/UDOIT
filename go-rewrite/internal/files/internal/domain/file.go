package domain

import "time"

type File struct {
	id int64
	courseID int64
	// fileName string
	// fileType string
	// lmsURL string
	// downloadURL string
	reviewerID int64
	reviewedOn time.Time
	isReviewed bool
	// status bool
	// isActive bool
	// updatedAt time.Time
	externalID string
	externalData map[string]any
}

func NewFile(courseID, reviewerID int64, reviewedOn time.Time, isReviewed bool, externalID string, externalData map[string]any) *File {
	return &File{
		courseID: courseID,
		reviewerID: reviewerID,
		reviewedOn: reviewedOn,
		isReviewed: isReviewed,
		externalID: externalID,
		externalData: externalData,
	}
}

func RehydrateFile(id, courseID, reviewerID int64, reviewedOn time.Time, isReviewed bool, externalID string, externalData map[string]any) *File {
	return &File{
		id: id,
		courseID: courseID,
		reviewerID: reviewerID,
		reviewedOn: reviewedOn,
		isReviewed: isReviewed,
		externalID: externalID,
		externalData: externalData,
	}
}

func (f *File) ID() int64 {
	return f.id
}

func (f *File) CourseID() int64 {
	return f.courseID
}

func (f *File) ReviewerID() int64 {
	return f.reviewerID
}

func (f *File) ReviewedOn() time.Time {
	return f.reviewedOn
}

func (f *File) IsReviewed() bool {
	return f.isReviewed
}

func (f *File) ExternalID() string {
	return f.externalID
}

func (f *File) ExternalData() map[string]any {
	return f.externalData
}