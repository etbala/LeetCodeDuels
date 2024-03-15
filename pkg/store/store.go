package store

/*
Defines an interface for database operations and a PostgreSQL implementation of
this interface. This allows for easy swapping of the database backend if needed
without changing the business logic.
*/

import "leetcodeduels/pkg/models"

// Store is an interface for datastore operations.
type Store interface {
	GetAllProblems() ([]models.Problem, error)
	GetRandomProblem() (*models.Problem, error)
	GetProblemsByTag(tagID int) ([]models.Problem, error)
	GetRandomProblemByTag(tagID int) (*models.Problem, error)
	GetAllTags() ([]string, error)
	GetTagsByProblem(problemID int) ([]string, error)
}
