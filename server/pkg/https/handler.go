package https

/*
Contains the HTTP handlers that use the DataStore to interact with the
database and return the data to the client.
*/

import (
	"encoding/json"
	"fmt"
	"io"
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

func OAuthExchangeToken(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON body
	var reqBody struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	code := reqBody.Code

	// Proceed to exchange the code for an access token
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
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode the response and handle errors
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

	err = store.DataStore.SaveOAuthUser(user.GithubID, user.Username, tokenResponse.AccessToken)
	if err != nil {
		http.Error(w, "Failed to save OAuth user", http.StatusInternalServerError)
		return
	}

	tokenString, err := auth.GenerateJWT(user.GithubID, user.Username)
	if err != nil {
		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	user, err := store.DataStore.GetUserProfile()
	if err != nil {
		http.Error(w, "Failed to fetch user profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil || user.GithubID == nil {
		http.Error(w, "Specified User Does Not Exist", http.NotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func AddPlayerToPool(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")

	gameManager := game.GetGameManager()
	success := gameManager.IsPlayerInSession(userID)

	respondWithJSON(w, http.StatusOK, success)
}

func RemovePlayerFromPool(w http.ResponseWriter, r *http.Request) {

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

func IsUserInGame(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")

	gameManager := game.GetGameManager()
	success := gameManager.IsPlayerInSession(userID)

	respondWithJSON(w, http.StatusOK, success)
}

func AddSubmission(w http.ResponseWriter, r *http.Request) {

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
