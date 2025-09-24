package services

import (
	"encoding/json"
	"fmt"
	"io"
	"leetcodeduels/config"
	"leetcodeduels/models"
	"leetcodeduels/store"
	"net/http"
	"net/url"
	"strings"
)

// ExchangeCodeForUser handles the full OAuth + upsert flow, purely functional.
func ExchangeCodeForUser(code string) (*models.User, error) {
	// load client credentials
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	token, err := exchangeCode(cfg.GITHUB_CLIENT_ID, cfg.GITHUB_CLIENT_SECRET, code)
	if err != nil {
		return nil, err
	}

	ghID, username, avatar_url, err := fetchGitHubUser(token)
	if err != nil {
		return nil, err
	}

	existing, err := store.DataStore.GetUserProfile(ghID)
	if err != nil {
		return nil, err
	}

	var user models.User
	if existing != nil {
		// update their token and return
		if err := store.DataStore.UpdateGithubAccessToken(ghID, token); err != nil {
			return nil, err
		}
		user = *existing
		user.AccessToken = token
	} else {
		discriminator, err := GenerateUniqueDiscriminator(username)
		if err != nil {
			return nil, err
		}

		if err := store.DataStore.SaveOAuthUser(ghID, token, username, discriminator, avatar_url); err != nil {
			return nil, err
		}
		created, err := store.DataStore.GetUserProfile(ghID)
		if err != nil {
			return nil, err
		}
		user = *created
	}

	return &user, nil
}

// exchangeCode calls GitHub’s /login/oauth/access_token to get a bearer token.
func exchangeCode(clientID, clientSecret, code string) (string, error) {
	form := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
	}
	req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	var tr struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		// …
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}
	if tr.Error != "" || tr.AccessToken == "" {
		return "", fmt.Errorf("oauth error: %s", tr.Error)
	}
	return tr.AccessToken, nil
}

// fetchGitHubUser fetches GitHub's /user endpoint using the bearer token.
func fetchGitHubUser(token string) (int64, string, string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", "", fmt.Errorf("fetch github user: %w", err)
	}
	defer resp.Body.Close()

	var gh struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gh); err != nil {
		return 0, "", "", fmt.Errorf("decode github user: %w", err)
	}
	return gh.ID, gh.Login, gh.AvatarURL, nil
}
