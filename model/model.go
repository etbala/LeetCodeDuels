package model

type Player struct {
    ID      string
    Tag     string
    Matched chan *Lobby
}

func NewPlayer(id, tag string) *Player {
    return &Player{
        ID:      id,
        Tag:     tag,
        Matched: make(chan *Lobby, 1),
    }
}

type Lobby struct {
    Player1 *Player
    Player2 *Player
}
