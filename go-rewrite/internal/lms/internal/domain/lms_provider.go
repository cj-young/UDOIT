package domain

import (
	"context"
	"rewritetest/internal/shared/auth"
)

type LMSProvider interface {
	DeleteFile(ctx context.Context, principal auth.Principal, config LMSProviderConfig, fileMapping LMSObjectMapping) error
}
