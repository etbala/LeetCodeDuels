package models

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Difficulty string

const (
	Easy   Difficulty = "Easy"
	Medium Difficulty = "Medium"
	Hard   Difficulty = "Hard"
)

func ParseDifficulty(difficulty string) (Difficulty, error) {
	switch difficulty {
	case "Easy":
		return Easy, nil
	case "Medium":
		return Medium, nil
	case "Hard":
		return Hard, nil
	default:
		return "", errors.New("invalid Difficulty value")
	}
}

func (s *Difficulty) UnmarshalJSON(data []byte) error {
	// Trim quotes from JSON string
	var difficultyStr string
	if err := json.Unmarshal(data, &difficultyStr); err != nil {
		return err
	}

	switch difficultyStr {
	case "Easy", "Medium", "Hard":
		*s = Difficulty(difficultyStr)
		return nil
	default:
		return fmt.Errorf("invalid match status: %s", difficultyStr)
	}
}

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
