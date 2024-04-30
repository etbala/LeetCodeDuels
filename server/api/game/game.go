package game

import (
	"leetcodeduels/pkg/store"
	"math"
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

var (
	instance *GameManager
	once     sync.Once
)

func GetGameManager() *GameManager {
	once.Do(func() {
		instance = &GameManager{
			Sessions: make(map[int]*Session),
			Players:  make(map[string]int),
		}
	})
	return instance
}

func resetGameManager() {
	instance = nil
	once = sync.Once{}
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

func (gm *GameManager) EndSession(sessionID int) {
	gm.Lock()
	defer gm.Unlock()

	session, exists := gm.Sessions[sessionID]
	if !exists {
		return // Handle error: session not found
	}

	// Determine winner and update session state
	session.InProgress = false
	session.EndTime = time.Now()
	// Logic to determine winner could be here

	// Remove session from active sessions and clear player mappings
	delete(gm.Sessions, sessionID)
	for _, player := range session.Players {
		delete(gm.Players, player.UUID)
	}
}

func (gm *GameManager) ForceEndSession(sessionID int) {
	gm.Lock()
	defer gm.Unlock()

	session, exists := gm.Sessions[sessionID]
	if !exists {
		return // Handle error: session not found
	}

	session.InProgress = false
	session.EndTime = time.Now()
	// Possible additional clean-up logic

	delete(gm.Sessions, sessionID)
	for _, player := range session.Players {
		delete(gm.Players, player.UUID)
	}
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

func (gm *GameManager) CalculateNewMMR(player1UUID, player2UUID, winnerUUID string, store *store.Store) error {
	player1, player2 := gm.Sessions[gm.Players[player1UUID]].Players[0], gm.Sessions[gm.Players[player2UUID]].Players[1]

	kFactor := 32
	rating1 := float64(player1.Rating)
	rating2 := float64(player2.Rating)

	// Calculate expected scores
	expScore1 := 1 / (1 + math.Pow(10, (rating2-rating1)/400))
	expScore2 := 1 - expScore1

	// Actual scores
	var actualScore1, actualScore2 float64
	if winnerUUID == player1UUID {
		actualScore1 = 1 // player1 wins
		actualScore2 = 0 // player2 loses
	} else if winnerUUID == player2UUID {
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
	err := store.UpdateUserRating(player1.UUID, newRating1)
	if err != nil {
		return err
	}
	err = store.UpdateUserRating(player2.UUID, newRating2)
	if err != nil {
		return err
	}

	return nil
}
