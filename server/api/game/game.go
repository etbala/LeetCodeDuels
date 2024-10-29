package game

import (
	"leetcodeduels/internal/interfaces"
	"leetcodeduels/pkg/models"
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

func (gm *GameManager) CreateSession(player1, player2 interfaces.Player, problem *models.Problem) *Session {
	gm.Lock()
	defer gm.Unlock()

	sessionID := len(gm.Sessions) + 1
	session := &Session{
		ID:         sessionID,
		InProgress: true,
		Problem:    *problem,
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
		// Handle error: session not found
		return
	}

	// Determine winner and update session state
	session.InProgress = false
	session.EndTime = time.Now()

	// TODO: Store session in history somewhere (probably in db)

	player1 := session.Players[0]
	player2 := session.Players[1]

	delete(gm.Players, player1.UUID)
	delete(gm.Players, player2.UUID)

	delete(gm.Sessions, sessionID)

	// TODO: Notify users that their session is complete

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

func (gm *GameManager) AddSubmission(playerID string, submission PlayerSubmission) {
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

	if submission.Status != Accepted {
		// TODO: Notify other player of submission

		return
	}

	// If the submission is successful, handle winner and close session
	session.Winner = session.Players[playerIndex]
	gm.EndSession(sessionID)
}

func (gm *GameManager) IsPlayerInSession(playerID string) bool {
	gm.Lock()
	defer gm.Unlock()

	_, ok := gm.Players[playerID]
	return ok
}

func (gm *GameManager) GetPlayersSessionID(playerID string) (int, err) {
	gm.Lock()
	defer gm.Unlock()

	sessionID, err := gm.Players[playerID]
	if err != nil {
		return nil, "Player not in session."
	}

	return id, nil
}

func (gm *GameManager) CalculateNewMMR(player1UUID, player2UUID, winnerUUID string) error {
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
	err := store.DataStore.UpdateUserRating(player1.UUID, newRating1)
	if err != nil {
		return err
	}
	err = store.DataStore.UpdateUserRating(player2.UUID, newRating2)
	if err != nil {
		return err
	}

	return nil
}
