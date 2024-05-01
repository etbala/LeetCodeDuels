package matchmaking

import "time"

type Lobby struct {
	Player1 *Player
	Player2 *Player
}

type Player struct {
	ID         string
	Username   string
	Rating     int
	Matched    chan *Lobby
	Tags       []string  // A slice of tags/flags for matchmaking
	JoinedAt   time.Time // The time when the player joined the matchmaking pool
	ForceMatch bool      // Whether the player has opted for forced matching after a timeout
}

func NewPlayer(id string, username string, tags []string, rating int) *Player {
	return &Player{
		ID:       id,
		Username: username,
		Tags:     tags,
		Matched:  make(chan *Lobby, 1),
	}
}

func (p *Player) GetID() string {
	return p.ID
}

func (p *Player) GetUsername() string {
	return p.Username
}

func (p *Player) GetRating() int {
	return p.Rating
}
