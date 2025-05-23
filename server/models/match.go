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
	c          LanguageType = "c"
	cpp        LanguageType = "cpp"
	csharp     LanguageType = "csharp"
	java       LanguageType = "java"
	python     LanguageType = "python"
	python3    LanguageType = "python3"
	javascript LanguageType = "javascript"
	typescript LanguageType = "typescript"
	php        LanguageType = "php"
	swift      LanguageType = "swift"
	kotlin     LanguageType = "kotlin"
	dart       LanguageType = "dart"
	golang     LanguageType = "go"
	ruby       LanguageType = "ruby"
	scala      LanguageType = "scala"
	rust       LanguageType = "rust"
	racket     LanguageType = "racket"
	erlang     LanguageType = "erlang"
	elixir     LanguageType = "elixir"
)

func ParseLang(lang string) (LanguageType, error) {
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
		return "", errors.New("invalid LanguageType value")
	}
}

func (l *LanguageType) UnmarshalJSON(data []byte) error {
	var langStr string
	if err := json.Unmarshal(data, &langStr); err != nil {
		return err
	}

	switch langStr {
	case "Go", "cpp": // Add other supported languages
		*l = LanguageType(langStr)
		return nil
	default:
		return fmt.Errorf("invalid language: %s", langStr)
	}
}

type PlayerSubmission struct {
	ID              int              `json:"SubmissionID"`
	PlayerID        int64            `json:"PlayerID"`
	PassedTestCases int              `json:"PassedTestCases"`
	TotalTestCases  int              `json:"TotalTestCases"`
	Status          SubmissionStatus `json:"Status"`
	Runtime         int              `json:"Runtime"`
	Memory          int              `json:"Memory"`
	Lang            LanguageType     `json:"Lang"`
	Time            time.Time        `json:"Time"`
}

type Session struct {
	ID          int
	InProgress  bool
	IsRated     bool
	Problem     Problem
	Players     []int64
	Submissions [][]PlayerSubmission // Map of player to that players submissions
	Winner      int64
	StartTime   time.Time
	EndTime     time.Time
}
