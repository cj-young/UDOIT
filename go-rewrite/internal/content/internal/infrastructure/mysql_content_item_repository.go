package infrastructure

import (
	"context"
	"database/sql"
	"rewritetest/internal/content/internal/domain"
	contentsqlc "rewritetest/internal/content/internal/infrastructure/sqlc"
)

type MySQLContentItemRepository struct {
	queries *contentsqlc.Queries
	db      *sql.DB
}

func NewMySQLContentItemRepository(db *sql.DB) *MySQLContentItemRepository {
	return &MySQLContentItemRepository{
		queries: contentsqlc.New(db),
		db:      db,
	}
}

func (r *MySQLContentItemRepository) GetByID(ctx context.Context, id int64) (*domain.ContentItem, error) {
	row, err := r.queries.GetContentItemByID(ctx, uint64(id))
	if err != nil {
		return nil, err
	}

	contentItem := domain.RehydrateContentItem(
		int64(row.ID),
		int64(row.CourseID),
		row.ContentHash,
		row.ExternalID,
		row.CreatedAt,
		row.UpdatedAt,
	)
	return contentItem, nil
}
func (r *MySQLContentItemRepository) GetByCourseID(ctx context.Context, courseID int64) ([]*domain.ContentItem, error) {
	rows, err := r.queries.GetContentItemsByCourseID(ctx, uint64(courseID))
	if err != nil {
		return nil, err
	}

	contentItems := make([]*domain.ContentItem, 0, len(rows))
	for _, row := range rows {
		contentItem := domain.RehydrateContentItem(
			int64(row.ID),
			int64(row.CourseID),
			row.ContentHash,
			row.ExternalID,
			row.CreatedAt,
			row.UpdatedAt,
		)
		contentItems = append(contentItems, contentItem)
	}

	return contentItems, nil
}

func (r *MySQLContentItemRepository) Create(ctx context.Context, contentItem *domain.ContentItem) error {
	return r.queries.CreateContentItem(ctx, contentsqlc.CreateContentItemParams{
		CourseID:    uint64(contentItem.CourseID()),
		ContentHash: contentItem.ContentHash(),
		ExternalID:  contentItem.ExternalID(),
	})
}

func (r *MySQLContentItemRepository) CreateMany(ctx context.Context, contentItems []*domain.ContentItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	qtx := r.queries.WithTx(tx)

	domainContentItemsParams := make([]contentsqlc.CreateContentItemParams, len(contentItems))
	for i, item := range contentItems {
		domainContentItemsParams[i] = contentsqlc.CreateContentItemParams{
			CourseID:    uint64(item.CourseID()),
			ContentHash: item.ContentHash(),
			ExternalID:  item.ExternalID(),
		}
	}

	for _, params := range domainContentItemsParams {
		err := qtx.CreateContentItem(ctx, params)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ domain.ContentItemRepository = (*MySQLContentItemRepository)(nil)