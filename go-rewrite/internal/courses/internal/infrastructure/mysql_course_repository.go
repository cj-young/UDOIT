package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"rewritetest/internal/courses/internal/domain"
	coursesqlc "rewritetest/internal/courses/internal/infrastructure/sqlc"
)

type MySQLCourseRepository struct {
	db *sql.DB
	queries *coursesqlc.Queries
}

func NewMySQLCourseRepository(db *sql.DB) *MySQLCourseRepository {
	return &MySQLCourseRepository{
		db: db,
		queries: coursesqlc.New(db),
	}
}

func (r *MySQLCourseRepository) Create(ctx context.Context, course *domain.Course) (int64, error) {
	externalDataJSON, err := json.Marshal(course.ExternalData())
	if err != nil {
		return 0, err
	}

	result, err := r.queries.CreateCourse(ctx, coursesqlc.CreateCourseParams{
		Title:        course.Title(),
		IsActive:     course.IsActive(),
		IsDirty:      course.IsDirty(),
		ExternalID:   course.ExternalID(),
		ExternalData: externalDataJSON,
		UpdatedAt:    course.UpdatedAt(),
		TenantID:     uint64(course.TenantID()),
	})
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
