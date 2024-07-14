package game

import (
	"leetcodeduels/pkg/models"
	"time"
)

type SubmissionStatus string

const (
	Accepted            SubmissionStatus = "Accepted"
	CompileError        SubmissionStatus = "Compile Error"
	MemoryLimitExceeded SubmissionStatus = "Memory Limit Exceeded"
	RuntimeError        SubmissionStatus = "Runtime Error"
	TimeLimitExceeded   SubmissionStatus = "Time Limit Exceeded"
	WrongAnswer         SubmissionStatus = "Wrong Answer"
)

type Player struct {
	UUID     string
	Username string
	Rating   int
	RoomID   int
}

type PlayerSubmission struct {
	PlayerUUID      string
	PassedTestCases int
	TotalTestCases  int
	Status          SubmissionStatus
	Runtime         int // ms
	Time            time.Time
}

type Session struct {
	ID          int
	InProgress  bool
	Problem     models.Problem
	Players     []Player             // Should be set
	Submissions [][]PlayerSubmission // Should be Map of player to that players submissions
	Winner      Player
	StartTime   time.Time
	EndTime     time.Time
}
