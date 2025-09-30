package handlers

import (
	"encoding/json"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func MatchesGet(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	vars := mux.Vars(r)
	matchID := vars["id"]

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("match_id", matchID)
	})
	l.Info().Msg("Received request for MatchesGet")

	uuid, err := uuid.Parse(matchID)
	if err != nil {
		l.Warn().Err(err).Msg("Invalid match ID format in path parameter")
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	session, err := services.GameManager.GetGame(matchID)
	if err != nil {
		l.Error().Err(err).Msg("Error checking for active game")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if session != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)
		return
	}

	session, err = store.DataStore.GetMatch(uuid)
	if err != nil {
		l.Error().Err(err).Msg("Failed to get match from datastore")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Match Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func MatchSubmissions(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	vars := mux.Vars(r)
	matchID := vars["id"]

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("match_id", matchID)
	})
	l.Info().Msg("Received request for MatchSubmissions")

	uuid, err := uuid.Parse(matchID)
	if err != nil {
		l.Warn().Err(err).Msg("Invalid match ID format in path parameter")
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	session, err := services.GameManager.GetGame(matchID)
	if err != nil {
		l.Error().Err(err).Msg("Error checking for active game")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if session != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session.Submissions)
		return
	}

	session, err = store.DataStore.GetMatch(uuid)
	if err != nil {
		l.Error().Err(err).Msg("Failed to get match from datastore")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Match Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.Submissions)
}
