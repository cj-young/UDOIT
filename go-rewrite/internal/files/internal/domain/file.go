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
}

func NewFile(id, courseID, reviewerID int64, reviewedOn time.Time, isReviewed bool) *File {
	return &File{
		id: id,
		courseID: courseID,
		reviewerID: reviewerID,
		reviewedOn: reviewedOn,
		isReviewed: isReviewed,
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