package matchmaking

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"
    "leetcodeduels/model"
)

type MatchmakingPool struct {
    sync.Mutex
    Players    []*model.Player
    MatchCheck time.Duration
}

func NewMatchmakingPool() *MatchmakingPool {
    mp := &MatchmakingPool{
        MatchCheck: 1 * time.Second, // Check for matches every 1 second
    }
    go mp.periodicMatchmaking() // Start the periodic matchmaking routine
    return mp
}

func (mp *MatchmakingPool) AddPlayer(player *model.Player) {
    mp.Lock()
    mp.Players = append(mp.Players, player)
    mp.Unlock()
}

func (mp *MatchmakingPool) periodicMatchmaking() {
    for {
        time.Sleep(mp.MatchCheck)
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
                    break
                }
            }
            if !matchFound {
                i++
            }
        }
        mp.Unlock()
    }
}

func (mp *MatchmakingPool) shouldMatch(player1, player2 *model.Player) bool {
    for _, tag1 := range player1.Tags {
        for _, tag2 := range player2.Tags {
            if tag1 == tag2 {
                return true
            }
        }
    }
    if time.Since(player1.JoinedAt) >= mp.MatchCheck && player1.ForceMatch ||
       time.Since(player2.JoinedAt) >= mp.MatchCheck && player2.ForceMatch {
        return true
    }
    return false
}

func (mp *MatchmakingPool) notifyMatch(player1, player2 *model.Player) {
    lobby := &model.Lobby{Player1: player1, Player2: player2}
    player1.Matched <- lobby
    player2.Matched <- lobby
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

func allowCORS(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
        if r.Method == "OPTIONS" {
            return
        }
        next(w, r)
    }
}

func matchmakeHandler(pool *MatchmakingPool) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        player := &model.Player{
            ID:       fmt.Sprintf("Player%d", pool.Size()+1),
            Tags:     []string{"Tag1"},
            Matched:  make(chan *model.Lobby, 1),
            JoinedAt: time.Now(),
        }

        pool.AddPlayer(player)

        select {
        case lobby := <-player.Matched:
            json.NewEncoder(w).Encode(lobby)
        case <-time.After(30 * time.Second):
            w.WriteHeader(http.StatusRequestTimeout)
            fmt.Fprintln(w, "No match found")
        }
    }
}

func main() {
    pool := NewMatchmakingPool()
    
    // Add a dummy player to the matchmaking pool
    dummyPlayer := &model.Player{
        ID:       "DummyPlayer",
        Tags:     []string{"Tag1"},
        Matched:  make(chan *model.Lobby, 1),
        JoinedAt: time.Now(),
    }
    pool.AddPlayer(dummyPlayer)

    http.Handle("/matchmake", allowCORS(matchmakeHandler(pool)))
    log.Fatal(http.ListenAndServe(":8080", nil))
}
