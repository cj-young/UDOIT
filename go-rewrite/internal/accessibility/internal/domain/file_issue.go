package domain

import (
	"fmt"
	"rewritetest/internal/shared/apperr"
	"time"
)

type review struct {
	reviewerID int64
	reviewedOn time.Time
}

type FileIssue struct {
	id        int64
	fileID    int64
	review    *review
	createdAt time.Time
	updatedAt time.Time
}

func NewFileIssue(fileID, reviewerID int64, reviewedOn time.Time) *FileIssue {
	return &FileIssue{
		fileID: fileID,
		review: &review{
			reviewerID: reviewerID,
			reviewedOn: reviewedOn,
		},
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}
}

func RehydrateFileIssue(
	id int64,
	fileID int64,
	reviewerID *int64,
	reviewedOn *time.Time,
	createdAt time.Time,
	updatedAt time.Time) *FileIssue {

	fileIssue := &FileIssue{
		id:        id,
		fileID:    fileID,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}

	// If neither review fields are nil then properly hydrate fields.
	// Otherwise dont do anything (avoids weird zero-value bug with IsReviewed())
	if reviewerID != nil && reviewedOn != nil {
		fileIssue.review = &review{
			reviewerID: *reviewerID,
			reviewedOn: *reviewedOn,
		}
	}

	return fileIssue
}

func (f *FileIssue) IsReviewed() bool {
	return f.review != nil
}

func (f *FileIssue) Review(reviewerID int64, reviewedOn time.Time) error {
	if f.IsReviewed() == true {
		fmt.Println("Printing review fields for debug")
		fmt.Println(f.review.reviewedOn)
		fmt.Println(f.review.reviewerID)
		return apperr.New(apperr.CodeInternal, "File issue has already been reviewed")
	} else {
		f.review = &review{
			reviewerID: reviewerID,
			reviewedOn: reviewedOn,
		}
		f.updatedAt = reviewedOn
	}
	return nil
}

func (f *FileIssue) ReviewedOn() time.Time {
	return f.review.reviewedOn
}

func (f *FileIssue) ReviewerID() int64 {
	return f.review.reviewerID
}

func (f *FileIssue) UpdatedAt() time.Time {
	return f.updatedAt
}

func (f *FileIssue) CreatedAt() time.Time {
	return f.createdAt
}

func (f *FileIssue) ID() int64 {
	return f.id
}
