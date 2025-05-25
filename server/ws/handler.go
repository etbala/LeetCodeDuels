package ws

import (
	"encoding/json"
	"fmt"
	"leetcodeduels/models"
	"leetcodeduels/services"
	"leetcodeduels/store"
)

type wsContext struct {
	UserID int64
	ConnID string
	SendCh chan<- []byte
}

func (ctx *wsContext) Handle(env messageEnvelope) {
	switch env.Type {
	case ClientMsgSendInvitation:
		ctx.handleSendInvitation(env.Payload)
	case ClientMsgAcceptInvitation:
		ctx.handleAcceptInvitation(env.Payload)
	case ClientMsgDeclineInvitation:
		ctx.handleDeclineInvitation(env.Payload)
	case ClientMsgCancelInvitation:
		ctx.handleCancelInvitation()
	case ClientMsgEnterQueue:
		ctx.handleEnterQueue(env.Payload)
	case ClientMsgLeaveQueue:
		ctx.handleLeaveQueue()
	case ClientMsgSubmission:
		ctx.handleSubmission(env.Payload)
	default:
		// unknown msg type
		ctx.replyError("unknown_type", fmt.Sprintf("unhandled message type: %s", env.Type))
	}
}

func (ctx *wsContext) handleSendInvitation(raw json.RawMessage) {
	var p SendInvitationPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		ctx.replyError("invalid_payload", "couldn't parse invitation")
		return
	}

	success, err := services.InviteManager.CreateInvite(ctx.UserID, p.InviteeID, p.MatchDetails)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}
	if success == false {
		// Invite already exists from this user
		// TODO: Either replace the existing invite or ignore this invite
		return
	}

	// Lookup invitee's active connection
	inviteeConn, err := ConnManager.GetConnection(p.InviteeID)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}
	if inviteeConn == "" {
		// TODO: tell client user is offline
		return
	}

	request := InvitationRequestPayload{InviterID: ctx.UserID, MatchDetails: p.MatchDetails}
	payload, _ := json.Marshal(request)

	// Forward the InvitationPayload to invitee via Redis pub/sub
	msg := Message{Type: ServerMsgInvitationRequest, Payload: payload}
	b, _ := json.Marshal(msg)

	err = ConnManager.SendToConn(inviteeConn, b)
	if err != nil {

	}
}

func (ctx *wsContext) handleAcceptInvitation(raw json.RawMessage) {
	var p AcceptInvitationPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		ctx.replyError("invalid_payload", "couldn't parse accept_invitation")
		return
	}

	invite, err := services.InviteManager.InviteDetails(p.InviterID)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}
	if invite == nil {
		b2, _ := json.Marshal(Message{Type: ServerMsgInviteDoesNotExist})
		ctx.SendCh <- b2
		return
	}

	// remove the invite
	removed, err := services.InviteManager.RemoveInvite(p.InviterID)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}
	if !removed {
		ctx.replyError("invite_not_found", "no pending invitation to accept")
		return
	}

	problem, err := store.DataStore.GetRandomProblemDuel(invite.MatchDetails.Tags, invite.MatchDetails.Difficulties)

	// start the session
	sessionID, err := services.GameManager.StartGame(
		[]int64{p.InviterID, ctx.UserID},
		*problem,
	)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}

	// re-fetch to read the chosen problem URL
	session, err := services.GameManager.GetGame(sessionID)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}

	problemURL := fmt.Sprintf("https://leetcode.com/problems/%s", session.Problem.Slug)

	// notify inviter
	inviterConn, err := ConnManager.GetConnection(p.InviterID)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}
	if inviterConn != "" {
		startInv := StartGamePayload{
			SessionID:  sessionID,
			ProblemURL: problemURL,
			OpponentID: ctx.UserID,
		}
		b, _ := json.Marshal(Message{Type: ServerMsgStartGame, Payload: MarshalPayload(startInv)})
		_ = ConnManager.SendToConn(inviterConn, b)
	}

	// notify accepter
	startYou := StartGamePayload{
		SessionID:  sessionID,
		ProblemURL: problemURL,
		OpponentID: p.InviterID,
	}
	b2, _ := json.Marshal(Message{Type: ServerMsgStartGame, Payload: MarshalPayload(startYou)})
	ctx.SendCh <- b2
}

func (ctx *wsContext) handleDeclineInvitation(raw json.RawMessage) {
	var p AcceptInvitationPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		ctx.replyError("invalid_payload", "couldn't parse decline_invitation")
		return
	}

	invite, err := services.InviteManager.InviteDetails(p.InviterID)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}
	if invite == nil {
		// No invite to decline
		return
	}

	inviterConn, err := ConnManager.GetConnection(p.InviterID)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}

	b, _ := json.Marshal(Message{Type: ServerMsgInvitationDeclined})
	_ = ConnManager.SendToConn(inviterConn, b)
}

func (ctx *wsContext) handleCancelInvitation() {
	invite, err := services.InviteManager.InviteDetails(ctx.UserID)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}
	if invite == nil {
		// No invite to cancel
		return
	}

	success, err := services.InviteManager.RemoveInvite(ctx.UserID)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}
	if success == false {
		// this shouldn’t really happen, but guard anyway
	}

	inviterConn, err := ConnManager.GetConnection(ctx.UserID)
	if err != nil {
		ctx.replyError("server_error", err.Error())
		return
	}

	payload := InvitationCanceledPayload{InviterID: ctx.UserID}
	b, _ := json.Marshal(Message{Type: ServerMsgInvitationCanceled, Payload: MarshalPayload(payload)})
	_ = ConnManager.SendToConn(inviterConn, b)
}

func (ctx *wsContext) handleEnterQueue(raw json.RawMessage) {
	ctx.replyError("unimplemented", "")
}

func (ctx *wsContext) handleLeaveQueue() {
	ctx.replyError("unimplemented", "")
}

func (ctx *wsContext) handleSubmission(raw json.RawMessage) {
	var p SubmissionPayload
	err := json.Unmarshal(raw, &p)
	if err != nil {

	}

	// Get the session id associated with userID
	sessionID, err := services.GameManager.GetSessionIDByPlayer(ctx.UserID)
	if err != nil {

	}
	if sessionID == "" {
		// User is not in-game
		return
	}

	session, err := services.GameManager.GetGame(sessionID)
	if err != nil {

	}
	if session == nil {
		// this shouldn’t really happen, but guard anyway
		return
	}

	submissionID := len(session.Submissions)
	submissionStatus, _ := models.ParseSubmissionStatus(p.Status)
	submissionLang, _ := models.ParseLang(p.Language)
	submission := models.PlayerSubmission{
		ID:              submissionID,
		PlayerID:        ctx.UserID,
		PassedTestCases: p.PassedTestCases,
		TotalTestCases:  p.TotalTestCases,
		Status:          submissionStatus,
		Runtime:         p.Runtime,
		Memory:          p.Memory,
		Lang:            submissionLang,
		Time:            p.Time,
	}

	// TODO: Verify submission information is correct against LeetCode's API

	err = services.GameManager.AddSubmission(sessionID, submission)
	if err != nil {

	}

	opponentID, err := services.GameManager.GetOpponent(sessionID, ctx.UserID)
	if err != nil {

	}

	opponentConn, err := ConnManager.GetConnection(opponentID)
	if err != nil {

	}

	if submissionStatus == models.Accepted {
		err = services.GameManager.CompleteGame(sessionID, ctx.UserID)
		if err != nil {

		}

		// TODO: Notify both players of completed game
		// TODO: Store this match in long-term storage
		return
	}

	// TODO: Notify opponent of failed player submission
	reply := OpponentSubmissionPayload{
		ID:              submissionID,
		PlayerID:        ctx.UserID,
		PassedTestCases: p.PassedTestCases,
		TotalTestCases:  p.TotalTestCases,
		Status:          p.Status,
		Runtime:         p.Runtime,
		Memory:          p.Memory,
		Language:        p.Language,
		Time:            p.Time,
	}

	payload, _ := json.Marshal(reply)

	// Forward the InvitationPayload to invitee via Redis pub/sub
	msg := Message{Type: ServerMsgInvitationRequest, Payload: payload}
	b, _ := json.Marshal(msg)

	err = ConnManager.SendToConn(opponentConn, b)
	if err != nil {

	}

}

func (ctx *wsContext) replyError(code, msg string) {
	errMsg := Message{
		Type:    ServerMsgError,
		Payload: MarshalPayload(ErrorPayload{Code: code, Message: msg}),
	}
	b, _ := json.Marshal(errMsg)
	ctx.SendCh <- b
}
