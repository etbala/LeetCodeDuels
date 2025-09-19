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
		"https://github.com/login/oauth/authorize?client_id=%s&state=%s",
		cfg.GITHUB_CLIENT_ID,
		escapedState,
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func AuthGitHubCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<html><body><h1>Success!</h1><p>You can close this window now.</p></body></html>")
}

func AuthGitHubExchange(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if payload.Code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	user, err := services.ExchangeCodeForUser(payload.Code)
	if err != nil {
		log.Printf("GitHub token exchange error: %v", err)
		http.Error(w, "Invalid code", http.StatusBadRequest)
		return
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("JWT generation error: %v", err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
