package services

import (
	"encoding/json"
	"fmt"
	"io"
	"leetcodeduels/config"
	"leetcodeduels/database"
	"leetcodeduels/models"
	"net/http"
	"net/url"
	"strings"
)

// ExchangeCodeForUser handles the full OAuth + upsert flow, purely functional.
func ExchangeCodeForUser(db database.I_DB, code, state string) (*models.User, error) {
	// load client credentials
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// 1) swap code → access token
	token, err := exchangeCode(cfg.GITHUB_CLIENT_ID, cfg.GITHUB_CLIENT_SECRET, code)
	if err != nil {
		return nil, err
	}

	// 2) fetch GitHub user (ID + login)
	ghID, login, err := fetchGitHubUser(token)
	if err != nil {
		return nil, err
	}

	// 3) see if user already exists
	existing, err := db.GetUserProfile(ghID)
	if err != nil {
		// if your DB layer returns ErrNotFound, you could check errors.Is(err, database.ErrNotFound)
		// here—but for simplicity we assume err!=nil is always a hard error.
		return nil, err
	}

	var user models.User
	if existing != nil {
		// update their token and return
		if err := db.UpdateGithubAccessToken(ghID, token); err != nil {
			return nil, err
		}
		user = *existing
		user.AccessToken = token
	} else {
		// new user: save and then re-fetch to get full fields
		if err := db.SaveOAuthUser(ghID, login, token); err != nil {
			return nil, err
		}
		created, err := db.GetUserProfile(ghID)
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
func fetchGitHubUser(token string) (int64, string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("fetch github user: %w", err)
	}
	defer resp.Body.Close()

	var gh struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gh); err != nil {
		return 0, "", fmt.Errorf("decode github user: %w", err)
	}
	return gh.ID, gh.Login, nil
}
