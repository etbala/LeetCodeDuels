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

	authRouter.HandleFunc("/github/initiate", func(w http.ResponseWriter, r *http.Request) {
		handlers.AuthGitHubInitiate(w, r)
	}).Methods("GET")

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

	// TODO: Friend System (friend invites & notification for invites system)

	// Get current user's profile information
	accountRouter.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		handlers.MyProfile(w, r)
	}).Methods("GET")

	accountRouter.HandleFunc("/me/delete", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteUser(w, r)
	}).Methods("POST")

	// Change username
	accountRouter.HandleFunc("/me/rename", func(w http.ResponseWriter, r *http.Request) {
		handlers.RenameUser(w, r)
	}).Methods("POST")

	// Change associated leetcode username
	accountRouter.HandleFunc("/me/lcrename", func(w http.ResponseWriter, r *http.Request) {
		handlers.RenameLCUser(w, r)
	}).Methods("POST")

	accountRouter.HandleFunc("/id/{username}", func(w http.ResponseWriter, r *http.Request) {
		// handlers.GetUserID(w, r) TODO: Implement
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}).Methods("GET")

	accountRouter.HandleFunc("/profile/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetProfile(w, r)
	}).Methods("GET")

	accountRouter.HandleFunc("/in-game/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.UserInGame(w, r)
	}).Methods("GET")

	// ----------------------
	// Matchmaking Routes
	// ----------------------
	mmRouter := r.PathPrefix("/matchmake").Subrouter()
	mmRouter.Use(authMiddleware)

	// Number of people currently in queue
	mmRouter.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
		handlers.QueueSize(w, r)
	}).Methods("GET")

	// ----------------------
	// Game Session Routes
	// ----------------------
	matchRouter := r.PathPrefix("/game").Subrouter()
	matchRouter.Use(authMiddleware)

	matchRouter.HandleFunc("/history/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchHistory(w, r)
	}).Methods("GET")

	// Get match details
	matchRouter.HandleFunc("/details/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchesGet(w, r)
	}).Methods("GET")

	matchRouter.HandleFunc("/submissions/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchSubmissions(w, r)
	}).Methods("GET")

	// --------------------
	// Problems Routes
	// --------------------
	problemRouter := r.PathPrefix("/problems").Subrouter()
	problemRouter.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		handlers.RandomProblem(w, r)
	}).Methods("GET")

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
