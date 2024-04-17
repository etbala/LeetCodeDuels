package game

import "time"

const BASE_PROBLEM_URL string = "https://leetcode.com/problemset/"

type PlayerInfo interface {
	GetID() string
	GetUsername() string
}

type Player struct {
	UUID     string
	Username string
	RoomID   int
}

type Question struct {
	ID         int
	Title      string
	TitleSlug  string
	Difficulty string
	Tags       []string
}

type PlayerSubmission struct {
	Question Question
	Pass     bool
	Time     time.Time
}

type Session struct {
	ID          int
	InProgress  bool
	Question    Question
	Players     []Player
	Submissions [][]PlayerSubmission
	Winner      Player
	StartTime   time.Time
	EndTime     time.Time
}
