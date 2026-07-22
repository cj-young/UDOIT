package domain

import "context"

type UserLMSMappingRepository interface {
	Upsert(ctx context.Context, mapping UserLMSMapping) error
	GetByUserID(ctx context.Context, userID int64) (*UserLMSMapping, error)
}
