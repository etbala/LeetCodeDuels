package game

import (
	"leetcodeduels/pkg/models"
	"time"
)

const BASE_PROBLEM_URL string = "https://leetcode.com/problemset/"

type PlayerInfo interface {
	GetID() string
	GetUsername() string
	GetRating() int
}

type Player struct {
	UUID     string
	Username string
	Rating   int
	RoomID   int
}

type PlayerSubmission struct {
	Problem           models.Problem
	PassedTestCases   int
	TotalTestCases    int
	TimeLimitExceeded bool
	RuntimeError      bool
	WrongAnswer       bool
	CompileError      bool
	Accepted          bool
	Runtime           int
	Time              time.Time
}

type Session struct {
	ID          int
	InProgress  bool
	Problem     models.Problem
	Players     []Player
	Submissions [][]PlayerSubmission
	Winner      Player
	StartTime   time.Time
	EndTime     time.Time
}
