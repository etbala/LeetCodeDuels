package matchmaking

import (
	"leetcodeduels/api/matchmaking/models"
	"sync"
	"time"
)

type MatchmakingPool struct {
	sync.Mutex
	Players    []*models.Player
	MatchCheck time.Duration
}

func NewMatchmakingPool() *MatchmakingPool {
	mp := &MatchmakingPool{
		MatchCheck: 1 * time.Second, // Check for matches every 5 seconds
	}
	go mp.periodicMatchmaking() // Start the periodic matchmaking routine
	return mp
}

func (mp *MatchmakingPool) AddPlayer(player *models.Player) {
	mp.Lock()
	mp.Players = append(mp.Players, player)
	mp.Unlock()
}

func (mp *MatchmakingPool) periodicMatchmaking() {
	for {
		time.Sleep(mp.MatchCheck) // Wait for the specified duration before checking for matches
		mp.Lock()
		i := 0
		for i < len(mp.Players) {
			player1 := mp.Players[i]
			matchFound := false
			for j := i + 1; j < len(mp.Players); j++ {
				player2 := mp.Players[j]
				if mp.shouldMatch(player1, player2) {
					mp.notifyMatch(player1, player2)
					mp.removePlayers(player1.ID, player2.ID)
					matchFound = true
					break // Exit the inner loop as we've found a match
				}
			}
			if !matchFound {
				i++ // Only increment if no match was found to avoid skipping players
			}
			// If a match was found, don't increment i as the next player will have shifted to the current index
		}
		mp.Unlock()
	}
}

func (mp *MatchmakingPool) shouldMatch(player1, player2 *models.Player) bool {
	// Check for overlapping flags or if force match is enabled after a timeout
	for _, tag1 := range player1.Tags {
		for _, tag2 := range player2.Tags {
			if tag1 == tag2 {
				return true
			}
		}
	}
	// Check if force matching is triggered (assuming a 'ForceMatch' field in Player model)
	if time.Since(player1.JoinedAt) >= mp.MatchCheck && player1.ForceMatch ||
		time.Since(player2.JoinedAt) >= mp.MatchCheck && player2.ForceMatch {
		return true
	}
	return false
}

func (mp *MatchmakingPool) notifyMatch(player1, player2 *models.Player) {
	lobby := &models.Lobby{Player1: player1, Player2: player2}
	player1.Matched <- lobby
	player2.Matched <- lobby
}

func (mp *MatchmakingPool) removePlayers(ids ...string) {
	var newPlayers []*models.Player
	for _, player := range mp.Players {
		keep := true
		for _, id := range ids {
			if player.ID == id {
				keep = false
				break
			}
		}
		if keep {
			newPlayers = append(newPlayers, player)
		}
	}
	mp.Players = newPlayers
}

func (mp *MatchmakingPool) Size() int {
	mp.Lock()
	defer mp.Unlock()
	return len(mp.Players)
}
