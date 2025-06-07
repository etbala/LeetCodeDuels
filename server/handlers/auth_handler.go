package handlers

import (
	"encoding/json"
	"fmt"
	"leetcodeduels/auth"
	"leetcodeduels/config"
	"leetcodeduels/services"
	"log"
	"net/http"
	"net/url"
)

func AuthGitHubInitiate(w http.ResponseWriter, r *http.Request) {
	state, err := auth.GenerateState()
	if err != nil {
		log.Printf("GenerateState error: %v", err)
		http.Error(w, "could not generate state", http.StatusInternalServerError)
		return
	}

	err = auth.StateStore.StoreState(r.Context(), state)
	if err != nil {
		log.Printf("StoreState error: %v", err)
		http.Error(w, "could not save state", http.StatusInternalServerError)
		return
	}

	cfg, _ := config.LoadConfig()

	escapedState := url.QueryEscape(state)
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s",
		cfg.GITHUB_CLIENT_ID,
		escapedState,
	)

	http.Redirect(w, r, authURL, http.StatusFound)
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
		http.Error(w, "invalid state", http.StatusUnauthorized)
		return
	}

	user, err := services.ExchangeCodeForUser(code)
	if err != nil {
		log.Printf("GitHub token exchange error: %v", err)
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("JWT generation error: %v", err)
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
