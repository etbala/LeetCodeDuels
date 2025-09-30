package ws

import (
	"leetcodeduels/services"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WSConnect(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	claims, err := services.GetClaimsFromRequest(r)
	if err != nil {
		l.Warn().Msg("Unauthorized attempt to open WebSocket connection")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Int64("user_id", claims.UserID)
	})

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		l.Error().Err(err).Msg("Failed to upgrade connection to WebSocket")
		return
	}

	l.Info().Msg("WebSocket connection established")

	userID := claims.UserID

	client := NewClient(userID, r.Context(), conn, ConnManager, l)
	ConnManager.register <- client
	go client.writePump()
	go client.readPump()
}
