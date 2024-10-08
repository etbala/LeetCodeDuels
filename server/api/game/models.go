package game

import (
	"errors"
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

func ParseSubmissionStatus(status string) (SubmissionStatus, error) {
	switch status {
	case "Accepted":
		return Accepted, nil
	case "Compile Error":
		return CompileError, nil
	case "Memory Limit Exceeded":
		return MemoryLimitExceeded, nil
	case "Runtime Error":
		return RuntimeError, nil
	case "Time Limit Exceeded":
		return TimeLimitExceeded, nil
	case "Wrong Answer":
		return WrongAnswer, nil
	default:
		return "", errors.New("invalid SubmissionStatus value")
	}
}

type Lang string // LeetCode supported languages (possible submission languages)

const (
	c          Lang = "c"
	cpp        Lang = "cpp"
	csharp     Lang = "csharp"
	java       Lang = "java"
	python     Lang = "python"
	python3    Lang = "python3"
	javascript Lang = "javascript"
	typescript Lang = "typescript"
	php        Lang = "php"
	swift      Lang = "swift"
	kotlin     Lang = "kotlin"
	dart       Lang = "dart"
	golang     Lang = "go"
	ruby       Lang = "ruby"
	scala      Lang = "scala"
	rust       Lang = "rust"
	racket     Lang = "racket"
	erlang     Lang = "erlang"
	elixir     Lang = "elixir"
)

func ParseLang(lang string) (Lang, error) {
	switch lang {
	case "c":
		return c, nil
	case "cpp":
		return cpp, nil
	case "csharp":
		return csharp, nil
	case "java":
		return java, nil
	case "python":
		return python, nil
	case "python3":
		return python3, nil
	case "javascript":
		return javascript, nil
	case "typescript":
		return typescript, nil
	case "php":
		return php, nil
	case "swift":
		return swift, nil
	case "kotlin":
		return kotlin, nil
	case "dart":
		return dart, nil
	case "go":
		return golang, nil
	case "ruby":
		return ruby, nil
	case "scala":
		return scala, nil
	case "rust":
		return rust, nil
	case "racket":
		return racket, nil
	case "erlang":
		return erlang, nil
	case "elixir":
		return elixir, nil
	default:
		return "", errors.New("invalid Lang value")
	}
}

type Player struct {
	UUID     string
	Username string
	Rating   int
	RoomID   int
}

type PlayerSubmission struct {
	ID              int
	PlayerUUID      string
	PassedTestCases int
	TotalTestCases  int
	Status          SubmissionStatus
	Runtime         int // ms
	Memory          int // Bytes (Convert to MB sometime later)
	Time            time.Time
	Lang            Lang
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
