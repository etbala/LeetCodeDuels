package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var StateStore *stateStore

type stateStore struct {
	client *redis.Client
	ctx    context.Context
}

// redisURL should be in the format "redis://<user>:<pass>@<host>:<port>/<db>"
func InitStateStore(redisURL string) error {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("invalid redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	StateStore = &stateStore{
		client: client,
		ctx:    context.Background(),
	}
	return nil
}

func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Stores state associated with provided context for 5 minutes
func (ss *stateStore) StoreState(ctx context.Context, state string) error {
	err := ss.client.Set(ctx, state, "1", 5*time.Minute).Err()
	return err
}

// Checks if state is valid for provided context. Deletes the state if it was valid.
func (ss *stateStore) ValidateState(ctx context.Context, state string) (bool, error) {
	existed, err := ss.client.Exists(ctx, state).Result()
	if err != nil {
		return false, err
	}
	return existed != 0, nil
}

func (ss *stateStore) Close() error {
	return ss.client.Close()
}
