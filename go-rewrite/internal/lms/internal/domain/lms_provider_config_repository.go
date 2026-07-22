package domain

import "context"

type LMSProviderConfigRepository interface {
	GetByTenant(ctx context.Context, tenantID int64) (*LMSProviderConfig, error)
}