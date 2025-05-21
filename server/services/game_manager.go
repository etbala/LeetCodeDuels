package services

import (
	"context"
	"encoding/json"
	"fmt"
	"leetcodeduels/models"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

var GameManager *gameManager

type gameManager struct {
	client *redis.Client
	ctx    context.Context
}

const gameKeyPrefix = "game:"

func InitGameManager(redisURL string) error {
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

// GetGame returns the session for the given ID (or nil if not found)
func (gm *gameManager) GetGame(sessionID string) (*models.Session, error) {
	key := gameKeyPrefix + sessionID
	data, err := gm.client.Get(gm.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	var session models.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	return &session, nil
}

// StartGame creates a new session, stores it in Redis, and returns its ID
func (gm *gameManager) StartGame(players []int64, problem models.Problem) (string, error) {
	sessionID := uuid.NewString()
	session := &models.Session{
		ID:          sessionID,
		InProgress:  true,
		IsCanceled:  false,
		IsRated:     false,
		Problem:     problem,
		Players:     players,
		Submissions: make([][]models.PlayerSubmission, len(players)),
		Winner:      0,
		StartTime:   time.Now(),
	}
	data, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}
	if err := gm.client.Set(gm.ctx, gameKeyPrefix+sessionID, data, 0).Err(); err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}
	return sessionID, nil
}

// AddSubmission appends a player's submission to the session and updates Redis
func (gm *gameManager) AddSubmission(sessionID string, submission models.PlayerSubmission) error {
	session, err := gm.GetGame(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	idx := -1
	for i, pid := range session.Players {
		if pid == submission.PlayerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("player %d not part of session", submission.PlayerID)
	}
	session.Submissions[idx] = append(session.Submissions[idx], submission)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return gm.client.Set(gm.ctx, gameKeyPrefix+sessionID, data, 0).Err()
}

// CompleteGame marks the session as completed, sets its end time, and expires it in 3 minutes
func (gm *gameManager) CompleteGame(sessionID string) error {
	session, err := gm.GetGame(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	session.InProgress = false
	session.EndTime = time.Now()
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return gm.client.Set(gm.ctx, gameKeyPrefix+sessionID, data, 3*time.Minute).Err()
}

// CancelGame marks the session as canceled, sets its end time, and expires it in 3 minutes
func (gm *gameManager) CancelGame(sessionID string) error {
	session, err := gm.GetGame(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	session.InProgress = false
	session.IsCanceled = true
	session.EndTime = time.Now()
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return gm.client.Set(gm.ctx, gameKeyPrefix+sessionID, data, 3*time.Minute).Err()
}
