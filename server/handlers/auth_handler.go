package handlers

import (
	"encoding/json"
	"leetcodeduels/auth"
	"leetcodeduels/services"
	"net/http"
)

func AuthGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	user, err := services.ExchangeCodeForUser(code, state)
	if err != nil {
		http.Error(w, "Invalid code or state", http.StatusBadRequest)
		return
	}

	token, err := auth.GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
