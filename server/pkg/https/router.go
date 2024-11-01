package https

import (
	"leetcodeduels/api/auth"

	"github.com/gorilla/mux"
)

// NewRouter initializes and returns a new router with configured routes.
func NewRouter() *mux.Router {
	publicRouter := mux.NewRouter()
	protectedRouter := publicRouter.PathPrefix("/").Subrouter()
	protectedRouter.Use(auth.JWTMiddleware)

	// Public Routes
	publicRouter.HandleFunc("/tags", GetAllTags).Methods("GET")
	publicRouter.HandleFunc("/oauth/exchange-token", OAuthExchangeToken).Methods("POST")
	publicRouter.HandleFunc("/user/check-ingame", IsUserInGame).Methods("GET")
	publicRouter.HandleFunc("/user/profile", GetUserProfile).Methods("GET")
	// publicRouter.HandleFunc("/user/match-history", GetUserMatchHistory).Methods("GET")

	// Websocket Route
	protectedRouter.HandleFunc("/ws", WsHandler).Methods("GET")

	// Protected Routes
	protectedRouter.HandleFunc("/game/submission", AddSubmission).Methods("POST")
	protectedRouter.HandleFunc("/matchmaking/enter", AddPlayerToPool).Methods("POST")
	protectedRouter.HandleFunc("/matchmaking/leave", RemovePlayerFromPool).Methods("POST")

	// Testing Funcs
	// r.HandleFunc("/random-problem", GetRandomProblem).Methods("GET")
	// r.HandleFunc("/random-problem-by-tag", GetRandomProblemByTag).Methods("GET", "POST")
	// r.HandleFunc("/problems", GetAllProblems).Methods("GET")
	// r.HandleFunc("/problems-by-tag", GetProblemsByTag).Methods("GET", "POST")
	// r.HandleFunc("/tags-of-problem", GetTagsByProblem).Methods("GET", "POST")

	return publicRouter
}
