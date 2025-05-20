package ws

import (
	"encoding/json"
	"time"
)

const (
	MessageTypeError              = "error"
	MessageTypeHeartbeat          = "heartbeat"
	MessageTypeSendInvitation     = "send_invitation"
	MessageTypeAcceptInvitation   = "accept_invitation"
	MessageTypeDeclineInvitation  = "decline_invitation"
	MessageTypeCancelInvitation   = "cancel_invitation"
	MessageTypeInvitationCanceled = "invitation_canceled"
	MessageTypeStartGame          = "start_game"
	MessageTypeGameOver           = "game_over"
	MessageTypeSubmission         = "submission"
	MessageTypeOpponentSubmission = "opponent_submission"
	MessageTypeOtherLogon         = "other_logon" // When another device logs into same account
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SendInvitationPayload struct {
	FromUserID   int64    `json:"fromUserID"`
	Difficulties []string `json:"difficulties"`
	Tags         []string `json:"tags"`
	// TODO: Add game information here
}

type AcceptInvitationPayload struct {
	InviteeID int64 `json:"inviteeID"`
}

type DeclineInvitationPayload struct {
	InviteeID int64 `json:"inviteeID"`
}

type InvitationCanceledPayload struct {
	InviteeID int64 `json:"inviteeID"`
}

type StartGamePayload struct {
	SessionID  int64  `json:"sessionID"`
	ProblemURL string `json:"problemURL"`
	OpponentID int64  `json:"opponentID"`
}

type SubmissionPayload struct {
	Status          string    `json:"status"`
	PassedTestCases int       `json:"PassedTestCases"`
	TotalTestCases  int       `json:"TotalTestCases"`
	Runtime         int       `json:"Runtime"`
	Memory          int       `json:"Memory"`
	Time            time.Time `json:"Time"`
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
