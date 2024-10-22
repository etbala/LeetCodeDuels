package https

import (
	"github.com/gorilla/mux"
)

// NewRouter initializes and returns a new router with configured routes.
func NewRouter() *mux.Router {
	r := mux.NewRouter()

	// OAuth Login
	r.HandleFunc("/oauth/authorize", OAuthAuthorize).Methods("GET", "POST")
	r.HandleFunc("/oauth/callback", OAuthCallback).Methods("GET", "POST")

	r.HandleFunc("/tags", GetAllTags).Methods("GET")

	r.HandleFunc("/user/check-ingame", IsUserInGame).Methods("GET", "POST")

	r.HandleFunc("/game/submission", AddSubmission).Methods("GET", "POST")

	/* Routes to be added
	// Game Session Handling
	r.HandleFunc("/matchmake", AddPlayerToPool).Methods("PUT")
	r.HandleFunc("/cancel-matchmake", RemovePlayerFromPool).Methods("PUT")
	r.HandleFunc("/user-profile", GetUserInfo).Methods("GET", "POST)")
	*/

	// Testing Funcs
	r.HandleFunc("/random-problem", GetRandomProblem).Methods("GET")
	r.HandleFunc("/random-problem-by-tag", GetRandomProblemByTag).Methods("GET", "POST")
	r.HandleFunc("/problems", GetAllProblems).Methods("GET")
	r.HandleFunc("/problems-by-tag", GetProblemsByTag).Methods("GET", "POST")
	r.HandleFunc("/tags-of-problem", GetTagsByProblem).Methods("GET", "POST")

	return r
}
