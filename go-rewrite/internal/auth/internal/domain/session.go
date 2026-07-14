package domain

import "time"

type Session struct {
	id        string
	userId    int64
	createdAt time.Time
	expiresAt time.Time
}

func NewSession(id string, userId int64, createdAt time.Time, expiresAt time.Time) Session {
	return Session{
		id:        id,
		userId:    userId,
		createdAt: createdAt,
		expiresAt: expiresAt,
	}
}

func (s Session) UserID() int64 {
	return s.userId
}

func (s Session) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}

func (s Session) ExpiresAt() time.Time {
	return s.expiresAt
}

func (s Session) ID() string {
	return s.id
}
