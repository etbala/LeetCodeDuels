package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"leetcodeduels/auth"
	"leetcodeduels/config"
	"leetcodeduels/services"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

// generateState returns a URL-safe random string.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// AuthGitHubInitiate stores a random state in Redis (TTL 5m) then redirects.
func AuthGitHubInitiate(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := generateState()
		if err != nil {
			http.Error(w, "could not generate state", http.StatusInternalServerError)
			return
		}

		// store `state` → "1" for 5 minutes
		if err := rdb.Set(r.Context(), state, "1", 5*time.Minute).Err(); err != nil {
			http.Error(w, "could not save oauth state", http.StatusInternalServerError)
			return
		}

		// build GitHub auth URL
		cfg, _ := config.LoadConfig()
		url := fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&state=%s",
			cfg.GITHUB_CLIENT_ID, state,
		)
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func AuthGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	user, err := services.ExchangeCodeForUser(code, state)
	if err != nil {
		http.Error(w, "Invalid code or state", http.StatusBadRequest)
		return
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
