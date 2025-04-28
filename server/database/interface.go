package database

import "leetcodeduels/models"

type I_DB interface {
	SaveOAuthUser(githubID int64, username string, accessToken string) error
	UpdateGithubAccessToken(githubID int64, newToken string) error

	GetUserProfile(githubID int64) (*models.User, error)
	GetUserRating(userID int64, newRating int) error
	UpdateUserRating(userID int64, newRating int) error

	GetAllProblems() ([]models.Problem, error) // For Debugging Only
	GetRandomProblem() (*models.Problem, error)
	GetProblemsByTag(tagID int) ([]models.Problem, error)
	GetRandomProblemByTag(tagID int) (*models.Problem, error)
	GetAllTags() ([]string, error)
	GetTagsByProblem(problemID int) ([]string, error)
	GetRandomProblemByDifficulty(difficulty models.Difficulty) (*models.Problem, error)
	GetRandomProblemByDifficultyAndTag(tagID int, difficulty models.Difficulty) (*models.Problem, error)
	GetRandomProblemForDuel(player1Tags, player2Tags []int, difficulties []models.Difficulty) (*models.Problem, error)
}
