package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"rewritetest/internal/accessibility/internal/domain"
	accessibilitysqlc "rewritetest/internal/accessibility/internal/infrastructure/sqlc"
)

type MySQLFileIssueRepository struct {
	queries *accessibilitysqlc.Queries
}

func NewMySQLFileIssueRepository(db *sql.DB) *MySQLFileIssueRepository {
	return &MySQLFileIssueRepository{
		queries: accessibilitysqlc.New(db),
	}
}

func (r *MySQLFileIssueRepository) ReplaceForCourse(ctx context.Context, courseID int64, fileIDs []int64) error {
	err := r.queries.DeleteFileIssuesByCourseID(ctx, uint64(courseID))
	if err != nil {
		return err
	}

	for _, fileID := range fileIDs {
		err = r.queries.CreateFileIssue(ctx, uint64(fileID))
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *MySQLFileIssueRepository) GetByID(ctx context.Context, ID int64) (*domain.FileIssue, error) {
	issue, err := r.queries.GetFileIssueByID(ctx, uint64(ID))
	if err != nil {
		return nil, err
	}

	var reviewerID *int64
	var reviewedOn *time.Time

	if issue.ReviewerID.Valid {
		reviewerID = &issue.ReviewerID.Int64
	}
	if issue.ReviewedOn.Valid {
		reviewedOn = &issue.ReviewedOn.Time
	}

	return domain.RehydrateFileIssue(
		int64(issue.ID),
		int64(issue.FileID),
		reviewerID,
		reviewedOn,
		issue.CreatedAt,
		issue.UpdatedAt,
	), nil
}

func (r *MySQLFileIssueRepository) Update(ctx context.Context, issue *domain.FileIssue) error {
	return r.queries.MarkFileReviewed(
		ctx,
		accessibilitysqlc.MarkFileReviewedParams{
			ID:         uint64(issue.ID()),
			ReviewerID: sql.NullInt64{Int64: issue.ReviewerID(), Valid: true},
			ReviewedOn: sql.NullTime{Time: issue.ReviewedOn(), Valid: !issue.ReviewedOn().IsZero()},
			UpdatedAt:  issue.UpdatedAt(),
		},
	)
}

var _ domain.FileIssueRepository = (*MySQLFileIssueRepository)(nil)
