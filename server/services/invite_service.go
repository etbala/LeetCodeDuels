package services

import (
	"context"
	"fmt"
	"leetcodeduels/models"

	"github.com/go-redis/redis/v8"
)

var InviteManager *inviteManager

type inviteManager struct {
	client *redis.Client
	ctx    context.Context
}

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

// Creates invite to a duel with another player. Expires after 3 min.
// Returns whether invite was created successfully or not. (Fails if inviter
// has already sent another invite)
func (i *inviteManager) CreateInvite(inviterID int64, inviteeID int64, matchDetails models.MatchDetails) (bool, error) {
	return true, nil
}

// Remove invite (for when invite is accepted/declined).
func (i *inviteManager) RemoveInvite(inviterID int64) (bool, error) {
	return true, nil
}

// Checks if user has sent an invite that has been responded to or expired.
func (i *inviteManager) IsAwaitingResponse(inviterID int64) (bool, error) {
	return true, nil
}

func (i *inviteManager) Close() error {
	return i.client.Close()
}
