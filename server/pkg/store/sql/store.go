package sql

import (
	"context"
	"database/sql"
	"fmt"
	"leetcodeduels/pkg/models"

	_ "github.com/microsoft/go-mssqldb"
)

type Store struct {
	db *sql.DB
}

// NewStore creates a new Store with a database connection.
func NewStore(databaseURL string) (*Store, error) {

	server := "lcduels-db.database.windows.net"
	database := "lc_duels"
	port := 1433
	user := "CloudSAb22a1f85"
	password := "gF7eV^c)aP]_L3M"

	// Build connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		server, user, password, port, database)
	var err error
	// Create connection pool
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %v", err)
	}
	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}
	fmt.Printf("Connected!")

	return &Store{db: db}, nil
}

// GetAllProblems retrieves all problems from the database.
func (s *Store) GetAllProblems() ([]models.Problem, error) {
	const query = `
		SELECT problem_id, name, url, difficulty
		FROM problems
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []models.Problem
	for rows.Next() {
		var p models.Problem
		if err := rows.Scan(&p.ProblemID, &p.Name, &p.URL, &p.Difficulty); err != nil {
			return nil, err
		}
		problems = append(problems, p)
	}
	return problems, nil
}

// Retrieve Random Problem from DB
func (s *Store) GetRandomProblem() (*models.Problem, error) {
	var p models.Problem
	const query = `
		SELECT problem_id, name, url, difficulty
		FROM problems
		ORDER BY RANDOM()
		LIMIT 1
	`

	err := s.db.QueryRow(query).Scan(&p.ProblemID, &p.Name, &p.URL, &p.Difficulty)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Retrieve All Problems with specific tag from DB
func (s *Store) GetProblemsByTag(tagID int) ([]models.Problem, error) {
	const query = `
		SELECT p.problem_id, p.name, p.url, p.difficulty
		FROM problems p
		INNER JOIN problem_tags pt
		ON p.problem_id = pt.problem_id
		WHERE pt.tag_id = $1
	`

	rows, err := s.db.Query(query, tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []models.Problem
	for rows.Next() {
		var p models.Problem
		if err := rows.Scan(&p.ProblemID, &p.Name, &p.URL, &p.Difficulty); err != nil {
			return nil, err
		}
		problems = append(problems, p)
	}
	return problems, nil
}

// Retrieve Random Problem with specific tag from DB
func (s *Store) GetRandomProblemByTag(tagID int) (*models.Problem, error) {
	var p models.Problem
	const query = `
		SELECT p.problem_id, p.name, p.url, p.difficulty
		FROM problems p
		INNER JOIN problem_tags pt
		ON p.problem_id = pt.problem_id
		WHERE pt.tag_id = $1
		ORDER BY RANDOM()
		LIMIT 1
	`
	err := s.db.QueryRow(query, tagID).Scan(&p.ProblemID, &p.Name, &p.URL, &p.Difficulty)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetAllTags retrieves all unique tags from the database.
func (s *Store) GetAllTags() ([]string, error) {
	const query = `
		SELECT DISTINCT tag_name
		FROM tags
	`

	rows, err := s.db.Query(query)
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

// GetTagsByProblem retrieves all tags associated with a specific problem.
func (s *Store) GetTagsByProblem(problemID int) ([]string, error) {
	const query = `
        SELECT t.tag_name
        FROM tags t
        INNER JOIN problem_tags pt ON t.tag_id = pt.tag_id
        WHERE pt.problem_id = $1
    `

	rows, err := s.db.Query(query, problemID)
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
