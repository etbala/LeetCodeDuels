package ws

import (
	"encoding/json"
	"time"
)

// Message defines the structure of messages exchanged via WebSockets
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type InvitationPayload struct {
	InvitationID string `json:"invitationID"`
	FromUserID   int64  `json:"fromUserID"`
	FromUsername string `json:"fromUsername"`
}

type InvitationResponsePayload struct {
	InvitationID string `json:"invitationID"`
	Accepted     bool   `json:"accepted"`
}

type StartGamePayload struct {
	SessionID  int    `json:"sessionID"`
	ProblemURL string `json:"problemURL"`
}

// SubmissionPayload is sent when a player makes a submission
type SubmissionPayload struct {
	ID              int    `json:"SubmissionID"`
	PlayerID        int64  `json:"PlayerID"`
	PassedTestCases int    `json:"PassedTestCases"`
	TotalTestCases  int    `json:"TotalTestCases"`
	Status          string `json:"Status"`
	Runtime         int    `json:"Runtime"`
	Memory          int    `json:"Memory"`
	Lang            string `json:"Lang"`
}

// OpponentSubmissionPayload notifies a player about the opponent's submission
type OpponentSubmissionPayload struct {
	ID              int       `json:"SubmissionID"`
	PlayerID        int64     `json:"playerID"`
	Status          string    `json:"status"`
	PassedTestCases int       `json:"PassedTestCases"`
	TotalTestCases  int       `json:"TotalTestCases"`
	Runtime         int       `json:"Runtime"`
	Memory          int       `json:"Memory"`
	Time            time.Time `json:"Time"`
}

// GameOverPayload announces the end of the game
type GameOverPayload struct {
	WinnerID  int64 `json:"winnerID"`
	SessionID int   `json:"sessionID"`
	Duration  int64 `json:"duration"` // in seconds
}

// ErrorPayload conveys error messages
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func MarshalPayload(v interface{}) json.RawMessage {
	bytes, _ := json.Marshal(v)
	return bytes
}
