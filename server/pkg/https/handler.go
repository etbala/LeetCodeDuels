package https

/*
Contains the HTTP handlers that use the store interface to interact with the
database and return the data to the client. This is where you would
*/

import (
	"encoding/json"
	"leetcodeduels/api/game"
	"leetcodeduels/pkg/store/db"
	"net/http"
	"strconv"
)

// Handler struct holds dependencies for HTTP handlers, e.g., the data store.
type Handler struct {
	store *db.Store
}

// NewHandler creates a new HTTP handler with dependencies.
func NewHandler(store *db.Store) *Handler {
	return &Handler{store: store}
}

// GetAllProblems handles the request to get all problems.
func (h *Handler) GetAllProblems(w http.ResponseWriter, r *http.Request) {
	problems, err := h.store.GetAllProblems()
	if err != nil {
		http.Error(w, "Failed to fetch problems: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, problems)
}

// GetRandomProblem handles the request to get a random problem.
func (h *Handler) GetRandomProblem(w http.ResponseWriter, r *http.Request) {
	problem, err := h.store.GetRandomProblem()
	if err != nil {
		http.Error(w, "Failed to fetch a random problem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, problem)
}

// GetProblemsByTag handles the request to get problems by a specific tag.
func (h *Handler) GetProblemsByTag(w http.ResponseWriter, r *http.Request) {
	tagIDStr := r.URL.Query().Get("tag_id")
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	problems, err := h.store.GetProblemsByTag(tagID)
	if err != nil {
		http.Error(w, "Failed to fetch problems by tag", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, problems)
}

// GetRandomProblemByTag handles the request to get a random problem by a specific tag.
func (h *Handler) GetRandomProblemByTag(w http.ResponseWriter, r *http.Request) {
	tagIDStr := r.URL.Query().Get("tag_id")
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	problem, err := h.store.GetRandomProblemByTag(tagID)
	if err != nil {
		http.Error(w, "Failed to fetch a random problem by tag", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, problem)
}

// GetAllTags handles the request to get all tags.
func (h *Handler) GetAllTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.store.GetAllTags()
	if err != nil {
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, tags)
}

// GetTagsByProblem handles the request to get tags for a specific problem.
func (h *Handler) GetTagsByProblem(w http.ResponseWriter, r *http.Request) {
	problemIDStr := r.URL.Query().Get("problem_id")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		http.Error(w, "Invalid problem ID", http.StatusBadRequest)
		return
	}

	tags, err := h.store.GetTagsByProblem(problemID)
	if err != nil {
		http.Error(w, "Failed to fetch tags for the problem", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, tags)
}

// Handles Login Attempts
func (h *Handler) AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	// Assuming this should only be called with POST for security reasons
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("user")
	password := r.FormValue("pass")

	success, err := h.store.AuthenticateUser(username, password)
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

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
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
	success, err := h.store.CreateUser(w, username, password, email)
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

func (h *Handler) IsUserInGame(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")

	gameManager := game.GetGameManager()
	success := gameManager.IsPlayerInSession(userID)

	respondWithJSON(w, http.StatusOK, success)

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
