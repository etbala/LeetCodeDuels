package ws

import (
	"leetcodeduels/services"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// GenerateWSTicket creates a short-lived, single-use ticket for a WebSocket connection.
func GenerateWSTicket(w http.ResponseWriter, r *http.Request) {
	claims, err := services.GetClaimsFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ticket := uuid.New().String()
	err = ConnManager.StoreTicket(r.Context(), ticket, claims.UserID, 15*time.Second) // Ticket is valid for 15 seconds
	if err != nil {
		log.Error().Err(err).Msg("Failed to store WS ticket")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ticket":"` + ticket + `"}`))
}
