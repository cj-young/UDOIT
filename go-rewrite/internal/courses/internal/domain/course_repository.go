package domain

import "context"

type CourseRepository interface {
	Create(ctx context.Context, course *Course) (int64, error)
}