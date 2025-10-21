package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type SubmissionStatus string

const (
	Accepted            SubmissionStatus = "Accepted"
	CompileError        SubmissionStatus = "Compile Error"
	MemoryLimitExceeded SubmissionStatus = "Memory Limit Exceeded"
	RuntimeError        SubmissionStatus = "Runtime Error"
	OutputLimitExceeded SubmissionStatus = "Output Limit Exceeded"
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
	case "Output Limit Exceeded":
		return OutputLimitExceeded, nil
	case "Time Limit Exceeded":
		return TimeLimitExceeded, nil
	case "Wrong Answer":
		return WrongAnswer, nil
	default:
		return "", errors.New("invalid SubmissionStatus value")
	}
}

func (s *SubmissionStatus) UnmarshalJSON(data []byte) error {
	// Trim quotes from JSON string
	var statusStr string
	if err := json.Unmarshal(data, &statusStr); err != nil {
		return err
	}

	switch statusStr {
	case "Accepted", "Compile Error", "Memory Limit Exceeded", "Runtime Error", "Time Limit Exceeded", "Wrong Answer":
		*s = SubmissionStatus(statusStr)
		return nil
	default:
		return fmt.Errorf("invalid submission status: %s", statusStr)
	}
}

type LanguageType string // LeetCode supported languages (possible submission languages)

const (
	C          LanguageType = "c"
	Cpp        LanguageType = "cpp"
	Csharp     LanguageType = "csharp"
	Java       LanguageType = "java"
	Python     LanguageType = "python"
	Python3    LanguageType = "python3"
	Javascript LanguageType = "javascript"
	Typescript LanguageType = "typescript"
	Php        LanguageType = "php"
	Swift      LanguageType = "swift"
	Kotlin     LanguageType = "kotlin"
	Dart       LanguageType = "dart"
	Golang     LanguageType = "go"
	Ruby       LanguageType = "ruby"
	Scala      LanguageType = "scala"
	Rust       LanguageType = "rust"
	Racket     LanguageType = "racket"
	Erlang     LanguageType = "erlang"
	Elixir     LanguageType = "elixir"
)

func ParseLang(lang string) (LanguageType, error) {
	switch lang {
	case "c":
		return C, nil
	case "cpp":
		return Cpp, nil
	case "csharp":
		return Csharp, nil
	case "java":
		return Java, nil
	case "python":
		return Python, nil
	case "python3":
		return Python3, nil
	case "javascript":
		return Javascript, nil
	case "typescript":
		return Typescript, nil
	case "php":
		return Php, nil
	case "swift":
		return Swift, nil
	case "kotlin":
		return Kotlin, nil
	case "dart":
		return Dart, nil
	case "go":
		return Golang, nil
	case "ruby":
		return Ruby, nil
	case "scala":
		return Scala, nil
	case "rust":
		return Rust, nil
	case "racket":
		return Racket, nil
	case "erlang":
		return Erlang, nil
	case "elixir":
		return Elixir, nil
	default:
		return "", errors.New("invalid LanguageType value")
	}
}

func (l *LanguageType) UnmarshalJSON(data []byte) error {
	var langStr string
	if err := json.Unmarshal(data, &langStr); err != nil {
		return err
	}

	switch langStr {
	case "c", "cpp", "csharp", "java", "python", "python3",
		"javascript", "typescript", "php", "swift", "kotlin",
		"dart", "go", "ruby", "scala", "rust", "racket",
		"erlang", "elixir":
		*l = LanguageType(langStr)
		return nil
	default:
		return fmt.Errorf("invalid language: %s", langStr)
	}
}

type MatchStatus string

const (
	MatchActive   MatchStatus = "Active"
	MatchWon      MatchStatus = "Won"
	MatchCanceled MatchStatus = "Canceled"
	MatchReverted MatchStatus = "Reverted"
)

func ParseMatchStatus(status string) (MatchStatus, error) {
	switch status {
	case "Accepted":
		return MatchActive, nil
	case "Won":
		return MatchWon, nil
	case "Canceled":
		return MatchCanceled, nil
	case "Reverted":
		return MatchReverted, nil
	default:
		return "", errors.New("invalid MatchStatus value")
	}
}

func (s *MatchStatus) UnmarshalJSON(data []byte) error {
	// Trim quotes from JSON string
	var statusStr string
	if err := json.Unmarshal(data, &statusStr); err != nil {
		return err
	}

	switch statusStr {
	case "Active", "Won", "Canceled", "Reverted":
		*s = MatchStatus(statusStr)
		return nil
	default:
		return fmt.Errorf("invalid match status: %s", statusStr)
	}
}

// Submission attached to Game Session
type PlayerSubmission struct {
	ID              int64            `json:"submissionID"`
	PlayerID        int64            `json:"playerID"`
	PassedTestCases int              `json:"passedTestCases"`
	TotalTestCases  int              `json:"totalTestCases"`
	Status          SubmissionStatus `json:"status"`
	Runtime         int              `json:"runtime"`
	Memory          int              `json:"memory"`
	Lang            LanguageType     `json:"lang"`
	Time            time.Time        `json:"time"`
}

type Session struct {
	ID          string             `json:"sessionID"`
	Status      MatchStatus        `json:"status"`
	IsRated     bool               `json:"rated"`
	Problem     Problem            `json:"problem"`
	Players     []int64            `json:"players"`
	Submissions []PlayerSubmission `json:"submissions"`
	Winner      int64              `json:"winner"` // <= 0 if no winner
	StartTime   time.Time          `json:"startTime"`
	EndTime     time.Time          `json:"endTime"`
}

type MatchDetails struct {
	IsRated      bool         `json:"isRated"`
	Difficulties []Difficulty `json:"difficulties"`
	Tags         []int        `json:"tags"`
}

type Invite struct {
	InviterID    int64        `json:"inviterID"`
	InviteeID    int64        `json:"inviteeID"`
	MatchDetails MatchDetails `json:"matchDetails"`
	CreatedAt    time.Time    `json:"createdAt"`
}
