package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Sessions is a server-side session store backed by Redis. The cookie carries only an
// opaque session id; the subject (tenant id, or "tenantID:customerID" for portal) lives
// in Redis with a TTL, so sessions are revocable and nothing sensitive is in the cookie.
type Sessions struct {
	rdb *redis.Client
}

func NewSessions(rdb *redis.Client) *Sessions {
	return &Sessions{rdb: rdb}
}

func key(kind, sid string) string { return fmt.Sprintf("sess:%s:%s", kind, sid) }

// Create stores subject under a fresh session id with the given TTL and returns the id.
func (s *Sessions) Create(ctx context.Context, kind, subject string, ttl time.Duration) (string, error) {
	sid := RandomToken()
	if err := s.rdb.Set(ctx, key(kind, sid), subject, ttl).Err(); err != nil {
		return "", err
	}
	return sid, nil
}

// Get returns the subject for a session id, or an error if missing/expired.
func (s *Sessions) Get(ctx context.Context, kind, sid string) (string, error) {
	return s.rdb.Get(ctx, key(kind, sid)).Result()
}

// Delete revokes a session.
func (s *Sessions) Delete(ctx context.Context, kind, sid string) error {
	return s.rdb.Del(ctx, key(kind, sid)).Err()
}
