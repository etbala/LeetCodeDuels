package handlers

import (
	"encoding/json"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"net/http"

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

func MatchHistory(w http.ResponseWriter, r *http.Request) {

}
