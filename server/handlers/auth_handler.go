package handlers

import (
	"encoding/json"
	"fmt"
	"leetcodeduels/auth"
	"leetcodeduels/config"
	"leetcodeduels/services"
	"net/http"
)

func AuthGitHubInitiate(w http.ResponseWriter, r *http.Request) {
	state, err := auth.GenerateState()
	if err != nil {
		http.Error(w, "could not generate state", http.StatusInternalServerError)
		return
	}

	err = auth.StateStore.StoreState(r.Context(), state)
	if err != nil {
		http.Error(w, "could not save state", http.StatusInternalServerError)
		return
	}

	cfg, _ := config.LoadConfig()
	url := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&state=%s",
		cfg.GITHUB_CLIENT_ID, state,
	)
	http.Redirect(w, r, url, http.StatusFound)
}

func AuthGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	valid, err := auth.StateStore.ValidateState(r.Context(), state)
	if err != nil {
		http.Error(w, "could not validate state", http.StatusInternalServerError)
		return
	}
	if !valid {
		http.Error(w, "Invalid state", http.StatusUnauthorized)
		return
	}

	user, err := services.ExchangeCodeForUser(code)
	if err != nil {
		http.Error(w, "Invalid code", http.StatusBadRequest)
		return
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
