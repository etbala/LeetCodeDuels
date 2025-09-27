package handlers

import (
	"encoding/json"
	"fmt"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func MatchesGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["id"]

	uuid, err := uuid.Parse(matchID)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	session, err := services.GameManager.GetGame(matchID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if session != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)
	}

	session, err = store.DataStore.GetMatch(uuid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Error: %s", err.Error()), http.StatusInternalServerError)
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
	vars := mux.Vars(r)
	matchID := vars["id"]

	uuid, err := uuid.Parse(matchID)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	session, err := services.GameManager.GetGame(matchID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if session != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session.Submissions)
		return
	}

	session, err = store.DataStore.GetMatch(uuid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Match Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.Submissions)
}
