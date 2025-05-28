package ws

import (
	"encoding/json"
	"leetcodeduels/models"
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
	ServerMsgError              = "server_error"
	ServerMsgInvitationRequest  = "invitation_request"
	ServerMsgInvitationCanceled = "invitation_canceled"
	ServerMsgInvitationDeclined = "invitation_declined"
	ServerMsgUserOffline        = "user_offline"
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
	InviteeID    int64               `json:"inviteeID"`
	MatchDetails models.MatchDetails `json:"matchDetails"`
}

type AcceptInvitationPayload struct {
	InviterID int64 `json:"inviterID"`
}

type DeclineInvitationPayload struct {
	InviterID int64 `json:"inviterID"`
}

type EnterQueuePayload struct {
	Difficulties []string `json:"difficulties"`
	Tags         []int    `json:"tags"`
}

type SubmissionPayload struct {
	Status          string    `json:"status"`
	PassedTestCases int       `json:"passedTestCases"`
	TotalTestCases  int       `json:"totalTestCases"`
	Runtime         int       `json:"runtime"`
	Memory          int       `json:"memory"`
	Language        string    `json:"language"`
	Time            time.Time `json:"time"`
}

type InvitationRequestPayload struct {
	InviterID    int64               `json:"inviterID"`
	MatchDetails models.MatchDetails `json:"matchDetails"`
}

type InvitationCanceledPayload struct {
	InviterID int64 `json:"inviterID"`
}

type StartGamePayload struct {
	SessionID  string `json:"sessionID"`
	ProblemURL string `json:"problemURL"`
	OpponentID int64  `json:"opponentID"`
}

// Notifies a player about the opponent's submission
type OpponentSubmissionPayload struct {
	ID              int       `json:"submissionID"`
	PlayerID        int64     `json:"playerID"`
	Status          string    `json:"status"`
	PassedTestCases int       `json:"passedTestCases"`
	TotalTestCases  int       `json:"totalTestCases"`
	Runtime         int       `json:"runtime"`
	Memory          int       `json:"memory"`
	Language        string    `json:"language"`
	Time            time.Time `json:"time"`
}

type GameOverPayload struct {
	WinnerID  int64  `json:"winnerID"`
	SessionID string `json:"sessionID"`
	Duration  int64  `json:"duration"` // in seconds
}

func MarshalPayload(v any) json.RawMessage {
	bytes, _ := json.Marshal(v)
	return bytes
}
