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
	ClientMsgForfeit           = "forfeit"   // No Payload
	ClientMsgHeartbeat         = "heartbeat" // No Payload
)

// Messages Server Sends
const (
	ServerMsgError              = "server_error"
	ServerMsgInvitationRequest  = "invitation_request"
	ServerMsgInvitationCanceled = "invitation_canceled"
	ServerMsgInvitationDeclined = "invitation_declined"
	ServerMsgUserOffline        = "user_offline"
	ServerMsgInviteDoesNotExist = "invitation_nonexistent" // No Payload
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
	ID                int64                   `json:"submissionID"`
	ProblemID         int                     `json:"problemID"`
	Status            models.SubmissionStatus `json:"status"`
	PassedTestCases   int                     `json:"passedTestCases"`
	TotalTestCases    int                     `json:"totalTestCases"`
	Runtime           int                     `json:"runtime"`
	RuntimePercentile float32                 `json:"runtimePercentile"`
	Memory            int                     `json:"memory"`
	MemoryPercentile  float32                 `json:"memoryPercentile"`
	Language          models.LanguageType     `json:"language"`
	Time              time.Time               `json:"time"`
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

// Notifies a player about submission their opponent made
type OpponentSubmissionPayload struct {
	ID       int64                   `json:"submissionID"`
	PlayerID int64                   `json:"playerID"`
	Status   models.SubmissionStatus `json:"status"`
	Language models.LanguageType     `json:"language"`
	Time     time.Time               `json:"time"`
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
