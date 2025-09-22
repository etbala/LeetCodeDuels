package handlers

import (
	"encoding/json"
	"fmt"
	"leetcodeduels/auth"
	"leetcodeduels/services"
	"log"
	"net/http"
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":          user.ID,
			"username":    user.Username,
			"lc_username": user.LeetCodeUsername,
		},
	})
}
