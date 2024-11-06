package matchmaking

import (
	"leetcodeduels/internal/enums"
	"testing"
	"time"
)

func TestMatchmaking(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	player1 := &Player{ID: 1, Tags: []int{1}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now(), ForceMatch: false}
	player2 := &Player{ID: 2, Tags: []int{1}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now(), ForceMatch: false}

	pool.AddPlayer(1, []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, []int{1}, false)
	pool.AddPlayer(2, []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, []int{1}, false)

	validateMatch(t, player1, 2)
	validateMatch(t, player2, 1)
}

func TestMatchmakingWithOverlappingFlags(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	player1 := &Player{ID: 1, Tags: []int{1, 2}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now(), ForceMatch: false}
	player2 := &Player{ID: 1, Tags: []int{2, 3}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now(), ForceMatch: false}

	pool.AddPlayer(1, []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, []int{1, 2}, false)
	pool.AddPlayer(2, []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, []int{2, 3}, false)

	validateMatch(t, player1, 2)
	validateMatch(t, player2, 1)
}

func TestMatchmakingWithNoCommonFlagsAndTimeout(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	player1 := &Player{ID: 1, Tags: []int{4}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now(), ForceMatch: true}
	player2 := &Player{ID: 2, Tags: []int{5}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now().Add(1 * time.Second), ForceMatch: true}

	pool.AddPlayer(1, []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, []int{4}, true)
	time.Sleep(1 * time.Second) // Simulate delay between player joins
	pool.AddPlayer(2, []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, []int{5}, true)

	// Wait for a short time to simulate the timeout mechanism
	time.Sleep(5 * time.Second) // Simulate the timeout for force match

	validateMatch(t, player1, 2)
	validateMatch(t, player2, 1)
}

func TestMatchmakingPrioritizesOldestPlayer(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	oldestPlayer := &Player{ID: 0, Tags: []int{1}, Difficulties: []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, Matched: make(chan *Lobby, 1), JoinedAt: time.Now()}

	pool.AddPlayer(0, []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, []int{1}, false)
	time.Sleep(1 * time.Second) // Simulate delay between player joins
	pool.AddPlayer(1, []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, []int{1}, false)

	// Wait a bit to allow the matchmaking process to occur
	time.Sleep(2 * time.Second)

	select {
	case lobby := <-oldestPlayer.Matched:
		if lobby.Player1.ID != 0 && lobby.Player2.ID != 0 {
			t.Errorf("Oldest player was not prioritized in matchmaking")
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Timeout occurred while waiting for a match for the oldest player")
	}
}

func validateMatch(t *testing.T, player *Player, expectedMatchID int64) {
	select {
	case lobby := <-player.Matched:
		if (lobby.Player1.ID != expectedMatchID) && (lobby.Player2.ID != expectedMatchID) {
			t.Errorf("Expected %d to be matched with %d, but did not happen", player.ID, expectedMatchID)
		}
	case <-time.After(10 * time.Second): // Waiting longer to ensure the test does not fail due to timing issues
		t.Errorf("Timeout occurred while waiting for a match for player %d", player.ID)
	}
}
