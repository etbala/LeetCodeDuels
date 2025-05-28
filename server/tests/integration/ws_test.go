package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"leetcodeduels/auth"
	"leetcodeduels/models"
	"leetcodeduels/ws"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func wsURL() string {
	return "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
}

func dialWS(t *testing.T, userID int64) *websocket.Conn {
	token, err := auth.GenerateJWT(userID)
	require.NoError(t, err)

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL(), header)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	return conn
}

func readMessage(t *testing.T, c *websocket.Conn) ws.Message {
	c.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var m ws.Message
	err := c.ReadJSON(&m)
	require.NoError(t, err)
	return m
}

func TestInvitationAcceptFlow(t *testing.T) {
	inviter := dialWS(t, 12345)
	defer inviter.Close()

	invitee := dialWS(t, 67890)
	defer invitee.Close()

	time.Sleep(100 * time.Millisecond)

	invite := ws.SendInvitationPayload{
		InviteeID:    67890,
		MatchDetails: models.MatchDetails{Tags: []int{1}, Difficulties: []models.Difficulty{models.Easy}},
	}
	err := inviter.WriteJSON(ws.Message{
		Type:    ws.ClientMsgSendInvitation,
		Payload: ws.MarshalPayload(invite),
	})
	require.NoError(t, err)

	reqMsg := readMessage(t, invitee)
	require.Equal(t, ws.ServerMsgInvitationRequest, reqMsg.Type)
	var reqPayload ws.InvitationRequestPayload
	require.NoError(t, json.Unmarshal(reqMsg.Payload, &reqPayload))
	require.Equal(t, int64(12345), reqPayload.InviterID)
	require.Equal(t, invite.MatchDetails, reqPayload.MatchDetails)

	accept := ws.AcceptInvitationPayload{InviterID: 12345}
	err = invitee.WriteJSON(ws.Message{
		Type:    ws.ClientMsgAcceptInvitation,
		Payload: ws.MarshalPayload(accept),
	})
	require.NoError(t, err)

	m1 := readMessage(t, inviter)
	m2 := readMessage(t, invitee)

	require.Equal(t, ws.ServerMsgStartGame, m1.Type)
	require.Equal(t, ws.ServerMsgStartGame, m2.Type)

	var p1, p2 ws.StartGamePayload
	require.NoError(t, json.Unmarshal(m1.Payload, &p1))
	require.NoError(t, json.Unmarshal(m2.Payload, &p2))

	require.NotEmpty(t, p1.SessionID)
	require.Equal(t, p1.SessionID, p2.SessionID)
	require.Equal(t, int64(67890), p1.OpponentID)
	require.Equal(t, int64(12345), p2.OpponentID)
}

func TestDeclineInvitation(t *testing.T) {
	inviter := dialWS(t, 12345)
	defer inviter.Close()
	invitee := dialWS(t, 67890)
	defer invitee.Close()

	invite := ws.SendInvitationPayload{
		InviteeID:    67890,
		MatchDetails: models.MatchDetails{Tags: []int{1}, Difficulties: []models.Difficulty{models.Easy}},
	}
	err := inviter.WriteJSON(ws.Message{
		Type:    ws.ClientMsgSendInvitation,
		Payload: ws.MarshalPayload(invite),
	})
	require.NoError(t, err)

	reqMsg := readMessage(t, invitee)
	require.Equal(t, ws.ServerMsgInvitationRequest, reqMsg.Type)
	var reqPayload ws.InvitationRequestPayload
	require.NoError(t, json.Unmarshal(reqMsg.Payload, &reqPayload))
	require.Equal(t, int64(12345), reqPayload.InviterID)
	require.Equal(t, invite.MatchDetails, reqPayload.MatchDetails)

	err = invitee.WriteJSON(ws.Message{
		Type:    ws.ClientMsgDeclineInvitation,
		Payload: ws.MarshalPayload(ws.DeclineInvitationPayload{InviterID: 12345}),
	})
	require.NoError(t, err)

	m := readMessage(t, inviter)
	require.Equal(t, ws.ServerMsgInvitationDeclined, m.Type)
}

func TestCancelInvitation(t *testing.T) {
	inviter := dialWS(t, 12345)
	defer inviter.Close()
	invitee := dialWS(t, 67890)
	defer invitee.Close()

	time.Sleep(100 * time.Millisecond)

	invite := ws.SendInvitationPayload{
		InviteeID:    67890,
		MatchDetails: models.MatchDetails{Tags: []int{1}, Difficulties: []models.Difficulty{models.Easy}},
	}
	err := inviter.WriteJSON(ws.Message{
		Type:    ws.ClientMsgSendInvitation,
		Payload: ws.MarshalPayload(invite),
	})
	require.NoError(t, err)

	reqMsg := readMessage(t, invitee)
	require.Equal(t, ws.ServerMsgInvitationRequest, reqMsg.Type)
	var reqPayload ws.InvitationRequestPayload
	require.NoError(t, json.Unmarshal(reqMsg.Payload, &reqPayload))
	require.Equal(t, int64(12345), reqPayload.InviterID)
	require.Equal(t, invite.MatchDetails, reqPayload.MatchDetails)

	err = inviter.WriteJSON(ws.Message{Type: ws.ClientMsgCancelInvitation})
	require.NoError(t, err)

	m := readMessage(t, invitee)
	require.Equal(t, ws.ServerMsgInvitationCanceled, m.Type)
	var p ws.InvitationCanceledPayload
	require.NoError(t, json.Unmarshal(m.Payload, &p))
	require.Equal(t, int64(12345), p.InviterID)
}

func TestUnknown(t *testing.T) {
	c := dialWS(t, 12345)
	defer c.Close()

	// unknown type
	err := c.WriteJSON(map[string]any{"type": "foo_bar", "payload": map[string]any{}})
	require.NoError(t, err)
	m := readMessage(t, c)
	require.Equal(t, ws.ServerMsgError, m.Type)
	var e ws.ErrorPayload
	require.NoError(t, json.Unmarshal(m.Payload, &e))
	require.Equal(t, "unknown_type", e.Code)
}
