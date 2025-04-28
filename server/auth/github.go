package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"leetcodeduels/config"
	"net/http"
	"net/url"
	"strings"
)

// COPIED FROM V1: Change to fit current style.

func OAuthExchangeToken(w http.ResponseWriter, r *http.Request) {
	// Parse JSON body
	var reqBody struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	code := reqBody.Code

	// Exchange code for access token
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	clientID := cfg.GITHUB_CLIENT_ID
	clientSecret := cfg.GITHUB_CLIENT_SECRET
	tokenURL := "https://github.com/login/oauth/access_token"

	formData := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		http.Error(w, "Bad code", http.StatusBadRequest)
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode response
	var tokenResponse struct {
		AccessToken      string `json:"access_token"`
		TokenType        string `json:"token_type"`
		Scope            string `json:"scope"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		ErrorURI         string `json:"error_uri"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &tokenResponse); err != nil {
		http.Error(w, "Failed to parse access token response", http.StatusInternalServerError)
		return
	}

	if tokenResponse.Error != "" {
		http.Error(w, fmt.Sprintf("Error exchanging code: %s - %s", tokenResponse.Error, tokenResponse.ErrorDescription), http.StatusBadRequest)
		return
	}

	if tokenResponse.AccessToken == "" {
		http.Error(w, "Access token not found in response", http.StatusInternalServerError)
		return
	}

	user, err := fetchGitHubUser(tokenResponse.AccessToken)
	if err != nil {
		http.Error(w, "Failed to fetch GitHub user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = store.DataStore.SaveOAuthUser(user.ID, user.Username, tokenResponse.AccessToken)
	if err != nil {
		http.Error(w, "Failed to save OAuth user", http.StatusInternalServerError)
		return
	}

}
