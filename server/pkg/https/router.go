package https

import (
	"leetcodeduels/pkg/store"

	"github.com/gorilla/mux"
)

// NewRouter initializes and returns a new router with configured routes.
func NewRouter(store store.Store) *mux.Router {
	handler := NewHandler(store)

	r := mux.NewRouter()

	// Register routes and handlers
	r.HandleFunc("/problems", handler.GetAllProblems).Methods("GET")
	r.HandleFunc("/random-problem", handler.GetRandomProblem).Methods("GET")
	r.HandleFunc("/problems-by-tag", handler.GetProblemsByTag).Methods("GET", "POST")
	r.HandleFunc("/random-problem-by-tag", handler.GetRandomProblemByTag).Methods("GET", "POST")
	r.HandleFunc("/tags", handler.GetAllTags).Methods("GET")
	r.HandleFunc("/tags-of-problem", handler.GetTagsByProblem).Methods("GET", "POST")

	return r
}
