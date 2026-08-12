package issues

import (
	"context"
	"database/sql"
	"rewritetest/internal/issues/internal/domain"
	"rewritetest/internal/issues/internal/infrastructure"
	"rewritetest/internal/shared/apperr"
	"time"
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

type Issue struct {
	ContentItemID int64
	ScanRule			string
	Status   			string
	Severity 			string
	ContentXPath	string
	Details				map[string]any
	FixedBy				int64
	FixedOn				time.Time
}

// RegisterNewIssues removes all persisted issues for the specified contentItem
// IDs and creates the issues specified in newIssues.
func (m *Module) RegisterNewIssues(ctx context.Context, newIssues []Issue, contentItemIDs []int64) error {
	m.issueRepository.DeleteByContentItemIDs(ctx, contentItemIDs)
	
	issues := make([]*domain.Issue, len(newIssues))
	for i, newIssue := range newIssues {

		scanRule := domain.ScanRule(newIssue.ScanRule)
		if !scanRule.IsValid() {
			return apperr.New(apperr.CodeInternal, "invalid_scan_rule", "Scan rule "+newIssue.ScanRule+" is invalid")
		}

		issueSeverity, err := domain.ParseIssueSeverity(newIssue.Severity)
		if err != nil {
			return err
		}

		issueStatus, err := domain.ParseIssueStatus(newIssue.Status)
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

func (m *Module) GetByCourseID(ctx context.Context, courseID int64) ([]Issue, error) {
	domainIssues, err := m.issueRepository.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	issues := make([]Issue, len(domainIssues))
	for i, domainIssue := range domainIssues {
		issues[i] = Issue{
			ContentItemID: domainIssue.ContentItemID(),
			ScanRule:     	domainIssue.ScanRule().String(),
			Status:       	domainIssue.Status().String(),
			Severity:     	domainIssue.Severity().String(),
			ContentXPath: 	domainIssue.ContentXPath(),
			Details:      	domainIssue.Details(),
			FixedBy:       	domainIssue.FixedBy(),
			FixedOn:       	domainIssue.FixedOn(),
		}
	}
	return issues, nil
}