package handlers

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func CanSendInvite(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())
	l.Warn().Msg("Unimplemented Endpoint Called")
	writeError(w, http.StatusNotImplemented, "Not Implemented")
}

func SentInvites(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())
	l.Warn().Msg("Unimplemented Endpoint Called")
	writeError(w, http.StatusNotImplemented, "Not Implemented")
}

func ReceivedInvites(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())
	l.Warn().Msg("Unimplemented Endpoint Called")
	writeError(w, http.StatusNotImplemented, "Not Implemented")
}
