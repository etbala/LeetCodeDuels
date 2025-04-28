package models

type Difficulty string

const (
	EASY   Difficulty = "Easy"
	MEDIUM Difficulty = "Medium"
	HARD   Difficulty = "Hard"
)

type Problem struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Slug       string     `json:"slug"`
	Difficulty Difficulty `json:"difficulty"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
