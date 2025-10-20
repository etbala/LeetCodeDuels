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

const (
	inviteKeyPrefix  = "invite:"
	inviterSetPrefix = "invites:sent:"
	inviteeSetPrefix = "invites:received:"
)

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

// Generates a unique invite key
func generateInviteKey(inviterID, inviteeID int64) string {
	return fmt.Sprintf("%s%d:%d", inviteKeyPrefix, inviterID, inviteeID)
}

// Stores a new invite with a 3-minute TTL, fails if one already exists
func (im *inviteManager) CreateInvite(inviterID, inviteeID int64, matchDetails models.MatchDetails) (bool, error) {
	inviterKey := inviterSetPrefix + strconv.FormatInt(inviterID, 10)

	inviterCount, err := im.client.SCard(im.ctx, inviterKey).Result()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("redis check inviter error: %w", err)
	}

	if inviterCount > 0 {
		return false, nil // Inviter already has an outgoing invite
	}

	inviteKey := generateInviteKey(inviterID, inviteeID)
	payload := models.Invite{
		InviterID:    inviterID,
		InviteeID:    inviteeID,
		MatchDetails: matchDetails,
		CreatedAt:    time.Now(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	pipe := im.client.TxPipeline()

	pipe.Set(im.ctx, inviteKey, data, 3*time.Minute)

	pipe.SAdd(im.ctx, inviterKey, inviteKey)
	pipe.Expire(im.ctx, inviterKey, 3*time.Minute)

	inviteeKey := inviteeSetPrefix + strconv.FormatInt(inviteeID, 10)
	pipe.SAdd(im.ctx, inviteeKey, inviteKey)
	pipe.Expire(im.ctx, inviteeKey, 3*time.Minute)

	_, err = pipe.Exec(im.ctx)
	if err != nil {
		return false, fmt.Errorf("failed to store invite: %w", err)
	}

	return true, nil
}

func (im *inviteManager) GetPendingInvites(inviteeID int64) ([]models.Invite, error) {
	inviteeKey := inviteeSetPrefix + strconv.FormatInt(inviteeID, 10)

	inviteKeys, err := im.client.SMembers(im.ctx, inviteeKey).Result()
	if err == redis.Nil {
		return []models.Invite{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("redis get invite keys failed: %w", err)
	}

	var invites []models.Invite

	if len(inviteKeys) == 0 {
		return invites, nil
	}

	pipe := im.client.Pipeline()
	inviteDataCmds := make([]*redis.StringCmd, len(inviteKeys))

	for i, key := range inviteKeys {
		inviteDataCmds[i] = pipe.Get(im.ctx, key)
	}

	_, err = pipe.Exec(im.ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	for _, cmd := range inviteDataCmds {
		data, err := cmd.Result()
		if err == redis.Nil {
			continue // Key might have expired since we got the list
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get invite data: %w", err)
		}

		var invite models.Invite
		if err := json.Unmarshal([]byte(data), &invite); err != nil {
			return nil, fmt.Errorf("failed to unmarshal invite: %w", err)
		}
		invites = append(invites, invite)
	}

	return invites, nil
}

func (im *inviteManager) InviteDetails(inviterID int64) (*models.Invite, error) {
	inviterKey := inviterSetPrefix + strconv.FormatInt(inviterID, 10)
	inviteKeys, err := im.client.SMembers(im.ctx, inviterKey).Result()
	if err == redis.Nil || len(inviteKeys) == 0 {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("redis get inviter keys failed: %w", err)
	}

	inviteKey := inviteKeys[0] // One invite per inviter

	data, err := im.client.Get(im.ctx, inviteKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	var invite models.Invite
	err = json.Unmarshal([]byte(data), &invite)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal invite: %w", err)
	}

	return &invite, nil
}

// Deletes an existing invite; returns true if one was removed
func (im *inviteManager) RemoveInvite(inviterID int64) (bool, error) {
	inviterKey := inviterSetPrefix + strconv.FormatInt(inviterID, 10)
	inviteKeys, err := im.client.SMembers(im.ctx, inviterKey).Result()
	if err == redis.Nil || len(inviteKeys) == 0 {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("redis get inviter keys failed: %w", err)
	}

	inviteKey := inviteKeys[0] // One invite per inviter

	data, err := im.client.Get(im.ctx, inviteKey).Result()
	if err == redis.Nil {
		im.client.Del(im.ctx, inviterKey)
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("redis get failed: %w", err)
	}

	var invite models.Invite
	err = json.Unmarshal([]byte(data), &invite)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal invite: %w", err)
	}

	inviteeKey := inviteeSetPrefix + strconv.FormatInt(invite.InviteeID, 10)

	pipe := im.client.Pipeline()
	pipe.Del(im.ctx, inviteKey)
	pipe.SRem(im.ctx, inviterKey, inviteKey)
	pipe.SRem(im.ctx, inviteeKey, inviteKey)

	_, err = pipe.Exec(im.ctx)
	if err != nil {
		return false, fmt.Errorf("failed to remove invite: %w", err)
	}

	return true, nil
}

// Checks if an invite by this inviter still exists (im.e., awaiting response)
func (im *inviteManager) IsAwaitingResponse(inviterID int64) (bool, error) {
	inviterKey := inviterSetPrefix + strconv.FormatInt(inviterID, 10)
	count, err := im.client.SCard(im.ctx, inviterKey).Result()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("redis check failed: %w", err)
	}
	return count > 0, nil
}

// Shuts down the Redis client for invites
func (im *inviteManager) Close() error {
	return im.client.Close()
}
