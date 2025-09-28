package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	maxMessageSize = 2048
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	writeWait      = 10 * time.Second
)

type Client struct {
	userID int64
	ctx    context.Context
	conn   *websocket.Conn
	send   chan []byte
	hub    *connManager
}

func NewClient(
	userID int64,
	ctx context.Context,
	conn *websocket.Conn,
	hub *connManager,
	l *zerolog.Logger,
) *Client {
	return &Client{
		userID: userID,
		ctx:    ctx,
		conn:   conn,
		send:   make(chan []byte, 256),
		hub:    hub,
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var env Message
		err = json.Unmarshal(raw, &env)
		if err != nil {
			c.sendError("invalid_message_format", "could not parse message envelope")
			continue
		}

		err = c.hub.HandleClientMessage(c, &env)
		if err != nil {
			c.sendError("handler_error", err.Error())
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// connManager closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendError(code, msg string) {
	payload, _ := json.Marshal(ErrorPayload{Code: code, Message: msg})

	errEnv := Message{
		Type:    ServerMsgError,
		Payload: payload,
	}
	raw, _ := json.Marshal(errEnv)
	select {
	case c.send <- raw:
	default:
		// drop if send buffer is full
	}
}
