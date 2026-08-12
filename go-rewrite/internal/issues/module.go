package issues

import (
	"context"
	"database/sql"
	"rewritetest/internal/issues/internal/domain"
	"rewritetest/internal/issues/internal/infrastructure"
	"rewritetest/internal/shared/apperr"
)

type Module struct {
	issueRepository domain.IssueRepository
}

func New(db *sql.DB) *Module {

	issueRepository := infrastructure.NewMySQLIssueRepository(db)
	
	return &Module{
		issueRepository: issueRepository,
	}
}

type NewIssue struct {
	ContentItemID int64
	ScanRule			string
	IssueStatus   string
	IssueSeverity string
	ContentXPath	string
	Preview				string
	Details				map[string]any
}
// RegisterNewIssues removes all persisted issues for the specified contentItem
// IDs and creates the issues specified in newIssues.
func (m *Module) RegisterNewIssues(ctx context.Context, newIssues []NewIssue, contentItemIDs []int64) error {
	m.issueRepository.DeleteByContentItemIDs(ctx, contentItemIDs)
	
	issues := make([]*domain.Issue, len(newIssues))
	for i, newIssue := range newIssues {

		scanRule := domain.ScanRule(newIssue.ScanRule)
		if !scanRule.IsValid() {
			return apperr.New(apperr.CodeInternal, "invalid_scan_rule", "Scan rule "+newIssue.ScanRule+" is invalid")
		}

		issueSeverity, err := domain.ParseIssueSeverity(newIssue.IssueSeverity)
		if err != nil {
			return err
		}

		issueStatus, err := domain.ParseIssueStatus(newIssue.IssueStatus)
		if err != nil {
			return err
		}
		
		issues[i] = domain.NewIssue(
			newIssue.ContentItemID,
			scanRule,
			newIssue.ContentXPath,
			issueStatus,
			issueSeverity,
			newIssue.Details,
		)
	}
	
	err := m.issueRepository.CreateMany(ctx, issues)
	if err != nil {
		return apperr.New(
			apperr.CodeInternal, "create_many_failed", "Failed to create many issues",
			apperr.WithCause(err),
		)
	}
	return nil
}

func (m *Module) DeleteByContentItemIDs(ctx context.Context, contentItemIDs []int64) error {
	return m.issueRepository.DeleteByContentItemIDs(ctx, contentItemIDs)
}