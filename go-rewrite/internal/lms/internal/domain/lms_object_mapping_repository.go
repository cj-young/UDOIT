package domain

import "context"

type LMSObjectMappingRepository interface {
	GetByTypeAndInternalID(ctx context.Context, objectType LMSObjectType, internalID int64) (*LMSObjectMapping, error)
}
