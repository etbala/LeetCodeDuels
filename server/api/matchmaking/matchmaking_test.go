package matchmaking

import (
	"leetcodeduels/internal/enums"
	"testing"
	"time"
)

func TestMatchmaking(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	pool.AddPlayer(1, []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, []int{1}, false)
	pool.AddPlayer(2, []enums.Difficulty{enums.EASY, enums.MEDIUM, enums.HARD}, []int{1}, false)

	validateMatch(t, pool, 1, 2)
}

func TestMatchmakingWithOverlappingTags(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	pool.AddPlayer(1, []enums.Difficulty{enums.EASY}, []int{1, 2}, false)
	pool.AddPlayer(2, []enums.Difficulty{enums.EASY}, []int{2, 3}, false)

	validateMatch(t, pool, 1, 2)
}

func TestMatchmakingWithNoCommonTagsAndTimeout(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	pool.AddPlayer(1, []enums.Difficulty{enums.EASY}, []int{4}, true)
	time.Sleep(1 * time.Second) // Simulate delay between player joins
	pool.AddPlayer(2, []enums.Difficulty{enums.EASY}, []int{5}, true)

	// Wait for the force match to trigger
	time.Sleep(5 * time.Second)

	validateMatch(t, pool, 1, 2)
}

func TestMatchmakingPrioritizesOldestPlayer(t *testing.T) {
	resetMatchmakingPool()
	pool := GetMatchmakingPool()

	pool.AddPlayer(0, []enums.Difficulty{enums.EASY}, []int{1}, false)
	time.Sleep(1 * time.Second) // Simulate delay
	pool.AddPlayer(1, []enums.Difficulty{enums.EASY}, []int{1}, false)

	// Wait for matchmaking
	time.Sleep(2 * time.Second)

	validateMatch(t, pool, 0, 1)
}

func validateMatch(t *testing.T, pool *MatchmakingPool, player1ID, expectedMatchID int64) {
	select {
	case match := <-pool.MatchChan:
		if (match.Player1ID != player1ID || match.Player2ID != expectedMatchID) &&
			(match.Player1ID != expectedMatchID || match.Player2ID != player1ID) {
			t.Errorf("Expected players %d and %d to be matched, but got %d and %d", player1ID, expectedMatchID, match.Player1ID, match.Player2ID)
		}
	case <-time.After(10 * time.Second):
		t.Errorf("Timeout occurred while waiting for a match between players %d and %d", player1ID, expectedMatchID)
	}
}
