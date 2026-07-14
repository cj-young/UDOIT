package domain

import "time"

type LTISession struct {
	id            int
	state         string
	nonce         string
	issuer        string
	clientID      string
	targetLinkURI string
	createdAt     time.Time
	expiresAt     time.Time
}

func NewLTISession(state, nonce, issuer, clientID, targetLinkURI string, createdAt, expiresAt time.Time) *LTISession {
	return &LTISession{
		state:         state,
		nonce:         nonce,
		issuer:        issuer,
		clientID:      clientID,
		targetLinkURI: targetLinkURI,
		createdAt:     createdAt,
		expiresAt:     expiresAt,
	}
}

func RehydrateLTISession(id int, state, nonce, issuer, clientID, targetLinkURI string, createdAt, expiresAt time.Time) *LTISession {
	return &LTISession{
		id:            id,
		state:         state,
		nonce:         nonce,
		issuer:        issuer,
		clientID:      clientID,
		targetLinkURI: targetLinkURI,
		createdAt:     createdAt,
		expiresAt:     expiresAt,
	}
}

func (l *LTISession) State() string {
	return l.state
}

func (l *LTISession) Nonce() string {
	return l.nonce
}

func (l *LTISession) Issuer() string {
	return l.issuer
}

func (l *LTISession) ClientID() string {
	return l.clientID
}

func (l *LTISession) IsValidState(state string) bool {
	return l.state == state
}

func (l *LTISession) IsValidNonce(nonce string) bool {
	return l.nonce == nonce
}

func (l *LTISession) TargetLinkURI() string {
	return l.targetLinkURI
}

func (l *LTISession) IsExpired() bool {
	return time.Now().After(l.expiresAt)
}
