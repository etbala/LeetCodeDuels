package handlers

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func QueueSize(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())
	l.Warn().Msg("Unimplemented Endpoint Called")
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}
