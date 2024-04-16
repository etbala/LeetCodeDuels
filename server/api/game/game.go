package game

import (
	"sync"
	"time"
)

// Handles Game Sessions
// Creates new session when matchmaker finds a game -> adds the players to this session
// Functions to update state of game session if user finishes problem
// Some way to check if a user is currently in a game session, useful to prevent users from being in multiple games at once

type GameManager struct {
	sync.Mutex
	Sessions map[int]*Session // Map session ID to Session
	Players  map[string]int   // Map player ID to session ID
}

func NewGameManager() *GameManager {
	return &GameManager{
		Sessions: make(map[int]*Session),
		Players:  make(map[string]int),
	}
}

func (gm *GameManager) CreateSession(player1, player2 PlayerInfo, question *Question) *Session {
	gm.Lock()
	defer gm.Unlock()

	sessionID := len(gm.Sessions) + 1
	session := &Session{
		ID:         sessionID,
		InProgress: true,
		Question:   *question,
		Players: []Player{
			{UUID: player1.GetID(), Username: player1.GetUsername(), RoomID: sessionID},
			{UUID: player2.GetID(), Username: player2.GetUsername(), RoomID: sessionID},
		},
		Submissions: make([][]PlayerSubmission, 2),
		StartTime:   time.Now(),
	}

	gm.Sessions[sessionID] = session
	gm.Players[player1.GetID()] = sessionID
	gm.Players[player2.GetID()] = sessionID

	return session
}

func (gm *GameManager) UpdateSessionForPlayer(playerID string, submission PlayerSubmission) {
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

	// Assuming player1 is always at index 0 and player2 at index 1
	playerIndex := 0
	if session.Players[1].UUID == playerID {
		playerIndex = 1
	}

	session.Submissions[playerIndex] = append(session.Submissions[playerIndex], submission)
	// Additional logic to check if the session ends, determine winner, etc., could be here
}

func (gm *GameManager) IsPlayerInSession(playerID string) bool {
	gm.Lock()
	defer gm.Unlock()

	_, ok := gm.Players[playerID]
	return ok
}
