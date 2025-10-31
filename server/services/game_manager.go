package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"leetcodeduels/models"
	"slices"
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

type gameSession struct {
	ID        string `redis:"id"`
	Status    string `redis:"status"`
	IsRated   bool   `redis:"isRated"`
	Problem   string `redis:"problem"`
	Players   string `redis:"players"`
	Winner    int64  `redis:"winner"`
	StartTime string `redis:"startTime"`
	EndTime   string `redis:"endTime"`
}

const (
	gameKeyPrefix       = "game:"        // Hash containing session metadata
	playerGameKeyPrefix = "player_game:" // String mapping playerID -> sessionID
	submissionsSuffix   = ":submissions" // List appended to gameKey
)

func gameKey(sessionID string) string {
	return gameKeyPrefix + sessionID
}
func submissionsKey(sessionID string) string {
	return gameKeyPrefix + sessionID + submissionsSuffix
}
func playerGameKey(playerID int64) string {
	return playerGameKeyPrefix + strconv.FormatInt(playerID, 10)
}

func InitGameManager(redisURL string) error {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("invalid redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	GameManager = &gameManager{
		client: client,
		ctx:    context.Background(),
	}
	return nil
}

// Returns the session for the given ID (or nil if not found)
func (gm *gameManager) GetGame(sessionID string) (*models.Session, error) {
	key := gameKey(sessionID)
	subKey := submissionsKey(sessionID)

	var gs gameSession
	if err := gm.client.HGetAll(gm.ctx, key).Scan(&gs); err != nil {
		return nil, fmt.Errorf("redis hgetall/scan failed: %w", err)
	}
	if gs.ID == "" {
		return nil, nil // Not found
	}

	submissionsData, err := gm.client.LRange(gm.ctx, subKey, 0, -1).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("redis lrange failed: %w", err)
	}

	return gm.assembleSession(&gs, submissionsData)
}

// Returns sessionID associated with playerID, or empty string if no associated session.
func (gm *gameManager) GetSessionIDByPlayer(playerID int64) (string, error) {
	key := playerGameKey(playerID)
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

// Finds the opponent of a player in a given session.
func (gm *gameManager) GetOpponent(sessionID string, userID int64) (int64, error) {
	key := gameKey(sessionID)
	playersData, err := gm.client.HGet(gm.ctx, key, "players").Result()
	if err == redis.Nil {
		return 0, errors.New("no session associated with provided sessionID")
	} else if err != nil {
		return 0, fmt.Errorf("redis hget failed: %w", err)
	}

	var players []int64
	if err := json.Unmarshal([]byte(playersData), &players); err != nil {
		return 0, fmt.Errorf("failed to unmarshal players: %w", err)
	}

	for _, pid := range players {
		if pid != userID {
			return pid, nil
		}
	}

	return 0, errors.New("unknown error retrieving opponent")
}

// Creates a new session, stores it in Redis, and returns its ID.
func (gm *gameManager) StartGame(players []int64, problem models.Problem) (string, error) {
	sessionID := uuid.NewString()
	key := gameKey(sessionID)

	problemData, err := json.Marshal(problem)
	if err != nil {
		return "", fmt.Errorf("failed to marshal problem: %w", err)
	}
	playersData, err := json.Marshal(players)
	if err != nil {
		return "", fmt.Errorf("failed to marshal players: %w", err)
	}

	sessionMap := map[string]interface{}{
		"id":        sessionID,
		"status":    string(models.MatchActive),
		"isRated":   false,
		"problem":   string(problemData),
		"players":   string(playersData),
		"winner":    0,
		"startTime": time.Now().Format(time.RFC3339Nano),
		"endTime":   "",
	}

	if err := gm.client.HSet(gm.ctx, key, sessionMap).Err(); err != nil {
		return "", fmt.Errorf("failed to store session hash: %w", err)
	}

	for _, pid := range players {
		if err := gm.client.Set(gm.ctx, playerGameKey(pid), sessionID, 0).Err(); err != nil {
			fmt.Printf("warning: could not set player_game for %d: %v\n", pid, err)
		}
	}
	return sessionID, nil
}

// AddSubmission appends a player's submission to the session.
func (gm *gameManager) AddSubmission(sessionID string, submission models.PlayerSubmission) error {
	subKey := submissionsKey(sessionID)

	playersData, err := gm.client.HGet(gm.ctx, gameKey(sessionID), "players").Result()
	if err != nil {
		return fmt.Errorf("failed to get players for session: %w", err)
	}
	var players []int64
	if err := json.Unmarshal([]byte(playersData), &players); err != nil {
		return fmt.Errorf("failed to unmarshal players: %w", err)
	}

	if !slices.Contains(players, submission.PlayerID) {
		return errors.New("player is not a participant in this session")
	}

	data, err := json.Marshal(submission)
	if err != nil {
		return fmt.Errorf("failed to marshal submission: %w", err)
	}

	return gm.client.RPush(gm.ctx, subKey, data).Err()
}

// Mark session as completed and sets a 3-minute expiry.
func (gm *gameManager) CompleteGame(sessionID string, winnerID int64) (*models.Session, error) {
	return gm.finalizeGame(sessionID, models.MatchWon, winnerID, 3*time.Minute)
}

// Mark session as canceled and sets a 3-minute expiry.
func (gm *gameManager) CancelGame(sessionID string) (*models.Session, error) {
	return gm.finalizeGame(sessionID, models.MatchCanceled, 0, 3*time.Minute)
}

func (gm *gameManager) Close() error {
	return gm.client.Close()
}

// finalizeGame is a common helper for completing or canceling a game.
func (gm *gameManager) finalizeGame(sessionID string, status models.MatchStatus, winnerID int64, expiry time.Duration) (*models.Session, error) {
	key := gameKey(sessionID)
	subKey := submissionsKey(sessionID)

	playersData, err := gm.client.HGet(gm.ctx, key, "players").Result()
	if err != nil {
		fmt.Printf("warning: could not get players for game %s: %v\n", sessionID, err)
		// Proceed with finalization anyway
	}

	updates := map[string]interface{}{
		"status":  string(status),
		"winner":  winnerID,
		"endTime": time.Now().Format(time.RFC3339Nano),
	}
	if err := gm.client.HSet(gm.ctx, key, updates).Err(); err != nil {
		return nil, fmt.Errorf("failed to finalize game hash: %w", err)
	}

	_ = gm.client.Expire(gm.ctx, key, expiry).Err()
	_ = gm.client.Expire(gm.ctx, subKey, expiry).Err()

	if playersData != "" {
		var players []int64
		if json.Unmarshal([]byte(playersData), &players) == nil {
			for _, pid := range players {
				_ = gm.client.Del(gm.ctx, playerGameKey(pid)).Err()
			}
		}
	}

	return gm.GetGame(sessionID)
}

// Helper to build models.Session struct from raw data fetched from Redis
func (gm *gameManager) assembleSession(gs *gameSession, submissionsData []string) (*models.Session, error) {
	var session models.Session
	var err error

	session.ID = gs.ID
	session.Status, _ = models.ParseMatchStatus(gs.Status)
	session.IsRated = gs.IsRated
	session.Winner = gs.Winner

	if gs.StartTime != "" {
		session.StartTime, _ = time.Parse(time.RFC3339Nano, gs.StartTime)
	}
	if gs.EndTime != "" {
		session.EndTime, _ = time.Parse(time.RFC3339Nano, gs.EndTime)
	}

	if err = json.Unmarshal([]byte(gs.Problem), &session.Problem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal problem: %w", err)
	}
	if err = json.Unmarshal([]byte(gs.Players), &session.Players); err != nil {
		return nil, fmt.Errorf("failed to unmarshal players: %w", err)
	}

	session.Submissions = make([]models.PlayerSubmission, 0, len(submissionsData))
	for _, subData := range submissionsData {
		var sub models.PlayerSubmission
		if err := json.Unmarshal([]byte(subData), &sub); err == nil {
			session.Submissions = append(session.Submissions, sub)
		} else {
			fmt.Printf("warning: could not unmarshal submission: %v\n", err)
		}
	}

	return &session, nil
}
