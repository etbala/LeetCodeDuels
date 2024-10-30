package store

import (
	"leetcodeduels/internal/enums"
	"time"
)

type Problem struct {
	ID         int              `json:"id"`
	Name       string           `json:"name"`
	Slug       string           `json:"slug"`
	Difficulty enums.Difficulty `json:"difficulty"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID          int64
	Username    string
	AccessToken string
	CreatedAt   time.Time
	UpdatedAt   time.Time // (Last Logged In)
	Rating      int
}

func NewProblem(ID int, name string, slug string, difficulty enums.Difficulty) *Problem {
	return &Problem{
		ID:         ID,
		Name:       name,
		Slug:       slug,
		Difficulty: difficulty,
	}
}

func NewTag(ID int, Name string) *Tag {
	return &Tag{
		ID:   ID,
		Name: Name,
	}
}

func NewUser(ID int64, username string, rating int) *User {
	return &User{
		ID:       ID,
		Username: username,
		Rating:   rating,
	}
}
