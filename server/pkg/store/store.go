package store

import "leetcodeduels/pkg/models"

// Store is an interface for database operations.
type Store interface {
	GetAllProblems() ([]models.Problem, error)
	GetRandomProblem() (*models.Problem, error)
	GetProblemsByTag(tagID int) ([]models.Problem, error)
	GetRandomProblemByTag(tagID int) (*models.Problem, error)
	GetAllTags() ([]string, error)
	GetTagsByProblem(problemID int) ([]string, error)
    CreateUser(w http.ResponseWriter, username, password, email string) (bool, error)
	AuthenticateUser(username, password string) (bool, error)
}
