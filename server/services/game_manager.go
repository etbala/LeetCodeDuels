package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"leetcodeduels/models"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

var GameManager *gameManager

type gameManager struct {
	client *redis.Client
	ctx    context.Context
}

const (
	gameKeyPrefix       = "game:"
	playerGameKeyPrefix = "player_game:"
)

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

// Returns the session for the given ID (or nil if not found)
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

// Returns session ID associated with given playerID. Returns empty string if no associated session.
func (gm *gameManager) GetSessionIDByPlayer(playerID int64) (string, error) {
	key := playerGameKeyPrefix + strconv.FormatInt(playerID, 10)
	sessionID, err := gm.client.Get(gm.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	return sessionID, nil
}

func (gm *gameManager) IsPlayerInGame(playerID int64) (bool, error) {
	sid, err := gm.GetSessionIDByPlayer(playerID)
	return sid != "", err
}

func (gm *gameManager) GetOpponent(sessionID string, userID int64) (int64, error) {
	session, err := gm.GetGame(sessionID)
	if err != nil {
		return 0, err
	}
	if session == nil {
		return 0, errors.New("No session associated with provided userID")
	}
	for i := 0; i < len(session.Players); i++ {
		if session.Players[i] != userID {
			return session.Players[i], nil
		}
	}
	return 0, errors.New("Unknown error retrieving opponent")
}

// Creates a new session, stores it in Redis, and returns its ID
func (gm *gameManager) StartGame(players []int64, problem models.Problem) (string, error) {
	sessionID := uuid.NewString()
	session := &models.Session{
		ID:          sessionID,
		Status:      models.MatchActive,
		IsRated:     false,
		Problem:     problem,
		Players:     players,
		Submissions: make([]models.PlayerSubmission, 0),
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
	for _, pid := range players {
		key := playerGameKeyPrefix + strconv.FormatInt(pid, 10)
		if err := gm.client.Set(gm.ctx, key, sessionID, 0).Err(); err != nil {
			fmt.Printf("warning: could not set player_game for %d: %v\n", pid, err)
		}
	}
	return sessionID, nil
}

// Appends a player's submission to the session and updates Redis
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
	session.Submissions = append(session.Submissions, submission)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return gm.client.Set(gm.ctx, gameKeyPrefix+sessionID, data, 0).Err()
}

// Marks the session as completed, sets its end time, and expires it in 3 minutes
func (gm *gameManager) CompleteGame(sessionID string, winnerID int64) error {
	session, err := gm.GetGame(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	session.Status = models.MatchWon
	session.EndTime = time.Now()
	session.Winner = winnerID // TODO: Verify winner is in session.Players
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	err = gm.client.Set(gm.ctx, gameKeyPrefix+sessionID, data, 3*time.Minute).Err()
	if err != nil {

	}

	for _, pid := range session.Players {
		key := playerGameKeyPrefix + strconv.FormatInt(pid, 10)
		_ = gm.client.Del(gm.ctx, key).Err()
	}
	return nil
}

// Marks the session as canceled, sets its end time, and expires it in 3 minutes
func (gm *gameManager) CancelGame(sessionID string) error {
	session, err := gm.GetGame(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	session.Status = models.MatchCanceled
	session.EndTime = time.Now()
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	err = gm.client.Set(gm.ctx, gameKeyPrefix+sessionID, data, 3*time.Minute).Err()
	if err != nil {
		return err
	}

	for _, pid := range session.Players {
		key := playerGameKeyPrefix + strconv.FormatInt(pid, 10)
		_ = gm.client.Del(gm.ctx, key).Err()
	}
	return nil
}

func (i *gameManager) Close() error {
	return i.client.Close()
}
