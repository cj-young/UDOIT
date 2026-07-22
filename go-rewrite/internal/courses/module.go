package courses

import (
	"context"
	"database/sql"

	"rewritetest/internal/courses/internal/application"
	"rewritetest/internal/courses/internal/infrastructure"
)

type Module struct {
	createCourseUseCase *application.CreateCourseUseCase
}

func New(db *sql.DB) *Module {

	courseRepository := infrastructure.NewMySQLCourseRepository(db)
	createCourseUseCase := application.NewCreateCourseUseCase(courseRepository)

	return &Module{
		createCourseUseCase: createCourseUseCase,
	}
}

func (m *Module) CreateCourse(ctx context.Context, title string, tenantID int64) (int64, error) {
	return m.createCourseUseCase.Execute(ctx, application.CreateCourseCommand{
		Title:    title,
		TenantID: tenantID,
	})
}