package api

import (
    "net/http"

    "github.com/gorilla/mux"
)

// SetupRoutes initializes and returns the main router with all route groups and middleware set up.
func SetupRoutes(authMiddleware mux.MiddlewareFunc) *mux.Router {
    r := mux.NewRouter()

    // --------------------
    // Authentication Routes
    // --------------------
    publicAuthRouter := r.PathPrefix("/auth").Subrouter()
	protectedAuthRouter := r.PathPrefix("/auth").Subrouter()
	protectedAuthRouter.Use(authMiddleware)

    // Public routes for initiating and handling GitHub OAuth
    publicAuthRouter.HandleFunc("/github/initiate", func(w http.ResponseWriter, r *http.Request) {
        // handlers.AuthGitHubInitiate(w, r)
    }).Methods("GET")

    publicAuthRouter.HandleFunc("/github/callback", func(w http.ResponseWriter, r *http.Request) {
        // handlers.AuthGitHubCallback(w, r)
    }).Methods("GET")

    // Authenticated endpoint to get current user info
    protectedAuthRouter.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
        // handlers.AuthMe(w, r)
    }).Methods("GET")


    // --------------------
    // Match & Game Session Routes
    // --------------------
    // All matches-related routes require authentication
	publicMatchRouter := r.PathPrefix("/matches").Subrouter()
    protectedMatchRouter := r.PathPrefix("/matches").Subrouter()
    matchRouter.Use(authMiddleware)

    // Invite and accept match
    protectedMatchRouter.HandleFunc("/invite", func(w http.ResponseWriter, r *http.Request) {
        // handlers.MatchesInvite(w, r)
    }).Methods("POST")

    protectedMatchRouter.HandleFunc("/accept", func(w http.ResponseWriter, r *http.Request) {
        // handlers.MatchesAccept(w, r)
    }).Methods("POST")

    // Get match details
    publicMatchRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
        // handlers.MatchesGet(w, r)
    }).Methods("GET")

    // Submit code for a match
    protectedMatchRouter.HandleFunc("/{id}/submission", func(w http.ResponseWriter, r *http.Request) {
        // handlers.MatchesSubmit(w, r)
    }).Methods("POST")


    // --------------------
    // Problems Routes
    // --------------------
    // Problem routes can be public (no authentication needed)
    problemRouter := r.PathPrefix("/problems").Subrouter()
    problemRouter.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
        // handlers.ProblemsRandom(w, r)
    }).Methods("GET")


    // --------------------
    // WebSocket Upgrader
    // --------------------
    // WebSocket endpoint requires authentication
    wsRouter := r.PathPrefix("/ws").Subrouter()
    wsRouter.Use(authMiddleware)

    wsRouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
        // handlers.WSConnect(w, r)
    }).Methods("GET")

    return r
}
