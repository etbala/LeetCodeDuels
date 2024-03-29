package models

import "time"

type Lobby struct {
	Player1 *Player
	Player2 *Player
}

type Player struct {
	ID         string
	Matched    chan *Lobby
	Tags       []string  // A slice of tags/flags for matchmaking
	JoinedAt   time.Time // The time when the player joined the matchmaking pool
	ForceMatch bool      // Whether the player has opted for forced matching after a timeout
}

func NewPlayer(id string, tags []string) *Player {
	return &Player{
		ID:      id,
		Tags:    tags,
		Matched: make(chan *Lobby, 1),
	}
}
