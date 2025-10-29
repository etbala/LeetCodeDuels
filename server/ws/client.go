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
	readTimeout    = 60 * time.Second
	writeWait      = 10 * time.Second
)

type Client struct {
	userID int64
	ctx    context.Context
	conn   *websocket.Conn
	send   chan []byte
	hub    *connManager
	log    *zerolog.Logger
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
		log:    l,
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		c.log.Info().Msg("Client disconnected, stopping read pump")
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(readTimeout))

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			c.log.Info().Err(err).Msg("Websocket closed")
			break
		}

		c.conn.SetReadDeadline(time.Now().Add(readTimeout))

		c.log.Debug().Bytes("raw_message", raw).Msg("Received message from client")

		var env Message
		err = json.Unmarshal(raw, &env)
		if err != nil {
			c.log.Warn().Err(err).Bytes("raw_message", raw).Msg("Failed to unmarshal message from client")
			c.sendError("invalid_message_format", "could not parse message envelope")
			continue
		}

		err = c.hub.HandleClientMessage(c, &env)
		if err != nil {
			c.log.Error().Err(err).Str("message_type", string(env.Type)).Msg("Error handling client message")
			c.sendError("handler_error", err.Error())
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for message := range c.send {
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}

	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (c *Client) sendError(code, msg string) {
	c.log.Warn().Str("error_code", code).Str("error_msg", msg).Msg("Sending error to client")

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
