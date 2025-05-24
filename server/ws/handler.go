package ws

import (
	"encoding/json"
	"fmt"
	"leetcodeduels/models"
	"leetcodeduels/services"
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
		ctx.handleCancelInvitation(env.Payload)
	case ClientMsgEnterQueue:
		ctx.handleEnterQueue(env.Payload)
	case ClientMsgLeaveQueue:
		ctx.handleLeaveQueue(env.Payload)
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

	matchDetails := models.MatchDetails{IsRated: p.IsRated, Difficulties: p.Difficulties, Tags: p.Tags}
	success, err := services.InviteManager.CreateInvite(ctx.UserID, p.InviteeID, matchDetails)
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

	request := InvitationRequestPayload{InviterID: ctx.UserID, IsRated: p.IsRated, Difficulties: p.Difficulties, Tags: p.Tags}
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
	err := json.Unmarshal(raw, &p)
	if err != nil {

	}

}

func (ctx *wsContext) handleDeclineInvitation(raw json.RawMessage) {

}

func (ctx *wsContext) handleCancelInvitation(raw json.RawMessage) {

}

func (ctx *wsContext) handleEnterQueue(raw json.RawMessage) {

}

func (ctx *wsContext) handleLeaveQueue(raw json.RawMessage) {

}

func (ctx *wsContext) handleSubmission(raw json.RawMessage) {

}

func (ctx *wsContext) replyError(code, msg string) {
	errMsg := Message{
		Type:    ServerMsgError,
		Payload: MarshalPayload(ErrorPayload{Code: code, Message: msg}),
	}
	b, _ := json.Marshal(errMsg)
	ctx.SendCh <- b
}
