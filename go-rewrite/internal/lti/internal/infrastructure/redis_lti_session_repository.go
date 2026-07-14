package infrastructure

import (
	"context"
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

func (r *RedisLTISessionRepository) Create(ctx context.Context, session *domain.LTISession) error {
	return r.client.Set(ctx, r.key(session.State()), session, r.ttl).Err()
}

func (r *RedisLTISessionRepository) GetByState(ctx context.Context, state string) (*domain.LTISession, error) {
	var session domain.LTISession
	err := r.client.Get(ctx, r.key(state)).Scan(&session)
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}
	return &session, nil
}

func (r *RedisLTISessionRepository) Delete(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, r.key(sessionID)).Err()
}

var _ domain.LTISessionRepository = (*RedisLTISessionRepository)(nil)
