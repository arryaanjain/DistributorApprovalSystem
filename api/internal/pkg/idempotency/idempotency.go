// Package idempotency enforces idempotency for financial operations.
// A client supplies an Idempotency-Key header; the server records the first
// response and replays it for duplicate requests within the TTL.
package idempotency

import (
	"context"
	"encoding/json"
	"time"
)

const DefaultTTL = 24 * time.Hour

// Store is the interface that the idempotency layer uses to persist key→result mappings.
// Implementations may use Redis or PostgreSQL.
type Store interface {
	// Get returns the cached response for a key, or (nil, nil) if not found.
	Get(ctx context.Context, key string) (*CachedResponse, error)
	// Set stores the response for a key with the given TTL.
	Set(ctx context.Context, key string, resp *CachedResponse, ttl time.Duration) error
	// Lock acquires an exclusive lock for a key.
	// Returns (true, nil) if the lock was acquired, (false, nil) if already locked.
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	// Unlock releases the lock for a key.
	Unlock(ctx context.Context, key string) error
}

// CachedResponse is what we persist for a completed idempotent request.
type CachedResponse struct {
	StatusCode int             `json:"status_code"`
	Body       json.RawMessage `json:"body"`
	CreatedAt  time.Time       `json:"created_at"`
}
