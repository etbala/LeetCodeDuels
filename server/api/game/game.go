package game

import (
	"leetcodeduels/api/game/models"
	"leetcodeduels/api/matchmaking"
	"sync"
	"time"
)

// Handles Game Sessions
// Creates new session when matchmaker finds a game -> adds the players to this session
// Functions to update state of game session if user finishes problem
// Some way to check if a user is currently in a game session, useful to prevent users from being in multiple games at once

type GameManager struct {
	sync.Mutex
	Sessions map[int]*models.Session // Map session ID to Session
	Players  map[int]string          // Map player ID to session ID
}

func NewGameManager() *GameManager {
	return &GameManager{
		Sessions: make(map[int]*models.Session),
		Players:  make(map[int]string),
	}
}

func (gm *GameManager) CreateSession(lobby *matchmaking.Lobby, question *models.Question) *models.Session {
	gm.Lock()
	defer gm.Unlock()

	sessionID := len(gm.Sessions) + 1 // Simple ID generation, consider a more robust method
	session := &models.Session{
		ID:          sessionID,
		InProgress:  true,
		Question:    *question,
		Players:     []models.Player{*lobby.Player1, *lobby.Player2},
		Submissions: make([][]models.PlayerSubmission, 2), // Assuming 2 players for now
		StartTime:   time.Now(),
	}

	gm.Sessions[sessionID] = session
	gm.Players[lobby.Player1.ID] = sessionID
	gm.Players[lobby.Player2.ID] = sessionID

	return session
}

func (gm *GameManager) UpdateSessionForPlayer(playerID int, submission models.PlayerSubmission) {
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

	// Assuming player1 is always at index 0 and player2 at index 1 for simplicity
	playerIndex := 0
	if session.Players[1].ID == playerID {
		playerIndex = 1
	}

	session.Submissions[playerIndex] = append(session.Submissions[playerIndex], submission)
	// Additional logic to check if the session ends, determine winner, etc., could be here
}

func (gm *GameManager) IsPlayerInSession(playerID int) bool {
	gm.Lock()
	defer gm.Unlock()

	_, ok := gm.Players[playerID]
	return ok
}
