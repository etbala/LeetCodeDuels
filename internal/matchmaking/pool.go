package matchmaking

import (
    "fmt"
    "matchmaking-system/internal/model"
    "sync"
    "time"
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
    go mp.tryMatch(player)
}

func (mp *MatchmakingPool) tryMatch(player *model.Player) {
    mp.Lock()
    defer mp.Unlock()

    for i, p := range mp.Players {
        if p.ID != player.ID && p.Tag == player.Tag {
            // Create a lobby and notify both players
            lobby := &model.Lobby{Player1: player, Player2: p}
            player.Matched <- lobby
            p.Matched <- lobby

            // Remove matched players from the pool
            mp.removePlayers(player.ID, p.ID)
            fmt.Printf("Matched %s and %s\n", player.ID, p.ID)
            return
        }
    }
}

func (mp *MatchmakingPool) removePlayers(ids ...string) {
    var newPlayers []*model.Player
    for _, p := range mp.Players {
        keep := true
        for _, id := range ids {
            if p.ID == id {
                keep = false
                break
            }
        }
        if keep {
            newPlayers = append(newPlayers, p)
        }
    }
    mp.Players = newPlayers
}
