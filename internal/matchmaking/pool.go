package matchmaking

import (
    "fmt"
    "sync"
    "leetcodeduels/model"
)

type MatchmakingPool struct {
    sync.Mutex
    Players []*model.Player
}

func NewMatchmakingPool() *MatchmakingPool {
    return &MatchmakingPool{}
}

func (mp *MatchmakingPool) AddPlayer(player *model.Player) {
    mp.Lock()
    defer mp.Unlock()

    mp.Players = append(mp.Players, player)
    // Trigger matchmaking whenever a new player is added
    mp.triggerMatchmaking()
}

func (mp *MatchmakingPool) triggerMatchmaking() {
    // Simplified example: trying to match the first two players if they have the same Tag
    if len(mp.Players) >= 2 && mp.Players[0].Tag == mp.Players[1].Tag {
        player1 := mp.Players[0]
        player2 := mp.Players[1]

        // Notify the players that they have been matched
        lobby := &model.Lobby{Player1: player1, Player2: player2}
        player1.Matched <- lobby
        player2.Matched <- lobby

        fmt.Printf("Matched %s and %s\n", player1.ID, player2.ID)

        // Remove the matched players from the pool
        mp.removePlayers(player1.ID, player2.ID)
    }
}

func (mp *MatchmakingPool) removePlayers(ids ...string) {
    var newPlayers []*model.Player
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
