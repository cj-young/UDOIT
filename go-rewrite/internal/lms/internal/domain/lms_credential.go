package domain

import "time"

type LMSCredential struct {
	userID    int64
	lmsKey    string
	payload   map[string]any
	expiresAt *time.Time
	active    bool
	createdAt time.Time
	updatedAt time.Time
}

func NewLMSCredential(userID int64, lmsKey string, payload map[string]any, expiresAt *time.Time, now time.Time) LMSCredential {
	if payload == nil {
		payload = map[string]any{}
	}

	return LMSCredential{
		userID:    userID,
		lmsKey:    lmsKey,
		payload:   payload,
		expiresAt: expiresAt,
		active:    true,
		createdAt: now,
		updatedAt: now,
	}
}

func RehydrateLMSCredential(userID int64, lmsKey string, payload map[string]any, expiresAt *time.Time, active bool, createdAt, updatedAt time.Time) LMSCredential {
	if payload == nil {
		payload = map[string]any{}
	}

	return LMSCredential{
		userID:    userID,
		lmsKey:    lmsKey,
		payload:   payload,
		expiresAt: expiresAt,
		active:    active,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (c LMSCredential) UserID() int64           { return c.userID }
func (c LMSCredential) LMSKey() string          { return c.lmsKey }
func (c LMSCredential) Payload() map[string]any { return c.payload }
func (c LMSCredential) ExpiresAt() *time.Time   { return c.expiresAt }
func (c LMSCredential) Active() bool            { return c.active }
