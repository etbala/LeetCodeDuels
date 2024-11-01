package game

import (
	"errors"
	"leetcodeduels/internal/ws"
	"leetcodeduels/pkg/store"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Handles Game Sessions
// Creates new session when matchmaker finds a game -> adds the players to this session
// Functions to update state of game session if user finishes problem
// Some way to check if a user is currently in a game session, useful to prevent users from being in multiple games at once

type GameManager struct {
	sync.Mutex
	Sessions          map[int]*Session // Map session ID to Session
	Players           map[int64]int    // Map player ID to session ID
	PlayerConnections map[int64]*websocket.Conn
}

var (
	instance *GameManager
	once     sync.Once
)

func GetGameManager() *GameManager {
	once.Do(func() {
		instance = &GameManager{
			Sessions: make(map[int]*Session),
			Players:  make(map[int64]int),
		}
	})
	return instance
}

func resetGameManager() {
	instance = nil
	once = sync.Once{}
}

// AddPlayerConnection associates a WebSocket connection with a player ID
func (gm *GameManager) AddPlayerConnection(playerID int64, conn *websocket.Conn) {
	gm.Lock()
	defer gm.Unlock()
	gm.PlayerConnections[playerID] = conn
}

// RemovePlayerConnection removes the association between a player ID and their connection
func (gm *GameManager) RemovePlayerConnection(playerID int64) {
	gm.Lock()
	defer gm.Unlock()
	delete(gm.PlayerConnections, playerID)
}

// SendMessageToPlayer sends a WebSocket message to a specific player
func (gm *GameManager) SendMessageToPlayer(playerID int64, message ws.Message) error {
	gm.Lock()
	conn, exists := gm.PlayerConnections[playerID]
	gm.Unlock()

	if !exists {
		return errors.New("player not connected")
	}

	return conn.WriteJSON(message)
}

// BroadcastToSession sends a message to all players in a specific session
func (gm *GameManager) BroadcastToSession(sessionID int, message ws.Message) {
	gm.Lock()
	session, exists := gm.Sessions[sessionID]
	if !exists {
		gm.Unlock()
		return
	}

	for _, player := range session.Players {
		conn, ok := gm.PlayerConnections[player]
		if ok {
			// It's safe to release the lock here as we're only reading
			gm.Unlock()
			conn.WriteJSON(message)
			gm.Lock()
		}
	}
	gm.Unlock()
}

func (gm *GameManager) CreateSession(player1ID, player2ID int64, problem *store.Problem) *Session {
	// Get username

	gm.Lock()
	defer gm.Unlock()

	sessionID := len(gm.Sessions) + 1
	session := &Session{
		ID:          sessionID,
		InProgress:  true,
		Problem:     *problem,
		Players:     []int64{player1ID, player2ID},
		Submissions: make([][]PlayerSubmission, 2),
		StartTime:   time.Now(),
	}

	gm.Sessions[sessionID] = session
	gm.Players[player1ID] = sessionID
	gm.Players[player2ID] = sessionID

	// Send start_game message to both players
	startPayload := ws.StartGamePayload{
		SessionID:  sessionID,
		ProblemURL: `https://leetcode.com/problems/` + problem.Slug,
	}

	message := ws.Message{
		Type:    ws.MessageTypeStartGame,
		Payload: ws.MarshalPayload(startPayload),
	}

	gm.BroadcastToSession(sessionID, message)

	return session
}

func (gm *GameManager) EndSession(sessionID int) {
	gm.Lock()
	session, exists := gm.Sessions[sessionID]
	if !exists {
		gm.Unlock()
		return
	}

	session.InProgress = false
	session.EndTime = time.Now()

	player1 := session.Players[0]
	player2 := session.Players[1]

	delete(gm.Players, player1)
	delete(gm.Players, player2)

	delete(gm.Sessions, sessionID)
	gm.Unlock()

	// Determine winner and send game_over message
	winnerID := session.Winner // Assuming Winner is set
	duration := session.EndTime.Sub(session.StartTime).Seconds()

	gameOverPayload := ws.GameOverPayload{
		WinnerID:  winnerID,
		SessionID: sessionID,
		Duration:  int64(duration),
	}

	gameOverMessage := ws.Message{
		Type:    ws.MessageTypeGameOver,
		Payload: ws.MarshalPayload(gameOverPayload),
	}

	gm.BroadcastToSession(sessionID, gameOverMessage)

	// Optionally, remove WebSocket connections if needed
	// For now, we keep connections open
}

func (gm *GameManager) ListSessions() []*Session {
	gm.Lock()
	defer gm.Unlock()

	var sessions []*Session
	for _, session := range gm.Sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (gm *GameManager) AddSubmission(playerID int64, submission PlayerSubmission) {
	gm.Lock()
	defer gm.Unlock()

	sessionID, ok := gm.Players[playerID]
	if !ok {
		// Handle error: Player not in any session
		return
	}

	session, ok := gm.Sessions[sessionID]
	if !ok {
		// Handle error: Session not found
		return
	}

	// Determine player index
	playerIndex := 0
	if session.Players[1] == playerID {
		playerIndex = 1
	}

	session.Submissions[playerIndex] = append(session.Submissions[playerIndex], submission)

	if submission.Status != Accepted {
		// Notify opponent about the submission
		opponentID, err := gm.GetOpponentID(sessionID, playerID)
		if err == nil {
			opponentSubmissionPayload := ws.OpponentSubmissionPayload{
				ID:              submission.ID,
				PlayerID:        playerID,
				Status:          string(submission.Status),
				PassedTestCases: submission.PassedTestCases,
				TotalTestCases:  submission.TotalTestCases,
				Runtime:         submission.Runtime,
				Memory:          submission.Memory,
				Time:            submission.Time,
			}
			opponentMessage := ws.Message{
				Type:    ws.MessageTypeOpponentSubmission,
				Payload: ws.MarshalPayload(opponentSubmissionPayload),
			}
			gm.SendMessageToPlayer(opponentID, opponentMessage)
		}
		return
	}

	// If the submission is accepted, determine the winner and end the session
	session.Winner = session.Players[playerIndex]
	gm.EndSession(sessionID)
}

func (gm *GameManager) IsPlayerInSession(playerID int64) bool {
	gm.Lock()
	defer gm.Unlock()

	_, ok := gm.Players[playerID]
	return ok
}

func (gm *GameManager) GetPlayersSessionID(playerID int64) (int, error) {
	gm.Lock()
	defer gm.Unlock()

	sessionID, err := gm.Players[playerID]
	if err {
		return -1, errors.New("Player not in session.")
	}

	return sessionID, nil
}

func (gm *GameManager) GetOpponentID(sessionID int, playerID int64) (int64, error) {
	gm.Lock()
	defer gm.Unlock()

	session, exists := gm.Sessions[sessionID]
	if !exists {
		return 0, errors.New("session not found")
	}

	for _, player := range session.Players {
		if player != playerID {
			return player, nil
		}
	}

	return 0, errors.New("opponent not found")
}

func (gm *GameManager) CalculateNewMMR(player1ID, player2ID, winnerID int64) error {
	player1, player2 := gm.Sessions[gm.Players[player1ID]].Players[0], gm.Sessions[gm.Players[player2ID]].Players[1]

	// Get Profiles of Players
	profile1, err := store.DataStore.GetUserProfile(player1ID)
	if err != nil {
		return errors.New("Invalid Player1ID when calculating mmr change after match.")
	}
	profile2, err := store.DataStore.GetUserProfile(player2ID)
	if err != nil {
		return errors.New("Invalid Player2ID when calculating mmr change after match.")
	}

	kFactor := 32
	rating1 := float64(profile1.Rating)
	rating2 := float64(profile2.Rating)

	// Calculate expected scores
	expScore1 := 1 / (1 + math.Pow(10, (rating2-rating1)/400))
	expScore2 := 1 - expScore1

	// Actual scores
	var actualScore1, actualScore2 float64
	if winnerID == player1ID {
		actualScore1 = 1 // player1 wins
		actualScore2 = 0 // player2 loses
	} else if winnerID == player2ID {
		actualScore1 = 0 // player1 loses
		actualScore2 = 1 // player2 wins
	} else {
		actualScore1 = 0.5 // draw
		actualScore2 = 0.5 // draw
	}

	// Calculate new ratings
	newRating1 := int(rating1 + float64(kFactor)*(actualScore1-expScore1))
	newRating2 := int(rating2 + float64(kFactor)*(actualScore2-expScore2))

	// Update the ratings in your database/store
	err = store.DataStore.UpdateUserRating(player1, newRating1)
	if err != nil {
		return err
	}
	err = store.DataStore.UpdateUserRating(player2, newRating2)
	if err != nil {
		return err
	}

	return nil
}
