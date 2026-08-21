package domain

import "time"

type File struct {
	id           int64
	courseID     int64
	fileName     string
	fileType     string
	updatedAt    time.Time
	isActive     bool
	isAvailable  bool
	isHidden     bool
	fileSize     int64
	downloadURL  string
	reviewerID   int64
	reviewedOn   time.Time
	isReviewed   bool
	externalID   string
	externalData map[string]any
}

func NewFile(courseID int64, fileName, fileType string, updatedAt time.Time, isActive, isAvailable, isHidden bool, fileSize int64, downloadURL string, reviewerID int64, reviewedOn time.Time, isReviewed bool, externalID string, externalData map[string]any) *File {
	return &File{
		courseID:     courseID,
		fileName:     fileName,
		fileType:     fileType,
		updatedAt:    updatedAt,
		isActive:     isActive,
		isAvailable:  isAvailable,
		isHidden:     isHidden,
		fileSize:     fileSize,
		downloadURL:  downloadURL,
		reviewerID:   reviewerID,
		reviewedOn:   reviewedOn,
		isReviewed:   isReviewed,
		externalID:   externalID,
		externalData: externalData,
	}
}

func RehydrateFile(id, courseID int64, fileName, fileType string, updatedAt time.Time, isActive, isAvailable, isHidden bool, fileSize int64, downloadURL string, reviewerID int64, reviewedOn time.Time, isReviewed bool, externalID string, externalData map[string]any) *File {
	return &File{
		id:           id,
		courseID:     courseID,
		fileName:     fileName,
		fileType:     fileType,
		updatedAt:    updatedAt,
		isActive:     isActive,
		isAvailable:  isAvailable,
		isHidden:     isHidden,
		fileSize:     fileSize,
		downloadURL:  downloadURL,
		reviewerID:   reviewerID,
		reviewedOn:   reviewedOn,
		isReviewed:   isReviewed,
		externalID:   externalID,
		externalData: externalData,
	}
}

func (f *File) ID() int64 {
	return f.id
}

func (f *File) CourseID() int64 {
	return f.courseID
}

func (f *File) FileName() string {
	return f.fileName
}

func (f *File) FileType() string {
	return f.fileType
}

func (f *File) UpdatedAt() time.Time {
	return f.updatedAt
}

func (f *File) IsActive() bool {
	return f.isActive
}

func (f *File) IsAvailable() bool {
	return f.isAvailable
}

func (f *File) IsHidden() bool {
	return f.isHidden
}

func (f *File) FileSize() int64 {
	return f.fileSize
}

func (f *File) DownloadURL() string {
	return f.downloadURL
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
