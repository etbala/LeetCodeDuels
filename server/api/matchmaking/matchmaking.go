package matchmaking

import (
	"fmt"
	"time"
)

func MatchmakingRequest(player *Player, pool *MatchmakingPool, timeout time.Duration) {
	select {
	case lobby := <-player.Matched:
		fmt.Printf("Match found for %s: %s and %s\n", player.ID, lobby.Player1.ID, lobby.Player2.ID)
	case <-time.After(timeout):
		fmt.Printf("No match found for %s within the timeout period.\n", player.ID)
	}
}
