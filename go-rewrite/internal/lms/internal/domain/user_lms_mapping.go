package domain

import "time"

type UserLMSMapping struct {
	userID         int64
	lmsKey         string
	externalUserID string
	apiDomain      string
	metadata       map[string]any
	createdAt      time.Time
	updatedAt      time.Time
}

func NewUserLMSMapping(userID int64, lmsKey, externalUserID, apiDomain string, metadata map[string]any, now time.Time) UserLMSMapping {
	if metadata == nil {
		metadata = map[string]any{}
	}

	return UserLMSMapping{
		userID:         userID,
		lmsKey:         lmsKey,
		externalUserID: externalUserID,
		apiDomain:      apiDomain,
		metadata:       metadata,
		createdAt:      now,
		updatedAt:      now,
	}
}

func RehydrateUserLMSMapping(userID int64, lmsKey, externalUserID, apiDomain string, metadata map[string]any, createdAt, updatedAt time.Time) UserLMSMapping {
	if metadata == nil {
		metadata = map[string]any{}
	}

	return UserLMSMapping{
		userID:         userID,
		lmsKey:         lmsKey,
		externalUserID: externalUserID,
		apiDomain:      apiDomain,
		metadata:       metadata,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (m UserLMSMapping) UserID() int64            { return m.userID }
func (m UserLMSMapping) LMSKey() string           { return m.lmsKey }
func (m UserLMSMapping) ExternalUserID() string   { return m.externalUserID }
func (m UserLMSMapping) APIDomain() string        { return m.apiDomain }
func (m UserLMSMapping) Metadata() map[string]any { return m.metadata }
