package handlers

import (
	"encoding/json"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func MatchesGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["id"]

	session, err := services.GameManager.GetGame(matchID)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if session != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)
	}

	session, err = store.DataStore.GetMatch(matchID)
	if err != nil {
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
	vars := mux.Vars(r)
	matchID := vars["id"]

	session, err := services.GameManager.GetGame(matchID)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if session != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session.Submissions)
	}

	session, err = store.DataStore.GetMatch(matchID)
	if err != nil {
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

func MatchHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	sessions, err := store.DataStore.GetPlayerMatches(userID)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}
