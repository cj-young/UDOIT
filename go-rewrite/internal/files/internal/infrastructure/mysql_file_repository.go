package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"rewritetest/internal/files/internal/domain"
	"time"
)

type MySQLFileRepository struct {
	db *sql.DB
}

func NewMySQLFileRepository(db *sql.DB) *MySQLFileRepository {
	return &MySQLFileRepository{
		db: db,
	}
}

func (r *MySQLFileRepository) GetFileByID(ctx context.Context, fileID int64) (*domain.File, error) {
	query := `
		SELECT id, course_id, reviewed_by_id, reviewed_on, reviewed, external_id, external_data
		FROM file_item
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, fileID)

	var (
		courseID int64
		reviewerID sql.NullInt64
		reviewedOn sql.NullTime
		isReviewed bool
		externalID sql.NullString
		externalDataStr sql.NullString
	)
	err := row.Scan(&fileID, &courseID, &reviewerID, &reviewedOn, &isReviewed, &externalID, &externalDataStr)
	if err != nil {
		return nil, err
	}

	externalData := make(map[string]any)
	if externalDataStr.Valid && len(externalDataStr.String) > 0 {
		if err := json.Unmarshal([]byte(externalDataStr.String), &externalData); err != nil {
			return nil, err
		}
	}

	file := domain.RehydrateFile(
		fileID,
		courseID,
		reviewerID.Int64,
		reviewedOn.Time,
		isReviewed,
		externalID.String,
		externalData,
	)

	return file, nil
}

func (r *MySQLFileRepository) UpdateFile(ctx context.Context, file *domain.File) error {
	query := `
		UPDATE file_item
		SET course_id = ?, reviewed_by_id = ?, reviewed_on = ?, reviewed = ?, external_id = ?, external_data = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, file.CourseID(), file.ReviewerID(), time.Now(), file.IsReviewed(), file.ExternalID(), file.ExternalData(), file.ID())
	return err
}

func (r *MySQLFileRepository) DeleteFile(ctx context.Context, fileID int64) error {
	query := `
		DELETE FROM file_item
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, fileID)
	return err
}

var _ domain.FileRepository = (*MySQLFileRepository)(nil)