package infrastructure

import (
	"context"
	"time"

	"rewritetest/internal/auth/internal/domain"

	"github.com/redis/go-redis/v9"
)

type RedisSessionRepository struct {
	client *redis.Client
	ttl    time.Duration
	prefix string
}

func NewRedisSessionRepository(client *redis.Client, ttl time.Duration, prefix string) *RedisSessionRepository {
	return &RedisSessionRepository{
		client: client,
		ttl:    ttl,
		prefix: prefix,
	}
}

func (r *RedisSessionRepository) key(id string) string {
	return r.prefix + id
}

func (r *RedisSessionRepository) Create(ctx context.Context, session domain.Session) error {
	expiration := r.ttl
	if !session.ExpiresAt().IsZero() {
		expiration = time.Until(session.ExpiresAt())
	}

	return r.client.Set(ctx, r.key(session.ID()), session, expiration).Err()
}

func (r *RedisSessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	var session domain.Session
	err := r.client.Get(ctx, r.key(id)).Scan(&session)
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}
	return &session, nil
}

func (r *RedisSessionRepository) DeleteByID(ctx context.Context, id string) error {
	return r.client.Del(ctx, r.key(id)).Err()
}
