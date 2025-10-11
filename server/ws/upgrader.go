package ws

import (
	"net/http"
	// remove "leetcodeduels/services" if no longer needed here

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WSConnect(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		l.Warn().Msg("Attempted to connect to WebSocket without a ticket")
		http.Error(w, "Bad Request: Missing ticket", http.StatusBadRequest)
		return
	}

	userID, err := ConnManager.ValidateTicket(r.Context(), ticket)
	if err != nil {
		l.Warn().Err(err).Str("ticket", ticket).Msg("Unauthorized attempt to open WebSocket connection with invalid ticket")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Int64("user_id", userID)
	})

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		l.Error().Err(err).Msg("Failed to upgrade connection to WebSocket")
		return
	}

	l.Info().Msg("WebSocket connection established")

	// todo: disconnect existing connection for this userID after sending other_logon message

	client := NewClient(userID, r.Context(), conn, ConnManager, l)
	ConnManager.register <- client
	go client.writePump()
	go client.readPump()
}
