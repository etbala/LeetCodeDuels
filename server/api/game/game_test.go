package game

import (
	"leetcodeduels/pkg/models"
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

func (mpi MockPlayerInfo) GetRating() int {
	return 1200
}

// TestCreateSession tests creating a new session correctly adds sessions and updates player mapping
func TestCreateSession(t *testing.T) {
	resetGameManager()
	gm := GetGameManager()
	player1 := MockPlayerInfo{"1", "Alice"}
	player2 := MockPlayerInfo{"2", "Bob"}
	problem := &models.Problem{
		ID:         101,
		Name:       "Example Problem",
		Slug:       "example-problem",
		Difficulty: "Easy",
	}

	session := gm.CreateSession(player1, player2, problem)

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

// TestAddSubmission tests updating session submissions for a player
func TestAddSubmission(t *testing.T) {
	resetGameManager()
	gm := GetGameManager()
	player1 := MockPlayerInfo{"1", "Alice"}
	player2 := MockPlayerInfo{"2", "Bob"}
	problem := &models.Problem{
		ID:         101,
		Name:       "Example Problem",
		Slug:       "example-problem",
		Difficulty: "Easy",
	}
	session := gm.CreateSession(player1, player2, problem)

	submission := PlayerSubmission{
		PlayerUUID:      player1.GetID(),
		PassedTestCases: 10,
		TotalTestCases:  10,
		Status:          Accepted,
		Runtime:         150,
		Time:            time.Now(),
	}

	gm.AddSubmission("1", submission)
	if len(session.Submissions[0]) != 1 {
		t.Errorf("Expected 1 submission for player 1, got %d", len(session.Submissions[0]))
	}
}

// TestIsPlayerInSession tests the check for a player's session existence
func TestIsPlayerInSession(t *testing.T) {
	resetGameManager()
	gm := GetGameManager()
	player1 := MockPlayerInfo{"1", "Alice"}
	player2 := MockPlayerInfo{"2", "Bob"}
	problem := &models.Problem{
		ID:         101,
		Name:       "Example Problem",
		Slug:       "example-problem",
		Difficulty: "Easy",
	}
	gm.CreateSession(player1, player2, problem)

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
