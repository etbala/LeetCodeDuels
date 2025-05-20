package store

import (
	"database/sql"
	"fmt"
	"leetcodeduels/models"
	"strings"
)

var DataStore *dataStore

type dataStore struct {
	db *sql.DB
}

func InitDataStore(connStr string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	DataStore = &dataStore{db: db}
	return nil
}

// SaveOAuthUser inserts or updates a GitHub OAuth user.
func (ds *dataStore) SaveOAuthUser(githubID int64, username string, accessToken string) error {
	query := `INSERT INTO github_oauth_users (github_id, username, access_token, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			ON CONFLICT (github_id)
			DO UPDATE SET access_token = $3, updated_at = NOW()`
	if _, err := ds.db.Exec(query, githubID, username, accessToken); err != nil {
		return fmt.Errorf("SaveOAuthUser: %w", err)
	}
	return nil
}

// UpdateGithubAccessToken updates only the access token for a user.
func (ds *dataStore) UpdateGithubAccessToken(githubID int64, newToken string) error {
	query := `UPDATE github_oauth_users SET access_token = $1, updated_at = NOW() WHERE github_id = $2`
	if _, err := ds.db.Exec(query, newToken, githubID); err != nil {
		return fmt.Errorf("UpdateGithubAccessToken: %w", err)
	}
	return nil
}

// GetUserProfile retrieves the full user record, or nil if not found.
func (ds *dataStore) GetUserProfile(githubID int64) (*models.User, error) {
	query := `SELECT github_id, username, access_token, created_at, updated_at, rating
			FROM github_oauth_users WHERE github_id = $1`
	row := ds.db.QueryRow(query, githubID)
	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.AccessToken, &u.CreatedAt, &u.UpdatedAt, &u.Rating); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetUserProfile: %w", err)
	}
	return &u, nil
}

// GetUserRating fetches a user's current rating.
func (ds *dataStore) GetUserRating(userID int64) (int, error) {
	var rating int
	query := `SELECT rating FROM github_oauth_users WHERE github_id = $1`
	if err := ds.db.QueryRow(query, userID).Scan(&rating); err != nil {
		return 0, fmt.Errorf("GetUserRating: %w", err)
	}
	return rating, nil
}

// UpdateUserRating sets a user's rating to newRating.
func (ds *dataStore) UpdateUserRating(userID int64, newRating int) error {
	query := `UPDATE github_oauth_users SET rating = $1 WHERE github_id = $2`
	if _, err := ds.db.Exec(query, newRating, userID); err != nil {
		return fmt.Errorf("UpdateUserRating: %w", err)
	}
	return nil
}

// GetAllProblems returns every problem in the table.
func (ds *dataStore) GetAllProblems() ([]models.Problem, error) {
	query := `SELECT id, name, slug, difficulty FROM problems`
	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("GetAllProblems: %w", err)
	}
	defer rows.Close()

	var out []models.Problem
	for rows.Next() {
		var p models.Problem
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
			return nil, fmt.Errorf("GetAllProblems scan: %w", err)
		}
		out = append(out, p)
	}
	return out, nil
}

// GetRandomProblem picks one random problem.
func (ds *dataStore) GetRandomProblem() (*models.Problem, error) {
	query := `SELECT id, name, slug, difficulty FROM problems ORDER BY RANDOM() LIMIT 1`
	var p models.Problem
	if err := ds.db.QueryRow(query).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblem: %w", err)
	}
	return &p, nil
}

// GetProblemsByTag returns all problems having the given tag ID.
func (ds *dataStore) GetProblemsByTag(tagID int) ([]models.Problem, error) {
	query := `
	SELECT p.id, p.name, p.slug, p.difficulty
	FROM problems p
	JOIN problem_tags pt ON p.id = pt.problem_id
	WHERE pt.tag_id = $1`
	rows, err := ds.db.Query(query, tagID)
	if err != nil {
		return nil, fmt.Errorf("GetProblemsByTag: %w", err)
	}
	defer rows.Close()

	var out []models.Problem
	for rows.Next() {
		var p models.Problem
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
			return nil, fmt.Errorf("GetProblemsByTag scan: %w", err)
		}
		out = append(out, p)
	}
	return out, nil
}

// GetRandomProblemByTag returns a single random problem for a tag.
func (ds *dataStore) GetRandomProblemByTag(tagID int) (*models.Problem, error) {
	query := `
	SELECT p.id, p.name, p.slug, p.difficulty
	FROM problems p
	JOIN problem_tags pt ON p.id = pt.problem_id
	WHERE pt.tag_id = $1
	ORDER BY RANDOM() LIMIT 1`
	var p models.Problem
	if err := ds.db.QueryRow(query, tagID).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblemByTag: %w", err)
	}
	return &p, nil
}

// GetAllTags returns every tag name.
func (ds *dataStore) GetAllTags() ([]string, error) {
	query := `SELECT name FROM tags`
	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("GetAllTags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("GetAllTags scan: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// GetTagsByProblem returns names of tags for a given problem ID.
func (ds *dataStore) GetTagsByProblem(problemID int) ([]string, error) {
	query := `
	SELECT t.name
	FROM tags t
	JOIN problem_tags pt ON t.id = pt.tag_id
	WHERE pt.problem_id = $1`
	rows, err := ds.db.Query(query, problemID)
	if err != nil {
		return nil, fmt.Errorf("GetTagsByProblem: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("GetTagsByProblem scan: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// GetRandomProblemByDifficulty picks a random problem with the given difficulty.
func (ds *dataStore) GetRandomProblemByDifficulty(difficulty models.Difficulty) (*models.Problem, error) {
	query := `SELECT id, name, slug, difficulty FROM problems WHERE difficulty = $1 ORDER BY RANDOM() LIMIT 1`
	var p models.Problem
	if err := ds.db.QueryRow(query, difficulty).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblemByDifficulty: %w", err)
	}
	return &p, nil
}

// GetRandomProblemByDifficultyAndTag picks a random problem filtered by difficulty and tag.
func (ds *dataStore) GetRandomProblemByDifficultyAndTag(tagID int, difficulty models.Difficulty) (*models.Problem, error) {
	query := `
	SELECT p.id, p.name, p.slug, p.difficulty
	FROM problems p
	JOIN problem_tags pt ON p.id = pt.problem_id
	WHERE pt.tag_id = $1 AND p.difficulty = $2
	ORDER BY RANDOM() LIMIT 1`
	var p models.Problem
	if err := ds.db.QueryRow(query, tagID, difficulty).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblemByDifficultyAndTag: %w", err)
	}
	return &p, nil
}

// GetRandomProblemForDuel picks a random problem matching preferences of both players and a set of difficulties.
func (ds *dataStore) GetRandomProblemForDuel(player1Tags, player2Tags []int, difficulties []models.Difficulty) (*models.Problem, error) {
	// Build IN clauses dynamically
	if len(difficulties) == 0 {
		return nil, fmt.Errorf("GetRandomProblemForDuel: no difficulties provided")
	}

	// placeholders
	tagPlaceholders := func(ids []int, start int) []string {
		ps := make([]string, len(ids))
		for i := range ids {
			ps[i] = fmt.Sprintf("$%d", start+i)
		}
		return ps
	}

	p1ph := tagPlaceholders(player1Tags, 1)
	p2ph := tagPlaceholders(player2Tags, len(player1Tags)+1)
	dp := tagPlaceholders(func() []int { return nil }(), len(player1Tags)+len(player2Tags)+1)
	// difficulties start index = len(p1)+len(p2)+1
	for i := range difficulties {
		dp = append(dp, fmt.Sprintf("$%d", len(player1Tags)+len(player2Tags)+1+i))
	}

	query := fmt.Sprintf(`
	SELECT p.id, p.name, p.slug, p.difficulty
	FROM problems p
	WHERE p.difficulty IN (%s)
	AND EXISTS (SELECT 1 FROM problem_tags pt1 WHERE pt1.problem_id = p.id AND pt1.tag_id IN (%s))
	AND EXISTS (SELECT 1 FROM problem_tags pt2 WHERE pt2.problem_id = p.id AND pt2.tag_id IN (%s))
	ORDER BY RANDOM() LIMIT 1`,
		strings.Join(dp, ","), strings.Join(p1ph, ","), strings.Join(p2ph, ","))

	// build args
	var args []interface{}
	for _, id := range append(player1Tags, player2Tags...) {
		args = append(args, id)
	}
	for _, d := range difficulties {
		args = append(args, d)
	}

	var p models.Problem
	if err := ds.db.QueryRow(query, args...).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblemForDuel: %w", err)
	}
	return &p, nil
}
