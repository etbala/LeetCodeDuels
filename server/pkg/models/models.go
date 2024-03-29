package models

// Database & HTTPS Models

type Problem struct {
	ProblemID  int    `json:"problem_id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
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

func NewProblem(problemID int, name string, url string, difficulty string) *Problem {
	return &Problem{
		ProblemID:  problemID,
		Name:       name,
		URL:        url,
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
