package matchmaking

// func MatchmakingRequest(player *Player, pool *MatchmakingPool, timeout time.Duration) {
// 	select {
// 	case lobby := <-player.Matched:
// 		fmt.Printf("Match found for %d: %d and %d\n", player.ID, lobby.Player1.ID, lobby.Player2.ID)
// 	case <-time.After(timeout):
// 		fmt.Printf("No match found for %d within the timeout period.\n", player.ID)
// 	}
// }
