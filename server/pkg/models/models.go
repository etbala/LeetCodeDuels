package models

/*
	Database & HTTPS Models
*/

type Problem struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Difficulty string `json:"difficulty"`
}

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

func NewProblem(ID int, name string, slug string, difficulty string) *Problem {
	return &Problem{
		ID:         ID,
		Name:       name,
		Slug:       slug,
		Difficulty: difficulty,
	}
}

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
