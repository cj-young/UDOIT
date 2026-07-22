package infrastructure

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	"rewritetest/internal/lti/internal/domain"

	"github.com/redis/go-redis/v9"
)

type RedisLTISessionRepository struct {
	client *redis.Client
	ttl    time.Duration
	prefix string
}

func NewRedisLTISessionRepository(client *redis.Client, ttl time.Duration, prefix string) *RedisLTISessionRepository {
	return &RedisLTISessionRepository{
		client: client,
		ttl:    ttl,
		prefix: prefix,
	}
}

func (r *RedisLTISessionRepository) key(state string) string {
	return r.prefix + state
}

type LTISessionStorageDTO struct {
	State string
	Nonce string
	Issuer string
	ClientID string
	TargetLinkURI string
	TenantID int64
	CreatedAt time.Time
	ExpiresAt time.Time
}


func (r *RedisLTISessionRepository) Create(ctx context.Context, session *domain.LTISession) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(ToLTISessionStorageDTO(session))
	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.key(session.State()), buf.Bytes(), r.ttl).Err()
}

func (r *RedisLTISessionRepository) GetByState(ctx context.Context, state string) (*domain.LTISession, error) {
	data, err := r.client.Get(ctx, r.key(state)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}

	dec := gob.NewDecoder(bytes.NewReader(data))

	var dto LTISessionStorageDTO
	err = dec.Decode(&dto)
	if err != nil {
		return nil, err
	}

	return FromLTISessionStorageDTO(&dto), nil
}

func (r *RedisLTISessionRepository) Delete(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, r.key(sessionID)).Err()
}

func ToLTISessionStorageDTO(session *domain.LTISession) *LTISessionStorageDTO {
	return &LTISessionStorageDTO{
		State:         session.State(),
		Nonce:         session.Nonce(),
		Issuer:        session.Issuer(),
		ClientID:      session.ClientID(),
		TargetLinkURI: session.TargetLinkURI(),
		TenantID:      session.TenantID(),
		CreatedAt:     session.CreatedAt(),
		ExpiresAt:     session.ExpiresAt(),
	}
}

func FromLTISessionStorageDTO(dto *LTISessionStorageDTO) *domain.LTISession {
	return domain.NewLTISession(
		dto.State,
		dto.Nonce,
		dto.Issuer,
		dto.ClientID,
		dto.TargetLinkURI,
		dto.TenantID,
		dto.CreatedAt,
		dto.ExpiresAt,
	)
}

var _ domain.LTISessionRepository = (*RedisLTISessionRepository)(nil)
