package api

import (
	"net/http"

	"leetcodeduels/handlers"

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

	// ----------------------
	// User / Account Routes
	// ----------------------
	accountRouter := r.PathPrefix("/user").Subrouter()
	accountRouter.Use(authMiddleware)

	// TODO: Friend System (friend invites & notification for invites system)

	accountRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetProfile(w, r)
	}).Methods("GET")

	accountRouter.HandleFunc("/{id}/in-game", func(w http.ResponseWriter, r *http.Request) {
		handlers.UserInGame(w, r)
	}).Methods("GET")

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

	// Get match details
	matchRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchesGet(w, r)
	}).Methods("GET")

	matchRouter.HandleFunc("/{id}/submissions", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchSubmissions(w, r)
	}).Methods("GET")

	matchRouter.HandleFunc("/history/{userid}", func(w http.ResponseWriter, r *http.Request) {
		handlers.MatchHistory(w, r)
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
		handlers.WSConnect(w, r)
	}).Methods("GET")

	return r
}
