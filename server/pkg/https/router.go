package https

import (
	"leetcodeduels/pkg/store/db"

	"github.com/gorilla/mux"
)

// NewRouter initializes and returns a new router with configured routes.
func NewRouter(store *db.Store) *mux.Router {
	handler := NewHandler(store)
	r := mux.NewRouter()

	r.HandleFunc("/tags", handler.GetAllTags).Methods("GET")
	r.HandleFunc("/login", handler.AuthenticateUser).Methods("GET", "POST")
	r.HandleFunc("/sign-up", handler.CreateUser).Methods("POST")
	r.HandleFunc("/check-user-ingame", handler.IsUserInGame).Methods("GET", "POST")

	/* Routes to be added
	// Game Session Handling
	r.HandleFunc("/matchmake", handler.AddPlayerToPool).Methods("PUT")
	r.HandleFunc("/cancel-matchmake", handler.RemovePlayerFromPool).Methods("PUT")
	r.HandleFunc("/matchmake", handler.GetMatch).Methods("POST") Post request sent that gives the submssion info
	*/

	// Testing Funcs
	r.HandleFunc("/random-problem", handler.GetRandomProblem).Methods("GET")
	r.HandleFunc("/random-problem-by-tag", handler.GetRandomProblemByTag).Methods("GET", "POST")
	r.HandleFunc("/problems", handler.GetAllProblems).Methods("GET")
	r.HandleFunc("/problems-by-tag", handler.GetProblemsByTag).Methods("GET", "POST")
	r.HandleFunc("/tags-of-problem", handler.GetTagsByProblem).Methods("GET", "POST")

	return r
}
