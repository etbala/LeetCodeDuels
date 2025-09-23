package api

import (
	"net/http"

	"leetcodeduels/handlers"
	"leetcodeduels/ws"

	"github.com/gorilla/mux"
)

// SetupRoutes initializes and returns the main router with all route groups and middleware set up.
func SetupRoutes(authMiddleware mux.MiddlewareFunc) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")

	// ----------------------
	// Authentication Routes
	// ----------------------
	authRouter := r.PathPrefix("/auth").Subrouter()

	authRouter.HandleFunc("/github/callback", func(w http.ResponseWriter, r *http.Request) {
		handlers.AuthGitHubCallback(w, r)
	}).Methods("GET")

	authRouter.HandleFunc("/github/exchange", func(w http.ResponseWriter, r *http.Request) {
		handlers.AuthGitHubExchange(w, r)
	}).Methods("POST")

	// ----------------------
	// User / Account Routes
	// ----------------------
	accountRouter := r.PathPrefix("/user").Subrouter()
	accountRouter.Use(authMiddleware)

	// TODO: Add Friend System (friend invites & notification for invites system)

	// GET /user/me
	// Returns the current authenticated user's profile information.
	// Response: handlers.UserProfile
	accountRouter.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		handlers.MyProfile(w, r)
	}).Methods("GET")

	// POST /user/me/delete
	// Deletes the current authenticated user's account.
	// Response: None (Status OK on success)
	accountRouter.HandleFunc("/me/delete", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteUser(w, r)
	}).Methods("POST")

	// POST /user/me/rename
	// Changes the current authenticated user's display name.
	// Request: handlers.RenameRequest
	// Response: handlers.UserProfile
	accountRouter.HandleFunc("/me/rename", func(w http.ResponseWriter, r *http.Request) {
		handlers.RenameUser(w, r)
	}).Methods("POST")

	// POST /user/me/lcrename
	// Changes the current authenticated user's linked LeetCode username.
	// Request: handlers.RenameRequest
	// Response: handlers.UserProfile
	accountRouter.HandleFunc("/me/lcrename", func(w http.ResponseWriter, r *http.Request) {
		handlers.RenameLCUser(w, r)
	}).Methods("POST")

	// GET /user/me/notifications
	// Returns a list of notifications for the current authenticated user.
	// Response: []handlers.Notification (TODO)
	accountRouter.HandleFunc("/me/notifications", func(w http.ResponseWriter, r *http.Request) {
		// handlers.MyNotifications(w, r)
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}).Methods("GET")

	// GET /user/id/{username}
	// Returns the user ID for a given username.
	// Response: handlers.UserIDResponse (TODO)
	accountRouter.HandleFunc("/id/{username}", func(w http.ResponseWriter, r *http.Request) {
		// handlers.GetUserID(w, r) TODO: Implement
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}).Methods("GET")

	// GET /user/profile/{id}
	// Returns the public profile information for a user by their user ID.
	// Response: handlers.UserProfile
	accountRouter.HandleFunc("/profile/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetProfile(w, r)
	}).Methods("GET")

	// GET /user/is-online/{id}
	// Returns whether a user is currently online.
	// Response: handlers.UserOnlineResponse (TODO)
	accountRouter.HandleFunc("/is-online/{id}", func(w http.ResponseWriter, r *http.Request) {
		// handlers.UserOnline(w, r)
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}).Methods("GET")

	// GET /user/in-game/{id}
	// Returns whether a user is currently in a game.
	// Response: handlers.UserInGameResponse (TODO)
	accountRouter.HandleFunc("/in-game/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.UserInGame(w, r)
	}).Methods("GET")

	// GET /user/search?username={username}
	// Searches for users by username (supports partial matches).
	// Response: []handlers.UserProfile (TODO)
	accountRouter.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		// handlers.SearchUsers(w, r)
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}).Methods("GET")

	// ----------------------
	// Matchmaking Routes
	// ----------------------
	mmRouter := r.PathPrefix("/matchmake").Subrouter()
	mmRouter.Use(authMiddleware)

	// GET /matchmake/count
	// Returns the current number of users in the matchmaking queue.
	// Response: handlers.QueueSizeResponse (TODO)
	mmRouter.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
		handlers.QueueSize(w, r)
	}).Methods("GET")

	// ----------------------
	// Game Session Routes
	// ----------------------
	matchRouter := r.PathPrefix("/game").Subrouter()
	matchRouter.Use(authMiddleware)

	// GET /game/history/{id}
	// Returns a list of past matches for the specified user ID.
	// Response: []handlers.MatchSummary
	matchRouter.HandleFunc("/history/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchHistory(w, r)
	}).Methods("GET")

	// GET /game/details/{id}
	// Returns detailed information about a specific match by its ID.
	// Response: models.Session
	matchRouter.HandleFunc("/details/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchesGet(w, r)
	}).Methods("GET")

	// GET /game/submissions/{id}
	// Returns a list of submissions for a specific match by its ID.
	// Response: []models.PlayerSubmission
	matchRouter.HandleFunc("/submissions/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchSubmissions(w, r)
	}).Methods("GET")

	// --------------------
	// Problems Routes
	// --------------------
	problemRouter := r.PathPrefix("/problems").Subrouter()

	// GET /problems/random
	// Returns a random problem matching the specified criteria.
	// Query Parameters:
	// - difficulties: Comma-separated list of difficulties (Easy, Medium, Hard)
	// - tags: Comma-separated list of tag IDs
	// Response: models.Problem
	problemRouter.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		handlers.RandomProblem(w, r)
	}).Methods("GET")

	// GET /problems/tags
	// Returns a list of all available problem tags.
	// Response: []models.Tag
	problemRouter.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		handlers.AllTags(w, r)
	}).Methods("GET")

	// --------------------
	// WebSocket Upgrader
	// --------------------
	wsRouter := r.PathPrefix("/ws").Subrouter()
	wsRouter.Use(authMiddleware)

	wsRouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		ws.WSConnect(w, r)
	}).Methods("GET")

	return r
}
