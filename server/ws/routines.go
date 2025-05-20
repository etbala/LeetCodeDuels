package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// generic wrapper for incoming messages.
type messageEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// what we publish when we want another instance to drop a given connID.
type disconnectMsg struct {
	UserID int64  `json:"userID"`
	ConnID string `json:"connID"`
}

// reads from the socket, handles messages, and triggers cleanup on error.
func ReadLoop(cm *ConnectionManager, userID int64, connID string, conn *websocket.Conn, sendCh chan<- []byte) {
	// Set up Pong handler / deadlines
	const pongWait = 60 * time.Second
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	defer cleanup(cm, userID, connID, conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			// EOF or protocol error → exit loop
			return
		}

		var env messageEnvelope
		if err := json.Unmarshal(msg, &env); err != nil {
			// malformed → ignore or send an ERROR frame back
			continue
		}

		switch env.Type {
		case MessageTypeHeartbeat:

		case MessageTypeError:

		case MessageTypeSubmission:

		default:
			// unknown type
		}
	}
}

// pumps messages from sendCh to the socket, sends regular pings, and cleans up on error or channel close.
func WriteLoop(cm *ConnectionManager, userID int64, connID string, conn *websocket.Conn, sendCh <-chan []byte) {
	const (
		writeWait  = 10 * time.Second
		pingPeriod = 54 * time.Second // must be < pongWait
	)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		cleanup(cm, userID, connID, conn)
	}()

	for {
		select {
		case msg, ok := <-sendCh:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// channel closed → close socket
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			// send a ping
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// tells whichever WS server holds that oldConnID to tear it down.
func PublishDisconnect(cm *ConnectionManager, userID int64, oldConnID string) error {
	dm := disconnectMsg{
		UserID: userID,
		ConnID: oldConnID,
	}
	payload, err := json.Marshal(dm)
	if err != nil {
		return fmt.Errorf("marshal disconnect message: %w", err)
	}
	// publish on the shared "disconnect" channel
	return cm.client.Publish(context.Background(), "disconnect", payload).Err()
}

func cleanup(cm *ConnectionManager, userID int64, connID string, conn *websocket.Conn) {
	conn.Close()
	stillOnline, _ := cm.RemoveConnection(userID, connID)
	if !stillOnline {
		// user is now offline (broadcast here is necessary)
	}
}
