package models

import "time"

type Lobby struct {
	Player1 *Player
	Player2 *Player
}

type Problem struct {
	ProblemID  int    `json:"problem_id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Difficulty string `json:"difficulty"`
}

// Tag represents a tag entity as stored in your database.
type Tag struct {
	TagID   int    `json:"tag_id"`
	TagName string `json:"tag_name"`
}

type User struct {
	UserID   int
	Username string
	Rating   int
	Friends  []string
}

type Player struct {
	ID         string
	Matched    chan *Lobby
	Tags       []string  // A slice of tags/flags for matchmaking
	JoinedAt   time.Time // The time when the player joined the matchmaking pool
	ForceMatch bool      // Whether the player has opted for forced matching after a timeout
}

func NewProblem(problemID int, name string, url string, difficulty string) *Problem {
	return &Problem{
		ProblemID:  problemID,
		Name:       name,
		URL:        url,
		Difficulty: difficulty,
	}
}

// NewTag creates a new instance of Tag.
func NewTag(tagID int, tagName string) *Tag {
	return &Tag{
		TagID:   tagID,
		TagName: tagName,
	}
}

func NewUser(userID int, username string, rating int, friends []string) *User {
	return &User{
		UserID:   userID,
		Username: username,
		Rating:   rating,
		Friends:  friends,
	}
}

func NewPlayer(id string, tags []string) *Player {
	return &Player{
		ID:      id,
		Tags:    tags,
		Matched: make(chan *Lobby, 1),
	}
}
