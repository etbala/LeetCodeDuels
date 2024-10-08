package https

/*
Contains the HTTP handlers that use the DataStore to interact with the
database and return the data to the client.
*/

import (
	"encoding/json"
	"leetcodeduels/api/game"
	"leetcodeduels/pkg/store"
	"net/http"
	"strconv"
	"time"
)

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

	statusStr := r.URL.Query().Get("Status")
	submissionIDStr := r.URL.Query().Get("SubmissionID")
	playerUUID := r.URL.Query().Get("PlayerUUID")
	runtimeStr := r.URL.Query().Get("Runtime")
	memoryStr := r.URL.Query().Get("Memory")
	langStr := r.URL.Query().Get("Lang")
	passedTestCasesStr := r.URL.Query().Get("PassedTestCases")
	totalTestCasesStr := r.URL.Query().Get("TotalTestCases")

	status, err := game.ParseSubmissionStatus(statusStr)
	if err != nil {
		http.Error(w, "Invalid submission status", http.StatusBadRequest)
		return
	}

	lang, err := game.ParseLang(langStr)
	if err != nil {
		http.Error(w, "Invalid language", http.StatusBadRequest)
		return
	}

	submissionID, err := strconv.Atoi(submissionIDStr)
	if err != nil {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	runtime, err := strconv.Atoi(runtimeStr)
	if err != nil {
		http.Error(w, "Invalid runtime value", http.StatusBadRequest)
		return
	}

	memory, err := strconv.Atoi(memoryStr)
	if err != nil {
		http.Error(w, "Invalid memory value", http.StatusBadRequest)
		return
	}

	passedTestCases, err := strconv.Atoi(passedTestCasesStr)
	if err != nil {
		http.Error(w, "Invalid passed test cases", http.StatusBadRequest)
		return
	}

	totalTestCases, err := strconv.Atoi(totalTestCasesStr)
	if err != nil {
		http.Error(w, "Invalid total test cases", http.StatusBadRequest)
		return
	}

	submission := game.PlayerSubmission{
		ID:              submissionID,
		PlayerUUID:      playerUUID,
		PassedTestCases: passedTestCases,
		TotalTestCases:  totalTestCases,
		Status:          status,
		Runtime:         runtime,
		Memory:          memory,
		Time:            time.Now(),
		Lang:            lang,
	}

	gm := game.GetGameManager()
	gm.AddSubmission(playerUUID, submission)
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
