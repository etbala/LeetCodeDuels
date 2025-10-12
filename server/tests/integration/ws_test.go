package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"leetcodeduels/models"
	"leetcodeduels/services"
	"leetcodeduels/ws"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func wsURL() string {
	return "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
}

func TestWSTicketFlow(t *testing.T) {
	t.Run("should successfully connect with a valid ticket", func(t *testing.T) {
		token, err := services.GenerateJWT(12345) // Test user
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.URL+"/api/v1/ws-ticket", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, err)
		var ticketResponse struct {
			Ticket string `json:"ticket"`
		}
		require.NoError(t, json.Unmarshal(body, &ticketResponse))
		require.NotEmpty(t, ticketResponse.Ticket)

		wsURLWithTicket := wsURL() + "?ticket=" + ticketResponse.Ticket
		conn, resp, err := websocket.DefaultDialer.Dial(wsURLWithTicket, nil)
		require.NoError(t, err, "WebSocket dial should succeed with a valid ticket")
		defer conn.Close()
		require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	})

	t.Run("should fail to connect with an invalid or no ticket", func(t *testing.T) {
		// Attempt to dial without a ticket
		_, resp, err := websocket.DefaultDialer.Dial(wsURL(), nil)
		require.Error(t, err, "WebSocket dial should fail without a ticket")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Attempt to dial with a fake ticket
		wsURLWithFakeTicket := wsURL() + "?ticket=fake-ticket-123"
		_, resp, err = websocket.DefaultDialer.Dial(wsURLWithFakeTicket, nil)
		require.Error(t, err, "WebSocket dial should fail with a fake ticket")
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func dialWS(t *testing.T, userID int64) *websocket.Conn {
	token, err := services.GenerateJWT(userID)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.URL+"/api/v1/ws-ticket", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)

	var ticketResponse struct {
		Ticket string `json:"ticket"`
	}
	require.NoError(t, json.Unmarshal(body, &ticketResponse))
	require.NotEmpty(t, ticketResponse.Ticket, "server should return a ticket")

	wsURLWithTicket := wsURL() + "?ticket=" + ticketResponse.Ticket
	conn, resp, err := websocket.DefaultDialer.Dial(wsURLWithTicket, nil)
	require.NoError(t, err, "WebSocket dial should succeed with a valid ticket")
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	time.Sleep(10 * time.Millisecond)
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

	details, err := services.InviteManager.InviteDetails(12345)
	require.NoError(t, err, "should be able to check invite details")
	require.Nil(t, details, "invite must be gone after accept")
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

	details, err := services.InviteManager.InviteDetails(12345)
	require.NoError(t, err, "looking up invite details should not error")
	require.Nil(t, details, "invite should have been deleted after decline")
}

func TestCancelInvitation(t *testing.T) {
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

	err = inviter.WriteJSON(ws.Message{Type: ws.ClientMsgCancelInvitation})
	require.NoError(t, err)

	m := readMessage(t, invitee)
	require.Equal(t, ws.ServerMsgInvitationCanceled, m.Type)
	var p ws.InvitationCanceledPayload
	require.NoError(t, json.Unmarshal(m.Payload, &p))
	require.Equal(t, int64(12345), p.InviterID)

	details, err := services.InviteManager.InviteDetails(12345)
	require.NoError(t, err)
	require.Nil(t, details, "invite must be gone after cancel")
}

func TestSubmissionFlow(t *testing.T) {
	inviterID := int64(12345)
	inviteeID := int64(67890)

	inviter := dialWS(t, inviterID)
	defer inviter.Close()
	invitee := dialWS(t, inviteeID)
	defer invitee.Close()

	invite := ws.SendInvitationPayload{
		InviteeID:    inviteeID,
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

	accept := ws.AcceptInvitationPayload{InviterID: inviterID}
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

	submission1 := ws.SubmissionPayload{
		Status:          "Compile Error",
		PassedTestCases: 3,
		TotalTestCases:  82,
		Runtime:         30,
		Memory:          5000,
		Language:        "c",
		Time:            time.Now(),
	}
	err = invitee.WriteJSON(ws.Message{
		Type:    ws.ClientMsgSubmission,
		Payload: ws.MarshalPayload(submission1),
	})
	require.NoError(t, err)

	opponentSubMsg := readMessage(t, inviter)
	require.Equal(t, ws.ServerMsgOpponentSubmission, opponentSubMsg.Type)

	var oppSub ws.OpponentSubmissionPayload
	require.NoError(t, json.Unmarshal(opponentSubMsg.Payload, &oppSub))

	game, err := services.GameManager.GetGame(p1.SessionID)
	require.NoError(t, err)
	require.Equal(t, 1, len(game.Submissions))
	require.Equal(t, inviteeID, game.Submissions[0].PlayerID)

	require.Equal(t, int64(inviteeID), oppSub.PlayerID)
	require.Equal(t, submission1.Status, oppSub.Status)
	require.Equal(t, submission1.PassedTestCases, oppSub.PassedTestCases)
	require.Equal(t, submission1.TotalTestCases, oppSub.TotalTestCases)
	require.Equal(t, submission1.Runtime, oppSub.Runtime)
	require.Equal(t, submission1.Memory, oppSub.Memory)
	require.Equal(t, submission1.Language, oppSub.Language)

	submission2 := ws.SubmissionPayload{
		Status:          "Accepted",
		PassedTestCases: 82,
		TotalTestCases:  82,
		Runtime:         70,
		Memory:          10000,
		Language:        "java",
		Time:            time.Now(),
	}
	err = inviter.WriteJSON(ws.Message{
		Type:    ws.ClientMsgSubmission,
		Payload: ws.MarshalPayload(submission2),
	})
	require.NoError(t, err)

	endMsg1 := readMessage(t, inviter)
	endMsg2 := readMessage(t, invitee)
	require.Equal(t, ws.ServerMsgGameOver, endMsg1.Type)
	require.Equal(t, ws.ServerMsgGameOver, endMsg2.Type)

	var endPayload1, endPayload2 ws.GameOverPayload
	require.NoError(t, json.Unmarshal(endMsg1.Payload, &endPayload1))
	require.NoError(t, json.Unmarshal(endMsg2.Payload, &endPayload2))

	require.Equal(t, endPayload1.WinnerID, int64(inviterID))
	require.Equal(t, endPayload1.WinnerID, endPayload2.WinnerID)
	require.Equal(t, endPayload1.SessionID, endPayload2.SessionID)
	require.Equal(t, endPayload1.Duration, endPayload2.Duration)

	game, err = services.GameManager.GetGame(endPayload1.SessionID)
	require.NoError(t, err)
	require.Equal(t, int64(inviterID), game.Winner)
	require.Equal(t, models.MatchWon, game.Status)
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
