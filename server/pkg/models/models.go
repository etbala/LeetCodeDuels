package models

import "leetcodeduels/internal/enums"

/*
	Database & HTTPS Models
*/

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
	UUID         string
	Username     string
	PasswordHash string
	Email        string
	Rating       int
	Friends      []string
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

func NewUser(UUID string, username string, rating int, friends []string) *User {
	return &User{
		UUID:     UUID,
		Username: username,
		Rating:   rating,
		Friends:  friends,
	}
}
