package cache

import (
	"context"
	"time"
)

// Cache is the abstraction over Redis that your app uses.
type Cache interface {
	// CSRF state
	SetState(ctx context.Context, state string, ttl time.Duration) error
	VerifyAndDeleteState(ctx context.Context, state string) (bool, error)

	// TODO: Match Session & Websocket Conns

}
