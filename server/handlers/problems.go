package handlers

import (
	"encoding/json"
	"leetcodeduels/store"
	"net/http"
)

func RandomProblem(w http.ResponseWriter, r *http.Request) {

}

func AllTags(w http.ResponseWriter, r *http.Request) {
	tags, err := store.DataStore.GetAllTags()
	if err != nil {
		http.Error(w, "could not retrieve tags", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}
