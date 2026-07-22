package infrastructure

import (
	"context"
	"database/sql"
	"rewritetest/internal/courses/internal/domain"
)

type MySQLCourseRepository struct {
	db *sql.DB
}

func NewMySQLCourseRepository(db *sql.DB) *MySQLCourseRepository {
	return &MySQLCourseRepository{
		db: db,
	}
}

func (r *MySQLCourseRepository) Create(ctx context.Context, course *domain.Course) (int64, error) {
	query := "INSERT INTO course (title, is_active, is_dirty, updated_at) VALUES (?, ?, ?, ?)"
	result, err := r.db.ExecContext(ctx, query, course.Title(), course.IsActive(), course.IsDirty(), course.UpdatedAt())
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
