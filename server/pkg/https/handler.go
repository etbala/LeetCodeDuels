package https

/*
Contains the HTTP handlers that use the DataStore to interact with the
database and return the data to the client.
*/

import (
	"encoding/json"
	"fmt"
	"leetcodeduels/api/auth"
	"leetcodeduels/api/game"
	"leetcodeduels/pkg/config"
	"leetcodeduels/pkg/models"
	"leetcodeduels/pkg/store"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func OAuthAuthorize(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	clientID := cfg.GITHUB_CLIENT_ID
	redirectURI := cfg.GITHUB_REDIRECT_URI

	// Generate a random string to prevent CSRF
	stateStore := auth.GetStateStore()
	state, err := stateStore.GenerateRandomState()
	if err != nil {
		return
	}

	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s&scope=user",
		clientID, redirectURI, state,
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func fetchGitHubUser(accessToken string) (*models.OAuthUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user: status %d", resp.StatusCode)
	}

	var user struct {
		ID       int64  `json:"id"`
		Username string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &models.OAuthUser{
		GithubID: user.ID,
		Username: user.Username,
	}, nil
}

func OAuthCallback(w http.ResponseWriter, r *http.Request) {
	stateStore := auth.GetStateStore()

	// Verify state for CSRF protection
	state := r.URL.Query().Get("state")
	err := stateStore.ValidateState(state)
	if err != nil {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Get the authorization code from the query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Exchange code for an access token
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
		http.Error(w, "Failed to create token request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode the response
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		http.Error(w, "Failed to parse access token response", http.StatusInternalServerError)
		return
	}

	user, err := fetchGitHubUser(tokenResponse.AccessToken)
	if err != nil {
		http.Error(w, "Failed to fetch GitHub user", http.StatusInternalServerError)
		return
	}

	err = store.DataStore.SaveOAuthUser(user.GithubID, user.Username, tokenResponse.AccessToken)
	if err != nil {
		http.Error(w, "Failed to save OAuth user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Login successful")
}

// GetAllProblems handles the request to get all problems.
func GetAllProblems(w http.ResponseWriter, r *http.Request) {
	problems, err := store.DataStore.GetAllProblems()
	if err != nil {
		http.Error(w, "Failed to fetch problems: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, problems)
}

// GetRandomProblem handles the request to get a random problem.
func GetRandomProblem(w http.ResponseWriter, r *http.Request) {
	problem, err := store.DataStore.GetRandomProblem()
	if err != nil {
		http.Error(w, "Failed to fetch a random problem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, problem)
}

// GetProblemsByTag handles the request to get problems by a specific tag.
func GetProblemsByTag(w http.ResponseWriter, r *http.Request) {
	tagIDStr := r.URL.Query().Get("tag_id")
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	problems, err := store.DataStore.GetProblemsByTag(tagID)
	if err != nil {
		http.Error(w, "Failed to fetch problems by tag", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, problems)
}

// GetRandomProblemByTag handles the request to get a random problem by a specific tag.
func GetRandomProblemByTag(w http.ResponseWriter, r *http.Request) {
	tagIDStr := r.URL.Query().Get("tag_id")
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	problem, err := store.DataStore.GetRandomProblemByTag(tagID)
	if err != nil {
		http.Error(w, "Failed to fetch a random problem by tag", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, problem)
}

// GetAllTags handles the request to get all tags.
func GetAllTags(w http.ResponseWriter, r *http.Request) {
	tags, err := store.DataStore.GetAllTags()
	if err != nil {
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, tags)
}

// GetTagsByProblem handles the request to get tags for a specific problem.
func GetTagsByProblem(w http.ResponseWriter, r *http.Request) {
	problemIDStr := r.URL.Query().Get("problem_id")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		http.Error(w, "Invalid problem ID", http.StatusBadRequest)
		return
	}

	tags, err := store.DataStore.GetTagsByProblem(problemID)
	if err != nil {
		http.Error(w, "Failed to fetch tags for the problem", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, tags)
}

// Handles Login Attempts
func AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	// Assuming this should only be called with POST for security reasons
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("user")
	password := r.FormValue("pass")

	success, err := store.DataStore.AuthenticateUser(username, password)
	if err != nil {
		// More specific error handling can be added here based on error type
		http.Error(w, "Error logging in: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !success {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Set a cookie to indicate successful authentication
	http.SetCookie(w, &http.Cookie{
		Name:     "loggedIn",
		Value:    "true",
		Path:     "/",
		HttpOnly: true,  // Protects against XSS attacks by not allowing JS access
		Secure:   true,  // Ensures cookie is sent over HTTPS
		MaxAge:   86400, // Expires after one day
	})

	// Respond with JSON on successful authentication
	respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	// Assume this is a POST request; handle method checking if necessary
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user information from form data instead of the query string
	username := r.FormValue("user")
	password := r.FormValue("pass")
	email := r.FormValue("email")

	// Create user and handle the response
	success, err := store.DataStore.CreateUser(w, username, password, email)
	if err != nil {
		http.Error(w, "Error creating user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !success {
		http.Error(w, "Failed to create user", http.StatusBadRequest)
		return
	}

	// Respond with JSON on successful creation
	respondWithJSON(w, http.StatusOK, map[string]bool{"success": success})
}

func IsUserInGame(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")

	gameManager := game.GetGameManager()
	success := gameManager.IsPlayerInSession(userID)

	respondWithJSON(w, http.StatusOK, success)

}

func AddSubmission(w http.ResponseWriter, r *http.Request) {

	// TODO: Verify api sender is authorized to submit for sent UUID
	// TODO: Sanitize params

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var submissionData game.PlayerSubmission

	// Validate Status
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&submissionData); err != nil {
		log.Printf("Error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received Submission Data: %+v", submissionData)

	submissionData.Time = time.Now()

	gm := game.GetGameManager()
	gm.AddSubmission(submissionData.PlayerUUID, submissionData)

	w.WriteHeader(http.StatusOK)
}

// Helper function to respond with JSON.
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}
