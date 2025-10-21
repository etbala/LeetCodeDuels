package api

import (
	"net/http"

	"leetcodeduels/handlers"
	"leetcodeduels/logging"
	"leetcodeduels/ws"

	"github.com/gorilla/mux"
)

// SetupRoutes initializes and returns the main router with all route groups and middleware set up.
func SetupRoutes(authMiddleware mux.MiddlewareFunc) *mux.Router {
	r := mux.NewRouter()
	r.Use(logging.RequestLogger)
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
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
	accountRouter := api.PathPrefix("/v1/users").Subrouter()
	accountRouter.Use(authMiddleware)

	// GET /users?username={username}&discriminator={discriminator}&limit={limit}
	// Returns a list of users with a matching username (and discriminator if provided).
	// Query Parameters:
	// - username: The username to search for (required)
	// - discriminator: The 4-digit discriminator (optional, for exact matches)
	// - limit: Maximum number of results to return (default 5, max 20)
	// Response: []models.UserInfoResponse
	accountRouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		handlers.SearchUsers(w, r)
	}).Methods("GET")

	// GET /users/me
	// Returns the current authenticated user's profile information.
	// Response: models.UserInfoResponse
	accountRouter.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		handlers.MyProfile(w, r)
	}).Methods("GET")

	// DELETE /users/me
	// Deletes the current authenticated user's account.
	accountRouter.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteUser(w, r)
	}).Methods("DELETE")

	// PATCH /users/me
	// Changes the current authenticated user's display name.
	// Request: models.UpdateUserRequest
	// Response: models.UpdateUserResponse
	accountRouter.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateUser(w, r)
	}).Methods("PATCH")

	// GET /users/me/notifications
	// Returns a list of notifications for the current authenticated user.
	// Note: For now just returns list of pending duel invites.
	// Response: []models.Notification
	accountRouter.HandleFunc("/me/notifications", func(w http.ResponseWriter, r *http.Request) {
		handlers.MyNotifications(w, r)
	}).Methods("GET")

	// GET /users/{id}
	// Returns the public profile information for a user by their user ID.
	// Response: models.UserInfoResponse
	accountRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetProfile(w, r)
	}).Methods("GET")

	// GET /users/{id}/status
	// Returns whether a user is currently online, offline, or in a game (and game ID if in-game).
	// Response: models.UserStatusResponse
	accountRouter.HandleFunc("/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		handlers.UserStatus(w, r)
	}).Methods("GET")

	// GET /users/{id}/matches?page={page_num}&limit={limit}
	// Returns a list of recent matches for a user by their user ID.
	// Query Parameters:
	// - page_num: The page number for pagination (default 1)
	// - limit: Maximum number of results per page (default 10, max 50)
	// Response: []models.Session
	accountRouter.HandleFunc("/{id}/matches", func(w http.ResponseWriter, r *http.Request) {
		handlers.UserMatches(w, r)
	}).Methods("GET")

	// ----------------------
	// Match Invite Routes
	// ----------------------
	inviteRouter := api.PathPrefix("/v1/invites").Subrouter()
	inviteRouter.Use(authMiddleware)

	// GET /invites/can_send
	// Checks if the current user can send a match invite.
	// Response: models.CanSendInviteResponse
	inviteRouter.HandleFunc("/can_send", func(w http.ResponseWriter, r *http.Request) {
		handlers.CanSendInvite(w, r)
	}).Methods("GET")

	// GET /invites/sent
	// Returns a list of match invites sent by the current user.
	// Response: []models.Invite
	inviteRouter.HandleFunc("/sent", func(w http.ResponseWriter, r *http.Request) {
		handlers.SentInvites(w, r)
	}).Methods("GET")

	// GET /invites/received
	// Returns a list of match invites received by the current user.
	// Response: []models.Invite
	inviteRouter.HandleFunc("/received", func(w http.ResponseWriter, r *http.Request) {
		handlers.ReceivedInvites(w, r)
	}).Methods("GET")

	// ----------------------
	// Matchmaking Routes
	// ----------------------
	mmRouter := api.PathPrefix("/v1/queue").Subrouter()
	mmRouter.Use(authMiddleware)

	// GET /queue/size
	// Returns the current number of users in the matchmaking queue.
	// Response: models.QueueSizeResponse
	mmRouter.HandleFunc("/size", func(w http.ResponseWriter, r *http.Request) {
		handlers.QueueSize(w, r)
	}).Methods("GET")

	// ----------------------
	// Game Session Routes
	// ----------------------
	matchRouter := api.PathPrefix("/v1/matches").Subrouter()
	matchRouter.Use(authMiddleware)

	// GET /matches/{id}
	// Returns detailed information about a specific match by its ID.
	// Response: models.Session
	matchRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchesGet(w, r)
	}).Methods("GET")

	// GET /matches/{id}/submissions
	// Returns a list of submissions for a specific match by its ID.
	// Response: []models.PlayerSubmission
	matchRouter.HandleFunc("/{id}/submissions", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchSubmissions(w, r)
	}).Methods("GET")

	// --------------------
	// Problems Routes
	// --------------------
	problemRouter := api.PathPrefix("/v1/problems").Subrouter()

	// GET /problems/random?difficulty[]=Easy&difficulty[]=Medium&tag[]=1&tag[]=2
	// Returns a random problem matching the specified criteria.
	// Query Parameters:
	// - difficulty[]: Difficulties (Easy, Medium, and/or Hard)
	// - tag[]: Filtered Tag IDs (problem will have at least one of these tags)
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

	// ----------------------
	// WebSocket Ticket Route
	// ----------------------
	wsTicketRouter := api.PathPrefix("/v1/ws-ticket").Subrouter()
	wsTicketRouter.Use(authMiddleware)
	wsTicketRouter.HandleFunc("", ws.GenerateWSTicket).Methods("POST")

	// --------------------
	// WebSocket Upgrader
	// --------------------
	wsRouter := r.PathPrefix("/ws").Subrouter()

	wsRouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		ws.WSConnect(w, r)
	}).Methods("GET")

	return r
}
