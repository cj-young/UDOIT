package domain

import "context"

type TenantRepository interface {
	Create(ctx context.Context, lmsKey string) (int64, error)
}
