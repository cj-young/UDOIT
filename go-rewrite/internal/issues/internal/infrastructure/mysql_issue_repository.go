package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"rewritetest/internal/issues/internal/domain"
	issuessqlc "rewritetest/internal/issues/internal/infrastructure/sqlc"
)

type MySQLIssueRepository struct {
	db      *sql.DB
	queries *issuessqlc.Queries
}

func NewMySQLIssueRepository(db *sql.DB) *MySQLIssueRepository {
	return &MySQLIssueRepository{
		db:      db,
		queries: issuessqlc.New(db),
	}
}

func (r *MySQLIssueRepository) DeleteByContentItemIDs(ctx context.Context, contentItemIDs []int64) error {
	uintContentItemIDs := make([]uint64, len(contentItemIDs))
	for i, v := range contentItemIDs {
		uintContentItemIDs[i] = uint64(v)
	}

	return r.queries.DeleteIssuesByContentItemIDs(ctx, uintContentItemIDs)
}

func (r *MySQLIssueRepository) CreateMany(ctx context.Context, issues []*domain.Issue) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	domainIssuesParams := make([]issuessqlc.CreateIssueParams, len(issues))
	for i, issue := range issues {
		var nullableFixedby sql.NullInt64
		if issue.FixedBy() != 0 {
			nullableFixedby = sql.NullInt64{
				Int64: issue.FixedBy(),
				Valid: true,
			}
		} else {
			nullableFixedby = sql.NullInt64{Valid: false}
		}

		detailsJSON, err := json.Marshal(issue.Details())
		if err != nil {
			return err
		}

		domainIssuesParams[i] = issuessqlc.CreateIssueParams{
			ContentItemID: uint64(issue.ContentItemID()),
			ScanRule:      issue.ScanRule().String(),
			ContentXpath:  issue.ContentXPath(),
			Status:        issue.Status().String(),
			Severity:      issue.Severity().String(),
			FixedBy:       nullableFixedby,
			FixedAt:       nullTime(issue.FixedAt()),
			Details:       detailsJSON,
		}
	}

	for _, params := range domainIssuesParams {
		err := qtx.CreateIssue(ctx, params)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *MySQLIssueRepository) GetByCourseID(ctx context.Context, courseID int64) ([]*domain.Issue, error) {
	issues, err := r.queries.GetIssuesByCourseID(ctx, uint64(courseID))
	if err != nil {
		return nil, err
	}

	domainIssues := make([]*domain.Issue, len(issues))
	for i, issue := range issues {
		scanRule, err := domain.ParseScanRule(issue.ScanRule)
		if err != nil {
			return nil, err
		}

		issueStatus, err := domain.ParseIssueStatus(issue.Status)
		if err != nil {
			return nil, err
		}

		issueSeverity, err := domain.ParseIssueSeverity(issue.Severity)
		if err != nil {
			return nil, err
		}

		var detailsMap map[string]any
		if err := json.Unmarshal(issue.Details, &detailsMap); err != nil {
			return nil, err
		}

		domainIssues[i] = domain.RehydrateIssue(
			int64(issue.ID),
			int64(issue.ContentItemID),
			scanRule,
			issue.ContentXpath,
			issueStatus,
			issueSeverity,
			int64(issue.FixedBy.Int64),
			issue.FixedAt.Time,
			detailsMap,
			issue.CreatedAt,
			issue.UpdatedAt,
		)
	}

	return domainIssues, nil
}

func nullTime(t time.Time) sql.NullTime {
	if !t.IsZero() {
		return sql.NullTime{
			Time:  t,
			Valid: true,
		}
	}
	return sql.NullTime{Valid: false}
}

var _ domain.IssueRepository = (*MySQLIssueRepository)(nil)
