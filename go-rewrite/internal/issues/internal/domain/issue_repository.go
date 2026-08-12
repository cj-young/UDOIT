package domain

import "context"

type IssueRepository interface {
	DeleteByContentItemIDs(ctx context.Context, contentItemIDs []int64) error
	CreateMany(ctx context.Context, issues []*Issue) error
	GetByCourseID(ctx context.Context, courseID int64) ([]*Issue, error)
}