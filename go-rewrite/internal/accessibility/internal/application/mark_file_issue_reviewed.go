package application

import (
	"context"
	"rewritetest/internal/accessibility/internal/domain"
	"time"
)

type MarkFileAsReviewedUseCase struct {
	FileIssueRepository domain.FileIssueRepository
}

type MarkFileReviewedCommand struct {
	UserID      int64
	FileIssueID int64
}

func NewFileissueReviewedUseCase(FileIssueRepository domain.FileIssueRepository) *MarkFileAsReviewedUseCase {
	return &MarkFileAsReviewedUseCase{FileIssueRepository: FileIssueRepository}
}

func (u *MarkFileAsReviewedUseCase) Execute(ctx context.Context, cmd MarkFileReviewedCommand) error {
	issue, err := u.FileIssueRepository.GetByID(ctx, cmd.FileIssueID)
	if err != nil {
		return err
	}

	if err := issue.Review(cmd.UserID, time.Now()); err != nil {
		return err
	}

	return u.FileIssueRepository.Update(ctx, issue)
}
