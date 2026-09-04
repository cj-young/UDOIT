package application

import (
	"context"

	"rewritetest/internal/accessibility/internal/domain"
)

type MarkHtmlAsReviewedUseCase struct {
	HTMLissueRepository domain.HTMLIssueRepository
}

func NewHtmlIssueReviewedUseCase(HTMLissueRepository domain.HTMLIssueRepository) *MarkHtmlAsReviewedUseCase {
	return &MarkHtmlAsReviewedUseCase{HTMLissueRepository: HTMLissueRepository}
}

func (u *MarkHtmlAsReviewedUseCase) Execute(ctx context.Context, issueID int64) error {
	issue, err := u.HTMLissueRepository.GetByID(ctx, issueID)
	if err != nil {
		return err
	}

	issue.MarkAsReviewed()

	return u.HTMLissueRepository.Update(ctx, issue)
}
