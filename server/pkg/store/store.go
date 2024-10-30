package store

import (
	"database/sql"
	"fmt"
	"leetcodeduels/internal/enums"
	"strings"

	_ "github.com/lib/pq"
)

var DataStore *dataStore

type dataStore struct {
	db *sql.DB
}

func InitDataStore(connStr string) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	DataStore = &dataStore{db: db}
}

func (ds *dataStore) SaveOAuthUser(githubID int64, username string, accessToken string) error {
	_, err := ds.db.Exec(`INSERT INTO github_oauth_users (github_id, username, access_token, created_at, updated_at)
						  VALUES ($1, $2, $3, NOW(), NOW())
						  ON CONFLICT (github_id) 
						  DO UPDATE SET access_token = $3, updated_at = NOW();
						  `, githubID, username, accessToken)
	if err != nil {
		return fmt.Errorf("failed to save OAuth user: %v", err)
	}
	return nil
}

func (ds *dataStore) GetUserProfile(githubID int64) (*User, error) {
	row := ds.db.QueryRow(`SELECT github_id, username, access_token, created_at, updated_at, rating 
			  			   FROM github_oauth_users WHERE github_id = $1`, githubID)
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.AccessToken, &user.CreatedAt, &user.UpdatedAt, &user.Rating)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get OAuth user: %v", err)
	}
	return &user, nil
}

func (ds *dataStore) UpdateGithubAccessToken(githubID int64, newToken string) error {
	_, err := ds.db.Exec(`UPDATE github_oauth_users SET access_token = $1, updated_at = NOW() WHERE github_id = $2`, newToken, githubID)
	if err != nil {
		return fmt.Errorf("failed to update GitHub access token: %v", err)
	}
	return nil
}

func (ds *dataStore) GetAllProblems() ([]Problem, error) {
	rows, err := ds.db.Query(`SELECT frontend_id, name, slug, difficulty 
							  FROM problems`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []Problem
	for rows.Next() {
		var p Problem
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
			return nil, err
		}
		problems = append(problems, p)
	}
	return problems, nil
}

func (ds *dataStore) GetRandomProblem() (*Problem, error) {
	var p Problem
	err := ds.db.QueryRow(`SELECT frontend_id, name, slug, difficulty
						FROM problems
						ORDER BY RANDOM()
						LIMIT 1`).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (ds *dataStore) GetProblemsByTag(tagID int) ([]Problem, error) {
	rows, err := ds.db.Query(`SELECT p.frontend_id, p.name, p.slug, p.difficulty
							FROM problems p
							INNER JOIN problem_tags pt
							ON p.frontend_id = pt.problem_num
							WHERE pt.tag_num = $1`, tagID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []Problem
	for rows.Next() {
		var p Problem
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
			return nil, err
		}
		problems = append(problems, p)
	}
	return problems, nil
}

func (ds *dataStore) GetRandomProblemByTag(tagID int) (*Problem, error) {
	var p Problem
	err := ds.db.QueryRow(`SELECT p.frontend_id, p.name, p.slug, p.difficulty
						FROM problems p 
						INNER JOIN problem_tags pt 
						ON p.frontend_id = pt.problem_id 
						WHERE pt.tag_id = $1 
						ORDER BY RANDOM() 
						LIMIT 1`, tagID).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (ds *dataStore) GetAllTags() ([]string, error) {
	rows, err := ds.db.Query(`SELECT DISTINCT name 
							 FROM tags`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (ds *dataStore) GetTagsByProblem(problemID int) ([]string, error) {
	rows, err := ds.db.Query(`SELECT t.name
							FROM tags t
							INNER JOIN problem_tags pt ON t.id = pt.tag_id
							WHERE pt.problem_id = $1`, problemID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (ds *dataStore) GetRandomProblemByDifficulty(difficulty enums.Difficulty) (*Problem, error) {
	var p Problem
	err := ds.db.QueryRow(`SELECT frontend_id, name, slug, difficulty
						FROM problems 
						WHERE difficulty = $1
						ORDER BY RANDOM() 
						LIMIT 1`, difficulty).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (ds *dataStore) GetRandomProblemByDifficultyAndTag(tagID int, difficulty enums.Difficulty) (*Problem, error) {
	var p Problem
	err := ds.db.QueryRow(`SELECT p.frontend_id, p.name, p.slug, p.difficulty
						FROM problems p 
						WHERE difficulty = $1
						INNER JOIN problem_tags pt 
						ON p.frontend_id = pt.problem_id 
						WHERE pt.tag_id = $2
						ORDER BY RANDOM() 
						LIMIT 1`, difficulty, tagID).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Gets a random problem that matches a tag preference from both Player 1 and Player 2, and has one of the provided difficulties
func (ds *dataStore) GetRandomProblemForDuel(player1Tags, player2Tags []int, difficulties []enums.Difficulty) (*Problem, error) {
	if len(difficulties) < 1 || len(difficulties) > 3 {
		// Must provide at least one difficulty
		return nil, fmt.Errorf("error: invalid number of difficulties provided")
	}

	// TODO / FIX: Currently uses ID of tags instead of String, WILL CURRENTLY BREAK OR NOT WORK

	// Generate placeholders for SQL IN clause
	placeholders := func(length int, startIdx int) []string {
		result := make([]string, length)
		for i := range result {
			result[i] = fmt.Sprintf("$%d", i+startIdx)
		}
		return result
	}
	player1TagsPlaceholders := placeholders(len(player1Tags), 1)
	player2TagsPlaceholders := placeholders(len(player2Tags), len(player1Tags)+1)
	difficultiesPlaceholders := placeholders(len(difficulties), len(player1Tags)+len(player2Tags)+1)

	query := fmt.Sprintf(`
        SELECT p.frontend_id, p.name, p.slug, p.difficulty
        FROM problems p
        WHERE p.difficulty IN (%s)
        AND EXISTS (
            SELECT 1
            FROM problem_tags pt1
            WHERE pt1.problem_id = p.frontend_id
            AND pt1.tag_id IN (%s)
        )
        AND EXISTS (
            SELECT 1
            FROM problem_tags pt2
            WHERE pt2.problem_id = p.frontend_id
            AND pt2.tag_id IN (%s)
        )
        ORDER BY RANDOM()
        LIMIT 1`,
		strings.Join(difficultiesPlaceholders, ","),
		strings.Join(player1TagsPlaceholders, ","),
		strings.Join(player2TagsPlaceholders, ","))

	// Prepare the arguments for the query
	args := make([]interface{}, 0, len(player1Tags)+len(player2Tags)+len(difficulties))
	for _, tag := range player1Tags {
		args = append(args, tag)
	}
	for _, tag := range player2Tags {
		args = append(args, tag)
	}
	for _, difficulty := range difficulties {
		args = append(args, difficulty)
	}

	// Execute the query
	var p Problem
	err := ds.db.QueryRow(query, args...).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (ds *dataStore) GetUserRating(userID int64, newRating int) error {
	_, err := ds.db.Exec(`SELECT rating FROM github_oauth_users WHERE github_id = $1`, newRating, userID)
	if err != nil {
		return fmt.Errorf("error getting user rating: %w", err)
	}
	return nil
}

func (ds *dataStore) UpdateUserRating(userID int64, newRating int) error {
	query := `UPDATE github_oauth_users SET rating = $1 WHERE github_id = $2`
	_, err := ds.db.Exec(query, newRating, userID)
	if err != nil {
		return fmt.Errorf("error updating user rating: %w", err)
	}
	return nil
}
