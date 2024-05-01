package store

import (
	"database/sql"
	"fmt"
	"leetcodeduels/pkg/config"
	"leetcodeduels/pkg/models"
	"net/http"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DataStore *dataStore

type dataStore struct {
	db *sql.DB
}

func init() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	connStr := cfg.DB_URL
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	DataStore = &dataStore{db: db}
}

func (ds *dataStore) GetAllProblems() ([]models.Problem, error) {
	rows, err := ds.db.Query(`SELECT frontend_id, name, slug, difficulty 
							  FROM problems`)
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

func (ds *dataStore) GetRandomProblem() (*models.Problem, error) {
	var p models.Problem
	err := ds.db.QueryRow(`SELECT frontend_id, name, slug, difficulty
						FROM problems
						ORDER BY RANDOM()
						LIMIT 1`).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (ds *dataStore) GetProblemsByTag(tagID int) ([]models.Problem, error) {
	rows, err := ds.db.Query(`SELECT p.frontend_id, p.name, p.slug, p.difficulty
							FROM problems p
							INNER JOIN problem_tags pt
							ON p.frontend_id = pt.problem_num
							WHERE pt.tag_num = $1`, tagID)

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

func (ds *dataStore) GetRandomProblemByTag(tagID int) (*models.Problem, error) {
	var p models.Problem
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

func (ds *dataStore) GetRandomProblemByDifficulty(difficulty string) (*models.Problem, error) {
	if difficulty != "Easy" && difficulty != "Medium" && difficulty != "Hard" {
		return nil, fmt.Errorf("invalid difficulty")
	}

	var p models.Problem
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

func (ds *dataStore) GetRandomProblemByDifficultyAndTag(tagID int, difficulty string) (*models.Problem, error) {
	if difficulty != "Easy" && difficulty != "Medium" && difficulty != "Hard" {
		return nil, fmt.Errorf("invalid difficulty")
	}

	var p models.Problem
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

func (ds *dataStore) CreateUser(w http.ResponseWriter, username, password, email string) (bool, error) {
	defaultRating := 1000
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	// Check if user is already in the database
	var exists int
	err = ds.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2", username, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	if exists > 0 {
		return false, fmt.Errorf("user with given username or email already exists")
	}

	// Insert the user into the database
	_, err = ds.db.Exec("INSERT INTO users (username, password_hash, email, rating) VALUES ($1, $2, $3, $4)",
		username, string(hashedPassword), email, defaultRating)
	if err != nil {
		return false, err
	}

	// Set a cookie to indicate successful sign-up
	http.SetCookie(w, &http.Cookie{
		Name:     "loggedIn",
		Value:    "true",
		Path:     "/",   // Cookie is valid for all paths
		HttpOnly: true,  // The cookie cannot be accessed by client-side APIs, like JavaScript.
		Secure:   true,  // Ensure this cookie is only sent over HTTPS.
		MaxAge:   86400, // Sets the cookie to expire after 1 day (86400 seconds).
	})

	return true, nil
}

func (ds *dataStore) AuthenticateUser(username, password string) (bool, error) {
	var hashedPassword string
	err := ds.db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&hashedPassword)
	if err != nil {
		return false, err // User not found or other database error
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, fmt.Errorf("invalid password") // Specific error for password mismatch
	}

	return true, nil
}

func (ds *dataStore) GetUserRating(UUID string, newRating int) error {
	_, err := ds.db.Exec(`SELECT rating FROM users WHERE uuid = $1`, newRating, UUID)
	if err != nil {
		return fmt.Errorf("error getting user rating: %w", err)
	}
	return nil
}

func (ds *dataStore) UpdateUserRating(UUID string, newRating int) error {
	query := `UPDATE users SET rating = $1 WHERE uuid = $2`
	_, err := ds.db.Exec(query, newRating, UUID)
	if err != nil {
		return fmt.Errorf("error updating user rating: %w", err)
	}
	return nil
}
