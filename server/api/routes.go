package api

import (
	"net/http"

	"leetcodeduels/handlers"

	"github.com/gorilla/mux"
)

// SetupRoutes initializes and returns the main router with all route groups and middleware set up.
func SetupRoutes(authMiddleware mux.MiddlewareFunc) *mux.Router {
	r := mux.NewRouter()

	// ----------------------
	// Authentication Routes
	// ----------------------
	authRouter := r.PathPrefix("/auth").Subrouter()

	authRouter.HandleFunc("/github/initiate", func(w http.ResponseWriter, r *http.Request) {
		// handlers.AuthGitHubInitiate(w, r)
	}).Methods("GET")

	authRouter.HandleFunc("/github/callback", func(w http.ResponseWriter, r *http.Request) {
		// handlers.AuthGitHubCallback(w, r)
	}).Methods("GET")

	// ----------------------
	// User / Account Routes
	// ----------------------
	accountRouter := r.PathPrefix("/user").Subrouter()
	accountRouter.Use(authMiddleware)

	// TODO: Friend System (friend invites & notification for invites system)

	accountRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		// handlers.GetProfile(w, r)
	}).Methods("GET")

	accountRouter.HandleFunc("/{id}/in-game", func(w http.ResponseWriter, r *http.Request) {
		// handlers.UserInGame(w, r)
	}).Methods("GET")

	// Get current user's profile information
	accountRouter.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		// handlers.MyProfile(w, r)
	}).Methods("GET")

	accountRouter.HandleFunc("/me/onboard", func(w http.ResponseWriter, r *http.Request) {
		// handlers.CreateUser(w, r)
	}).Methods("GET")

	accountRouter.HandleFunc("/me/delete", func(w http.ResponseWriter, r *http.Request) {
		// handlers.DeleteUser(w, r)
	}).Methods("GET")

	// Change username
	accountRouter.HandleFunc("/me/rename", func(w http.ResponseWriter, r *http.Request) {
		// handlers.RenameUser(w, r)
	}).Methods("GET")

	// Change associated leetcode username
	accountRouter.HandleFunc("/me/lcrename", func(w http.ResponseWriter, r *http.Request) {
		// handlers.RenameLCUser(w, r)
	}).Methods("GET")

	// ----------------------
	// Invitation Routes
	// ----------------------
	inviteRouter := r.PathPrefix("/invites").Subrouter()
	inviteRouter.Use(authMiddleware)

	inviteRouter.HandleFunc("/invite", func(w http.ResponseWriter, r *http.Request) {
		// handlers.SendInvite(w, r)
	}).Methods("POST")

	inviteRouter.HandleFunc("/invite/cancel", func(w http.ResponseWriter, r *http.Request) {
		// handlers.CancelInvite(w, r)
	}).Methods("POST")

	inviteRouter.HandleFunc("/invite/{id}/accept", func(w http.ResponseWriter, r *http.Request) {
		// handlers.AcceptInvite(w, r)
	}).Methods("POST")

	inviteRouter.HandleFunc("/invite/{id}/decline", func(w http.ResponseWriter, r *http.Request) {
		// handlers.DeclineInvite(w, r)
	}).Methods("POST")

	// ----------------------
	// Matchmaking Routes
	// ----------------------
	mmRouter := r.PathPrefix("/matchmake").Subrouter()
	mmRouter.Use(authMiddleware)

	// TODO: Figure out nice way to pass all params through /enter

	mmRouter.HandleFunc("/enter", func(w http.ResponseWriter, r *http.Request) {
		// handlers.EnterQueue(w, r)
		http.Error(w, "Not yet implemented", http.StatusNotImplemented)
	}).Methods("POST")

	mmRouter.HandleFunc("/leave", func(w http.ResponseWriter, r *http.Request) {
		// handlers.LeaveQueue(w, r)
		http.Error(w, "Not yet implemented", http.StatusNotImplemented)
	}).Methods("POST")

	// ----------------------
	// Game Session Routes
	// ----------------------
	matchRouter := r.PathPrefix("/game").Subrouter()
	matchRouter.Use(authMiddleware)

	// Get match details
	matchRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		// handlers.MatchesGet(w, r)
	}).Methods("GET")

	// Send code submission results/details
	matchRouter.HandleFunc("/{id}/submission", func(w http.ResponseWriter, r *http.Request) {
		// handlers.MatchesSubmit(w, r)
	}).Methods("POST")

	// --------------------
	// Problems Routes
	// --------------------
	problemRouter := r.PathPrefix("/problems").Subrouter()
	problemRouter.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		// handlers.ProblemsRandom(w, r)
	}).Methods("GET")

	problemRouter.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		// handlers.AllTags(w, r)
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
