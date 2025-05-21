package services

import (
	"context"
	"encoding/json"
	"fmt"
	"leetcodeduels/models"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var InviteManager *inviteManager

type inviteManager struct {
	client *redis.Client
	ctx    context.Context
}

const inviteKeyPrefix = "invite:"

func InitInviteManager(redisURL string) error {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("invalid redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	InviteManager = &inviteManager{
		client: client,
		ctx:    context.Background(),
	}
	return nil
}

// Stores a new invite with a 3-minute TTL, fails if one already exists
func (i *inviteManager) CreateInvite(inviterID, inviteeID int64, matchDetails models.MatchDetails) (bool, error) {
	key := inviteKeyPrefix + strconv.FormatInt(inviterID, 10)
	// Ensure no existing invite
	if exists, err := i.client.Exists(i.ctx, key).Result(); err != nil {
		return false, err
	} else if exists != 0 {
		return false, nil
	}
	payload := models.InvitePayload{
		InviteeID:    inviteeID,
		MatchDetails: matchDetails,
		CreatedAt:    time.Now(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}
	// Store with TTL
	if err := i.client.Set(i.ctx, key, data, 3*time.Minute).Err(); err != nil {
		return false, err
	}
	return true, nil
}

// Deletes an existing invite; returns true if one was removed
func (i *inviteManager) RemoveInvite(inviterID int64) (bool, error) {
	key := inviteKeyPrefix + strconv.FormatInt(inviterID, 10)
	deleted, err := i.client.Del(i.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return deleted > 0, nil
}

// Checks if an invite key still exists (i.e., awaiting response)
func (i *inviteManager) IsAwaitingResponse(inviterID int64) (bool, error) {
	key := inviteKeyPrefix + strconv.FormatInt(inviterID, 10)
	exists, err := i.client.Exists(i.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// Shuts down the Redis client for invites
func (i *inviteManager) Close() error {
	return i.client.Close()
}
