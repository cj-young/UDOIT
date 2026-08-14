package application

import (
	"context"

	"rewritetest/internal/issues"
	"rewritetest/internal/scanner/internal/domain"
)

type CreateReportUseCase struct {
	reportRepository domain.ReportRepository
	issueRetriever   IssueRetriever
}

type IssueRetriever interface {
	GetByCourseID(ctx context.Context, courseID int64) ([]issues.Issue, error)
}

func NewCreateReportUseCase(reportRepository domain.ReportRepository, issueRetriever IssueRetriever) *CreateReportUseCase {
	return &CreateReportUseCase{
		reportRepository: reportRepository,
		issueRetriever:   issueRetriever,
	}
}

type CreateReportCommand struct {
	UserID   int64
	CourseID int64
}

func (u *CreateReportUseCase) Execute(ctx context.Context, cmd CreateReportCommand) error {
	// TODO: add user-based authorization

	issues, err := u.issueRetriever.GetByCourseID(ctx, cmd.CourseID)
	if err != nil {
		return err
	}

	var errorCount, suggestionCount, fixedCount, resolvedCount, fileCount int

	for _, issue := range issues {
		switch issue.Status {
		case "fixed":
			fixedCount++
		case "marked_as_reviewed":
			resolvedCount++
		case "active":
			if issue.Severity == "error" {
				errorCount++
			} else {
				suggestionCount++
			}
		}
	}

	report := domain.NewReport(
		cmd.CourseID,
		errorCount,
		suggestionCount,
		fileCount,
		cmd.UserID,
		fixedCount,
		resolvedCount,
	)

	err = u.reportRepository.Create(ctx, report)
	if err != nil {
		return err
	}

	return nil
}
