package infrastructure

import (
	"context"
	"database/sql"

	"rewritetest/internal/scanner/internal/domain"
	scannersqlc "rewritetest/internal/scanner/internal/infrastructure/sqlc"
)

type MySQLReportRepository struct {
	queries *scannersqlc.Queries
}

func NewMySQLReportRepository(db *sql.DB) *MySQLReportRepository {
	return &MySQLReportRepository{
		queries: scannersqlc.New(db),
	}
}

func (r *MySQLReportRepository) Create(ctx context.Context, report *domain.Report) error {
	return r.queries.CreateReport(ctx, scannersqlc.CreateReportParams{
		ID:              uint64(report.ID()),
		CourseID:        uint64(report.CourseID()),
		ErrorCount:      uint32(report.ErrorCount()),
		SuggestionCount: uint32(report.SuggestionCount()),
		FileCount:       uint32(report.FileCount()),
		ScannedBy:       uint64(report.ScannedBy()),
		ContentFixed:    uint64(report.ContentFixed()),
		ContentResolved: uint64(report.ContentResolved()),
	})
}

var _ domain.ReportRepository = (*MySQLReportRepository)(nil)
