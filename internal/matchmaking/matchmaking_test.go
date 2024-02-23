package matchmaking

import (
    "leetcodeduels/model"
    "testing"
)

func TestMatchmaking(t *testing.T) {
    pool := NewMatchmakingPool()

    player1 := &model.Player{ID: "Player1", Tag: "Tag1", Matched: make(chan *model.Lobby, 1)} // Buffer channel to avoid blocking
    player2 := &model.Player{ID: "Player2", Tag: "Tag1", Matched: make(chan *model.Lobby, 1)} // Buffer channel to avoid blocking

    pool.AddPlayer(player1)
    pool.AddPlayer(player2)

    // Check if both players received their match notifications
    select {
    case lobby := <-player1.Matched:
        if lobby.Player2.ID != "Player2" {
            t.Errorf("Player1 was not matched with Player2")
        }
    default:
        t.Error("Player1 did not receive a match")
    }

    select {
    case lobby := <-player2.Matched:
        if lobby.Player1.ID != "Player1" {
            t.Errorf("Player2 was not matched with Player1")
        }
    default:
        t.Error("Player2 did not receive a match")
    }

    // Check if the pool is empty after the match
    if size := pool.Size(); size != 0 {
        t.Errorf("Expected pool to be empty after matching, but found %d players", size)
    }
}
