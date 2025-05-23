package ws

import (
	"encoding/json"
	"time"
)

// Messages Client Sends
const (
	ClientMsgSendInvitation    = "send_invitation"
	ClientMsgAcceptInvitation  = "accept_invitation"
	ClientMsgDeclineInvitation = "decline_invitation"
	ClientMsgCancelInvitation  = "cancel_invitation" // No Payload
	ClientMsgEnterQueue        = "enter_queue"
	ClientMsgLeaveQueue        = "leave_queue" // No Payload
	ClientMsgSubmission        = "submission"
)

// Messages Server Sends
const (
	ServerMsgError              = "error"
	ServerMsgInvitationCanceled = "invitation_canceled"
	ServerMsgInviteDoesNotExist = "invitation_nonexistant" // No Payload
	ServerMsgStartGame          = "start_game"
	ServerMsgGameOver           = "game_over"
	ServerMsgOpponentSubmission = "opponent_submission"
	ServerMsgOtherLogon         = "other_logon" // When another device logs into same account
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
	IsRated      bool     `json:"IsRated"` // Possibly replace with "Mode" param in the future
	Difficulties []string `json:"difficulties"`
	Tags         []string `json:"tags"`
}

type AcceptInvitationPayload struct {
	InviteeID int64 `json:"inviteeID"`
}

type DeclineInvitationPayload struct {
	InviteeID int64 `json:"inviteeID"`
}

type EnterQueue struct {
	Difficulties []string `json:"difficulties"`
	Tags         []string `json:"tags"`
}

type SubmissionPayload struct {
	Status          string    `json:"status"`
	PassedTestCases int       `json:"PassedTestCases"`
	TotalTestCases  int       `json:"TotalTestCases"`
	Runtime         int       `json:"Runtime"`
	Memory          int       `json:"Memory"`
	Time            time.Time `json:"Time"`
}

type InvitationCanceledPayload struct {
	InviteeID int64 `json:"inviteeID"`
}

type StartGamePayload struct {
	SessionID  int64  `json:"sessionID"`
	ProblemURL string `json:"problemURL"`
	OpponentID int64  `json:"opponentID"`
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
