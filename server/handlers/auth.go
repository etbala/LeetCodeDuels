package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"leetcodeduels/config"
	"leetcodeduels/models"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// todo: nicer way of doing this?
func AuthGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	errorParam := r.URL.Query().Get("error")

	var title, message, codeJS, errorJS string

	if errorParam != "" {
		title = "Authentication Failed"
		message = fmt.Sprintf("Error: %s", errorParam)
		errorJS = errorParam
		codeJS = ""
	} else if code != "" {
		title = "Authentication Successful!"
		message = "You can close this window now."
		codeJS = code
		errorJS = ""
	} else {
		title = "Authentication Error"
		message = "No authorization code received."
		codeJS = ""
		errorJS = "no_code"
	}

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>GitHub Authentication</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
        }
        .container {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            text-align: center;
            max-width: 400px;
        }
        .success { color: #10b981; }
        .error { color: #ef4444; }
        .spinner {
            border: 3px solid #f3f3f3;
            border-top: 3px solid #667eea;
            border-radius: 50%%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 20px auto;
        }
        @keyframes spin {
            0%% { transform: rotate(0deg); }
            100%% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>%s</h2>
        <div class="spinner"></div>
        <p>%s</p>
    </div>
    <script>
        // Post message to all possible extension origins
        const message = {
            type: 'github-oauth-callback',
            code: %q,
            error: %q
        };
        
        // Try posting to different extension origins
        const extensionOrigins = [
            'chrome-extension://*',
            'moz-extension://*',
            'safari-web-extension://*'
        ];
        
        // Post to opener (the extension tab that opened this)
        if (window.opener) {
            extensionOrigins.forEach(origin => {
                try {
                    window.opener.postMessage(message, origin);
                } catch (e) {
                    // Ignore errors for invalid origins
                }
            });
            
            // Also try wildcard for development
            window.opener.postMessage(message, '*');
        }
        
        // Auto-close after 2 seconds
        setTimeout(() => {
            window.close();
        }, 2000);
    </script>
</body>
</html>`, title, message, codeJS, errorJS)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, htmlContent)
}

func AuthGitHubExchange(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" {
		// todo: in production, validate against known extension IDs
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	}

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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

	user, err := exchangeCodeForUser(payload.Code)
	if err != nil {
		log.Printf("GitHub token exchange error: %v", err)
		http.Error(w, "Invalid code", http.StatusBadRequest)
		return
	}

	token, err := services.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("JWT generation error: %v", err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	var response = models.TokenExchangeResponse{
		Token: token,
		User: models.UserInfoResponse{
			ID:            user.ID,
			Username:      user.Username,
			Discriminator: user.Discriminator,
			LCUsername:    user.LeetCodeUsername,
			AvatarURL:     user.AvatarURL,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handles the full OAuth + upsert flow, purely functional.
func exchangeCodeForUser(code string) (*models.User, error) {
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
		discriminator, err := services.GenerateUniqueDiscriminator(username)
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
