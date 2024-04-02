package db

import (
	"database/sql"
	"fmt"
	"leetcodeduels/pkg/config"
	"leetcodeduels/pkg/models"

	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

// NewStore creates a new Store with a database connection.
func NewStore(db_url string) (*Store, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	connStr := cfg.DB_URL
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	//defer db.Close()

	var version string
	if err = db.QueryRow("select version()").Scan(&version); err != nil {
		panic(err)
	}

	fmt.Printf("version=%s\n", version)
	return &Store{db: db}, nil
}

// GetAllProblems retrieves all problems from the database.
func (s *Store) GetAllProblems() ([]models.Problem, error) {
	rows, err := s.db.Query("SELECT id, num, name, slug, difficulty
							FROM problems")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []models.Problem
	for rows.Next() {
		var p models.Problem
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
			return nil, err
		}
		problems = append(problems, p)
	}
	return problems, nil
}

// Retrieve Random Problem from DB
func (s *Store) GetRandomProblem() (*models.Problem, error) {
	var p models.Problem
	err := s.db.QueryRow("SELECT problem_num, name, slug, difficulty 
						  FROM problems 
						  ORDER BY RANDOM() 
						  LIMIT 1").Scan(&p.ProblemID, &p.Name, &p.URL, &p.Difficulty)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Retrieve All Problems with specific tag from DB
func (s *Store) GetProblemsByTag(tagID int) ([]models.Problem, error) {
	rows, err := s.db.Query("SELECT p.problem_id, p.name, p.url, p.difficulty
							 FROM problems p
							 INNER JOIN problem_tags pt
							 ON p.problem_id = pt.problem_id
							 WHERE pt.tag_id = $1", tagID)
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
