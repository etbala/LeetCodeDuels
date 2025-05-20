package ws

import (
	"encoding/json"
	"time"
)

const (
	MessageTypeStartGame          = "start_game"
	MessageTypeGameOver           = "game_over"
	MessageTypeError              = "error"
	MessageTypeHeartbeat          = "heartbeat"
	MessageTypeSendInvitation     = "send_invitation"
	MessageTypeSubmission         = "submission"
	MessageTypeOpponentSubmission = "opponent_submission"
	MessageTypeTerminate          = "terminate"      // When another user logs on
	MessageTypeInternalError      = "internal_error" // When server crashed and is shutting down
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type InvitationPayload struct {
	FromUserID   int64    `json:"fromUserID"`
	FromUsername string   `json:"fromUsername"`
	Difficulties []string `json:"difficulties"`
	Tags         []string `json:"tags"`
	// TODO: Add game information here
}

type StartGamePayload struct {
	SessionID  int    `json:"sessionID"`
	ProblemURL string `json:"problemURL"`
}

// Notifies a player about the opponent's submission
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

type GameOverPayload struct {
	WinnerID  int64 `json:"winnerID"`
	SessionID int   `json:"sessionID"`
	Duration  int64 `json:"duration"` // in seconds
}

func MarshalPayload(v any) json.RawMessage {
	bytes, _ := json.Marshal(v)
	return bytes
}
