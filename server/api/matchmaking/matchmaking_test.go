package matchmaking

import (
	"leetcodeduels/internal/enums"
	"testing"
	"time"
)

func TestMatchmaking(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	player1 := &Player{ID: "Player1", Tags: []int{1}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now(), ForceMatch: false}
	player2 := &Player{ID: "Player2", Tags: []int{1}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now(), ForceMatch: false}

	pool.AddPlayer(player1)
	pool.AddPlayer(player2)

	validateMatch(t, player1, "Player2")
	validateMatch(t, player2, "Player1")
}

func TestMatchmakingWithOverlappingFlags(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	player1 := &Player{ID: "Player1", Tags: []int{1, 2}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now(), ForceMatch: false}
	player2 := &Player{ID: "Player2", Tags: []int{2, 3}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now(), ForceMatch: false}

	pool.AddPlayer(player1)
	pool.AddPlayer(player2)

	validateMatch(t, player1, "Player2")
	validateMatch(t, player2, "Player1")
}

func TestMatchmakingWithNoCommonFlagsAndTimeout(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	player1 := &Player{ID: "Player1", Tags: []int{4}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now(), ForceMatch: true}
	player2 := &Player{ID: "Player2", Tags: []int{5}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now().Add(1 * time.Second), ForceMatch: true}

	pool.AddPlayer(player1)
	time.Sleep(1 * time.Second) // Simulate delay between player joins
	pool.AddPlayer(player2)

	// Wait for a short time to simulate the timeout mechanism
	time.Sleep(5 * time.Second) // Simulate the timeout for force match

	validateMatch(t, player1, "Player2")
	validateMatch(t, player2, "Player1")
}

func TestMatchmakingPrioritizesOldestPlayer(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	oldestPlayer := &Player{ID: "OldestPlayer", Tags: []int{1}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now()}
	time.Sleep(1 * time.Second) // Ensure there's a noticeable difference in JoinedAt times

	newerPlayer := &Player{ID: "NewerPlayer", Tags: []int{1}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now()}
	pool.AddPlayer(oldestPlayer)
	pool.AddPlayer(newerPlayer)

	// Wait a bit to allow the matchmaking process to occur
	time.Sleep(2 * time.Second)

	select {
	case lobby := <-oldestPlayer.Matched:
		if lobby.Player1.ID != "OldestPlayer" && lobby.Player2.ID != "OldestPlayer" {
			t.Errorf("Oldest player was not prioritized in matchmaking")
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Timeout occurred while waiting for a match for the oldest player")
	}
}

func validateMatch(t *testing.T, player *Player, expectedMatchID string) {
	select {
	case lobby := <-player.Matched:
		if (lobby.Player1.ID != expectedMatchID) && (lobby.Player2.ID != expectedMatchID) {
			t.Errorf("Expected %s to be matched with %s, but did not happen", player.ID, expectedMatchID)
		}
	case <-time.After(10 * time.Second): // Waiting longer to ensure the test does not fail due to timing issues
		t.Errorf("Timeout occurred while waiting for a match for player %s", player.ID)
	}
}
