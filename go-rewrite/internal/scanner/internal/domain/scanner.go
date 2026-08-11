package domain

import (
	"context"
	"rewritetest/internal/lms"
)

type Scanner interface {
	ScanContent(ctx context.Context, items []lms.ContentItemDTO) error
}
