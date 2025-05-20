package ws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
)

var ConnManager *connManager

type connManager struct {
	client *redis.Client
	ctx    context.Context
}

// redisURL should be in the format "redis://<user>:<pass>@<host>:<port>/<db>"
func InitConnManager(redisURL string) error {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("invalid redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	ConnManager = &connManager{
		client: client,
		ctx:    context.Background(),
	}
	return nil
}

// AddConnection atomically registers a new connection for this user.
// It returns the oldConnID (if any) so that your WS handler can
// immediately terminate that socket.
func (c *connManager) AddConnection(userID int64, connID string) (oldConnID string, err error) {
	key := fmt.Sprintf("connection:%d", userID)
	oldConnID, err = c.client.GetSet(c.ctx, key, connID).Result()
	if err == redis.Nil {
		// no previous connection
		oldConnID = ""
		err = nil
	} else if err != nil {
		return "", err
	}
	// mark this user online
	if err := c.client.SAdd(c.ctx, "online_users", strconv.FormatInt(userID, 10)).Err(); err != nil {
		return oldConnID, err
	}
	return oldConnID, nil
}

// RemoveConnection removes the given connID for the user—**but only if it
// matches the current active connID**.  Returns (stillOnline, error).
// If this was the active connection, the user goes fully offline.
func (c *connManager) RemoveConnection(userID int64, connID string) (stillOnline bool, err error) {
	key := fmt.Sprintf("connection:%d", userID)

	current, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		// no active connection recorded
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if current != connID {
		// this socket wasn't the active one—ignore
		return true, nil
	}

	// it is the active one, so delete and mark offline
	if err := c.client.Del(c.ctx, key).Err(); err != nil {
		return false, err
	}
	if err := c.client.SRem(c.ctx, "online_users", strconv.FormatInt(userID, 10)).Err(); err != nil {
		return false, err
	}
	return false, nil
}

// GetConnection fetches the single active connID for a user (or "" if offline).
func (c *connManager) GetConnection(userID int64) (string, error) {
	key := fmt.Sprintf("connection:%d", userID)
	connID, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return connID, err
}

// IsOnline checks if a user has any active connections.
func (c *connManager) IsOnline(userID int64) (bool, error) {
	return c.client.SIsMember(c.ctx, "online_users", strconv.FormatInt(userID, 10)).Result()
}

// AllOnlineUsers returns the list of userIDs currently online.
func (c *connManager) AllOnlineUsers() ([]int64, error) {
	strs, err := c.client.SMembers(c.ctx, "online_users").Result()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(strs))
	for _, s := range strs {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid userID in online_users: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// OnlineCount returns the number of users currently online.
func (c *connManager) OnlineCount() (int64, error) {
	return c.client.SCard(c.ctx, "online_users").Result()
}

// Close shuts down the Redis client.
func (c *connManager) Close() error {
	return c.client.Close()
}
