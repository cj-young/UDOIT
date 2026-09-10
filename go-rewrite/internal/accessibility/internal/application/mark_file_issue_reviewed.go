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

func (u *MarkFileAsReviewedUseCase) Execute(ctx context.Context, fileIssueID int64, userID int64) error {
	issue, err := u.FileIssueRepository.GetByID(ctx, fileIssueID)
	if err != nil {
		return err
	}

	if err := issue.Review(userID, time.Now()); err != nil {
		return err
	}

	return u.FileIssueRepository.Update(ctx, issue)
}
