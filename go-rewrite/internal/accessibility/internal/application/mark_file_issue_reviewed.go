package application

import (
	"context"
	"rewritetest/internal/accessibility/internal/domain"
	"time"
)

type MarkFileAsReviewedUseCase struct {
	FileIssueRepository domain.FileIssueRepository
}

func NewFileissueReviewedUseCase(FileIssueRepository domain.FileIssueRepository) *MarkFileAsReviewedUseCase {
	return &MarkFileAsReviewedUseCase{FileIssueRepository: FileIssueRepository}
}

func (u *MarkFileAsReviewedUseCase) Execute(ctx context.Context, fileIssueID int64) error {
	issue, err := u.FileIssueRepository.GetByID(ctx, fileIssueID)
	if err != nil {
		return err
	}

	// TODO: Properly handle reviewerID and remove hardcoded val
	if err := issue.Review(1, time.Now()); err != nil {
		return err
	}

	return u.FileIssueRepository.Update(ctx, issue)
}
