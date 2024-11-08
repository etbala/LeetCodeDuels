package matchmaking

import (
	"errors"
	"leetcodeduels/api/game"
	"leetcodeduels/internal/enums"
	"leetcodeduels/pkg/store"
	"sync"
	"time"
)

type MatchmakingPool struct {
	sync.Mutex
	Players    map[int64]*PlayerInfo
	MatchCheck time.Duration
	MatchChan  chan Match
}

var (
	poolInstance *MatchmakingPool
	poolOnce     sync.Once
)

func GetMatchmakingPool() *MatchmakingPool {
	poolOnce.Do(func() {
		poolInstance = &MatchmakingPool{
			Players:    make(map[int64]*PlayerInfo),
			MatchCheck: 2 * time.Second,       // Check for matches every 2 seconds
			MatchChan:  make(chan Match, 100), // Buffered channel
		}
		go poolInstance.periodicMatchmaking() // Start the periodic matchmaking routine
	})
	return poolInstance
}

func resetMatchmakingPool() {
	poolInstance = nil
	poolOnce = sync.Once{}
}

func (mp *MatchmakingPool) AddPlayer(id int64, difficulties []enums.Difficulty, tags []int, forceMatch bool) error {
	mp.Lock()
	defer mp.Unlock()

	// Check if player is already in matchmaking pool
	if _, exists := mp.Players[id]; exists {
		return errors.New("Player is already in the matchmaking pool")
	}

	playerInfo := &PlayerInfo{
		ID:           id,
		Difficulties: difficulties,
		Tags:         tags,
		JoinedAt:     time.Now(),
		ForceMatch:   forceMatch,
	}

	mp.Players[id] = playerInfo

	return nil
}

func (mp *MatchmakingPool) periodicMatchmaking() {
	ticker := time.NewTicker(mp.MatchCheck)
	defer ticker.Stop()

	for {
		<-ticker.C // Wait for the specified duration before checking for matches
		mp.Lock()
		playerIDs := make([]int64, 0, len(mp.Players))
		for id := range mp.Players {
			playerIDs = append(playerIDs, id)
		}

		for i := 0; i < len(playerIDs); i++ {
			player1 := mp.Players[playerIDs[i]]
			matchFound := false
			for j := i + 1; j < len(playerIDs); j++ {
				player2 := mp.Players[playerIDs[j]]
				if mp.shouldMatch(player1, player2) {
					success := mp.createMatch(player1.ID, player2.ID)
					if success {
						mp.MatchChan <- Match{Player1ID: player1.ID, Player2ID: player2.ID}
						mp.RemovePlayers(player1.ID, player2.ID)
						matchFound = true
						break
					}
				}
			}
			if matchFound {
				// Remove player1 from the list to prevent further matching in this cycle
				playerIDs = append(playerIDs[:i], playerIDs[i+1:]...)
				i-- // Adjust index after removal
			}
		}
		mp.Unlock()
	}
}

func (mp *MatchmakingPool) shouldMatch(player1, player2 *PlayerInfo) bool {
	// Check for overlapping Difficulties
	for _, dif1 := range player1.Difficulties {
		for _, dif2 := range player2.Difficulties {
			if dif1 == dif2 {
				return true
			}
		}
	}

	// Check if force matching is triggered
	if time.Since(player1.JoinedAt) >= mp.MatchCheck && player1.ForceMatch ||
		time.Since(player2.JoinedAt) >= mp.MatchCheck && player2.ForceMatch {
		return true
	}
	return false
}

func (mp *MatchmakingPool) createMatch(player1ID, player2ID int64) bool {
	player1 := mp.Players[player1ID]
	player2 := mp.Players[player2ID]

	var difficulties []enums.Difficulty
	for _, dif1 := range player1.Difficulties {
		for _, dif2 := range player2.Difficulties {
			if dif1 == dif2 {
				difficulties = append(difficulties, dif1)
			}
		}
	}

	// Assuming GetRandomProblemForDuel uses the IDs to fetch problems
	prob, err := store.DataStore.GetRandomProblemForDuel(player1.Tags, player2.Tags, difficulties)
	if prob == nil || err != nil {
		return false
	}

	gameManager := game.GetGameManager()
	gameManager.CreateSession(player1ID, player2ID, prob)
	return true
}

func (mp *MatchmakingPool) RemovePlayers(ids ...int64) {
	for _, id := range ids {
		delete(mp.Players, id)
	}
}

func (mp *MatchmakingPool) Size() int {
	mp.Lock()
	defer mp.Unlock()
	return len(mp.Players)
}
