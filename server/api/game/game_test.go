package game

import (
	"testing"
	"time"
)

// MockPlayerInfo implements PlayerInfo interface for testing
type MockPlayerInfo struct {
	id       string
	username string
}

func (mpi MockPlayerInfo) GetID() string {
	return mpi.id
}

func (mpi MockPlayerInfo) GetUsername() string {
	return mpi.username
}

// TestNewGameManager tests the instantiation of a new GameManager
func TestNewGameManager(t *testing.T) {
	gm := NewGameManager()
	if gm == nil {
		t.Error("NewGameManager() failed, got nil")
	}
	if len(gm.Sessions) != 0 || len(gm.Players) != 0 {
		t.Error("NewGameManager() should initialize empty maps")
	}
}

// TestCreateSession tests creating a new session correctly adds sessions and updates player mapping
func TestCreateSession(t *testing.T) {
	gm := NewGameManager()
	player1 := MockPlayerInfo{"1", "Alice"}
	player2 := MockPlayerInfo{"2", "Bob"}
	question := &Question{
		ID:         101,
		Title:      "Example Problem",
		TitleSlug:  "example-problem",
		Difficulty: "Easy",
	}

	session := gm.CreateSession(player1, player2, question)

	if session == nil {
		t.Fatal("CreateSession() failed, got nil")
	}
	if len(gm.Sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(gm.Sessions))
	}
	if sessionID, ok := gm.Players["1"]; !ok || sessionID != session.ID {
		t.Errorf("Player1 is not correctly mapped to session")
	}
	if sessionID, ok := gm.Players["2"]; !ok || sessionID != session.ID {
		t.Errorf("Player2 is not correctly mapped to session")
	}
	if session.InProgress != true {
		t.Error("Session should be marked as in progress")
	}
	if session.StartTime.After(time.Now()) {
		t.Error("Session start time is in the future")
	}
}

// TestUpdateSessionForPlayer tests updating session submissions for a player
func TestUpdateSessionForPlayer(t *testing.T) {
	gm := NewGameManager()
	player1 := MockPlayerInfo{"1", "Alice"}
	player2 := MockPlayerInfo{"2", "Bob"}
	question := &Question{
		ID:         101,
		Title:      "Example Problem",
		TitleSlug:  "example-problem",
		Difficulty: "Easy",
	}
	session := gm.CreateSession(player1, player2, question)

	submission := PlayerSubmission{
		Question: *question,
		Pass:     true,
		Time:     time.Now(),
	}

	gm.UpdateSessionForPlayer("1", submission)
	if len(session.Submissions[0]) != 1 {
		t.Errorf("Expected 1 submission for player 1, got %d", len(session.Submissions[0]))
	}
}

// TestIsPlayerInSession tests the check for a player's session existence
func TestIsPlayerInSession(t *testing.T) {
	gm := NewGameManager()
	player1 := MockPlayerInfo{"1", "Alice"}
	player2 := MockPlayerInfo{"2", "Bob"}
	question := &Question{
		ID:         101,
		Title:      "Example Problem",
		TitleSlug:  "example-problem",
		Difficulty: "Easy",
	}
	gm.CreateSession(player1, player2, question)

	if !gm.IsPlayerInSession("1") {
		t.Error("Player1 should be in a session")
	}
	if !gm.IsPlayerInSession("2") {
		t.Error("Player2 should be in a session")
	}
	if gm.IsPlayerInSession("3") {
		t.Error("Player3 should not be in a session")
	}
}