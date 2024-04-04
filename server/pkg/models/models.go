package models

/*
	Database & HTTPS Models
*/

type Problem struct {
	ID         int    `json:"id"`
	FrontendID int    `json:"frontend_id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Difficulty string `json:"difficulty"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID       int
	Username string
    PasswordHash string
    Email String
	Rating   int
	Friends  []string
}

func NewProblem(ID int, frontendID int, name string, slug string, difficulty string) *Problem {
	return &Problem{
		ID:         ID,
		FrontendID: frontendID,
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

func NewUser(ID int, username string, rating int, friends []string) *User {
	return &User{
		ID:       ID,
		Username: username,
		Rating:   rating,
		Friends:  friends,
	}
}
