package ws

import (
	"encoding/json"
	"time"
)

const (
	MessageTypeStartGame          = "start_game"
	MessageTypeSubmission         = "submission"
	MessageTypeOpponentSubmission = "opponent_submission"
	MessageTypeGameOver           = "game_over"
	MessageTypeError              = "error"
	MessageTypeHeartbeat          = "heartbeat"
	MessageTypeSendInvitation     = "send_invitation"
	MessageTypeInvitationResponse = "invitation_response"
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// --------------------------------
// Client Payloads (Sent by Client)
// --------------------------------

type InvitationResponsePayload struct {
	FromUserID int64 `json:"fromUserID"`
	Accepted   bool  `json:"accepted"`
}

// Note: Temporary, Until LeetCode API is used to verify submissions.
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

// --------------------------------
// Server Payloads (Sent by Server)
// --------------------------------

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type InvitationPayload struct {
	FromUserID   int64  `json:"fromUserID"`
	FromUsername string `json:"fromUsername"`
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

func MarshalPayload(v interface{}) json.RawMessage {
	bytes, _ := json.Marshal(v)
	return bytes
}
