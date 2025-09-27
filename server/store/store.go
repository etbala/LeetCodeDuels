package store

import (
	"database/sql"
	"fmt"
	"leetcodeduels/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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

// Insert or update an OAuth user based on their GitHub ID.
func (ds *dataStore) SaveOAuthUser(githubID int64, accessToken string, username string, discriminator string, avatar_url string) error {
	query := `INSERT INTO users (id, access_token, username, discriminator, avatar_url, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			ON CONFLICT (id)
			DO UPDATE SET access_token = $2, updated_at = NOW()`
	if _, err := ds.db.Exec(query, githubID, accessToken, username, discriminator, avatar_url); err != nil {
		return fmt.Errorf("SaveOAuthUser: %w", err)
	}
	return nil
}

// Update a user's GitHub access token.
func (ds *dataStore) UpdateGithubAccessToken(githubID int64, newToken string) error {
	query := `UPDATE users SET access_token = $1, updated_at = NOW() WHERE id = $2`
	if _, err := ds.db.Exec(query, newToken, githubID); err != nil {
		return fmt.Errorf("UpdateGithubAccessToken: %w", err)
	}
	return nil
}

// Return the full user record by GitHub ID, or nil if not found.
func (ds *dataStore) GetUserProfile(githubID int64) (*models.User, error) {
	query := `SELECT id, access_token, 	username, discriminator, 
			lc_username, avatar_url, created_at, updated_at, rating
			FROM users WHERE id = $1`
	row := ds.db.QueryRow(query, githubID)
	var u models.User
	err := row.Scan(&u.ID, &u.AccessToken, &u.Username, &u.Discriminator,
		&u.LeetCodeUsername, &u.AvatarURL, &u.CreatedAt,
		&u.UpdatedAt, &u.Rating)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetUserProfile: %w", err)
	}
	return &u, nil
}

// Retrieves the full user record by username + discriminator, or nil if not found.
func (ds *dataStore) GetUserProfileByUsername(username string, discriminator string) (*models.User, error) {
	query := `SELECT id, access_token, 	username, discriminator,
			lc_username, avatar_url, created_at, updated_at, rating
			FROM users WHERE username = $1 AND discriminator = $2`
	row := ds.db.QueryRow(query, username, discriminator)
	var u models.User
	err := row.Scan(&u.ID, &u.AccessToken, &u.Username, &u.Discriminator,
		&u.LeetCodeUsername, &u.AvatarURL, &u.CreatedAt,
		&u.UpdatedAt, &u.Rating)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetUserProfile by username: %w", err)
	}
	return &u, nil
}

// Returns a list of users whose usernames contain the given substring, limited to 'limit' results.
func (ds *dataStore) SearchUsersByUsername(username string, limit int) ([]models.UserInfoResponse, error) {
	query := `SELECT id, username, discriminator, lc_username, avatar_url, rating
			FROM users WHERE username ILIKE $1 LIMIT $2`

	// Only match usernames that start with given substring
	rows, err := ds.db.Query(query, username+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("SearchUsersByUsername: %w", err)
	}
	defer rows.Close()

	var users []models.UserInfoResponse
	for rows.Next() {
		var u models.UserInfoResponse
		if err := rows.Scan(&u.ID, &u.Username, &u.Discriminator,
			&u.LCUsername, &u.AvatarURL, &u.Rating); err != nil {
			return nil, fmt.Errorf("SearchUsersByUsername: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

// GetUserRating fetches a user's current rating.
func (ds *dataStore) GetUserRating(userID int64) (int, error) {
	var rating int
	query := `SELECT rating FROM users WHERE id = $1`
	err := ds.db.QueryRow(query, userID).Scan(&rating)
	if err != nil {
		return 0, fmt.Errorf("GetUserRating: %w", err)
	}
	return rating, nil
}

// Checks if a given username + discriminator combo exists.
func (ds *dataStore) DiscriminatorExists(username string, discriminator string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND discriminator = $2)`
	err := ds.db.QueryRow(query, username, discriminator).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("DiscriminatorExists: %w", err)
	}
	return exists, nil
}

// Updates a user's provided profile fields. Only non-empty fields are updated.
func (ds *dataStore) UpdateUser(userID int64, username string, discriminator string, lc_username string) error {
	var queryParts []string
	var args []interface{}
	argId := 1

	if username != "" && discriminator != "" {
		queryParts = append(queryParts, fmt.Sprintf("username = $%d", argId))
		args = append(args, username)
		argId++

		queryParts = append(queryParts, fmt.Sprintf("discriminator = $%d", argId))
		args = append(args, discriminator)
		argId++
	}

	if lc_username != "" {
		queryParts = append(queryParts, fmt.Sprintf("lc_username = $%d", argId))
		args = append(args, lc_username)
		argId++
	}

	if len(queryParts) == 0 {
		return fmt.Errorf("UpdateUser: no fields to update")
	}

	queryParts = append(queryParts, "updated_at = NOW()")
	args = append(args, userID)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d",
		strings.Join(queryParts, ", "),
		argId)

	_, err := ds.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}

	return nil
}

// Sets a user's username and discriminator to new values.
func (ds *dataStore) UpdateUsernameDiscriminator(userID int64, username string, discriminator string) error {
	query := `UPDATE users SET username = $1, discriminator = $2 WHERE id = $3`
	_, err := ds.db.Exec(query, username, discriminator, userID)
	if err != nil {
		return fmt.Errorf("UpdateUsernameDiscriminator: %w", err)
	}
	return nil
}

// Sets a user's linked LeetCode username to a new value.
func (ds *dataStore) UpdateLCUsername(userID int64, newLCUsername string) error {
	query := `UPDATE users SET lc_username = $1 WHERE id = $2`
	_, err := ds.db.Exec(query, newLCUsername, userID)
	if err != nil {
		return fmt.Errorf("UpdateLCUsername: %w", err)
	}
	return nil
}

// Sets a user's rating to a new value.
func (ds *dataStore) UpdateUserRating(userID int64, newRating int) error {
	query := `UPDATE users SET rating = $1 WHERE id = $2`
	_, err := ds.db.Exec(query, newRating, userID)
	if err != nil {
		return fmt.Errorf("UpdateUserRating: %w", err)
	}
	return nil
}

// Deletes a user and all associated data (matches, submissions, etc).
// TODO: Remove? We don't want to delete historical match data, what data do we have to delete?
func (ds *dataStore) DeleteUser(userID int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := ds.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("DeleteUser: %w", err)
	}
	return nil
}

// Returns a list of users whose usernames start with the given prefix, limited to 'limit' results.
func (ds *dataStore) GetMatchingUsers(username string, limit int) ([]models.User, error) {
	query := `SELECT id, username, discriminator, lc_username, avatar_url, rating, created_at
			FROM users WHERE username ILIKE $1 ORDER BY rating DESC LIMIT $2`
	rows, err := ds.db.Query(query, username+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("GetMatchingUsers: %w", err)
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Discriminator,
			&u.LeetCodeUsername, &u.AvatarURL, &u.Rating, &u.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetMatchingUsers scan: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

// Return all problems in database
func (ds *dataStore) GetAllProblems() ([]models.Problem, error) {
	query := `SELECT id, name, slug, difficulty FROM problems WHERE is_paid = false`
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

// Return all problems that have a given tag.
func (ds *dataStore) GetProblemsByTag(tagID int) ([]models.Problem, error) {
	query := `
	SELECT p.id, p.name, p.slug, p.difficulty
	FROM problems p
	JOIN problem_tags pt ON p.id = pt.problem_id
	WHERE pt.tag_id = $1
		AND p.is_paid = false`
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

// Returns all tags in the database.
func (ds *dataStore) GetAllTags() ([]models.Tag, error) {
	query := `SELECT id, name FROM tags`
	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("GetAllTags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("GetAllTags scan: %w", err)
		}
		t := models.Tag{ID: id, Name: name}
		tags = append(tags, t)
	}
	return tags, nil
}

// Returns names of all tags of a given problem.
func (ds *dataStore) GetTagsByProblem(problemID int) ([]models.Tag, error) {
	query := `
	SELECT t.name, t.id
	FROM tags t
	JOIN problem_tags pt ON t.id = pt.tag_id
	WHERE pt.problem_id = $1`
	rows, err := ds.db.Query(query, problemID)
	if err != nil {
		return nil, fmt.Errorf("GetTagsByProblem: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.Name, &t.ID); err != nil {
			return nil, fmt.Errorf("GetTagsByProblem scan: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// Returns a single random problem from the database.
func (ds *dataStore) GetRandomProblem() (*models.Problem, error) {
	query := `SELECT id, name, slug, difficulty FROM problems WHERE is_paid = false ORDER BY RANDOM() LIMIT 1`
	var p models.Problem
	if err := ds.db.QueryRow(query).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblem: %w", err)
	}
	return &p, nil
}

// Returns a single random problem that has the specified tag.
func (ds *dataStore) GetRandomProblemByTag(tagID int) (*models.Problem, error) {
	query := `
	SELECT p.id, p.name, p.slug, p.difficulty
	FROM problems p
	JOIN problem_tags pt ON p.id = pt.problem_id
	WHERE pt.tag_id = $1
		AND p.is_paid = false
	ORDER BY RANDOM() LIMIT 1`
	var p models.Problem
	if err := ds.db.QueryRow(query, tagID).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblemByTag: %w", err)
	}
	return &p, nil
}

// Returns a single random problem that has ANY of the specified tags.
func (ds *dataStore) GetRandomProblemByTags(tagIDs []int) (*models.Problem, error) {
	if len(tagIDs) == 0 {
		return ds.GetRandomProblem()
	}
	query := `
	SELECT p.id, p.name, p.slug, p.difficulty
	FROM problems p
	JOIN problem_tags pt ON p.id = pt.problem_id
	WHERE pt.tag_id = ANY($1)
		AND p.is_paid = false
	ORDER BY RANDOM() LIMIT 1`
	var p models.Problem
	if err := ds.db.QueryRow(query, pq.Array(tagIDs)).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblemByTags: %w", err)
	}
	return &p, nil
}

// Returns a single random problem that matches any of the specified difficulties.
func (ds *dataStore) GetRandomProblemByDifficulties(difficulties []models.Difficulty) (*models.Problem, error) {
	if len(difficulties) == 0 {
		return ds.GetRandomProblem() // no filter if no difficulties provided
	}
	query := `
	SELECT p.id, p.name, p.slug, p.difficulty
	FROM problems p
	WHERE p.difficulty = ANY($1)
		AND p.is_paid = false
	ORDER BY RANDOM() LIMIT 1`
	var p models.Problem
	if err := ds.db.QueryRow(query, pq.Array(difficulties)).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblemByDifficulties: %w", err)
	}
	return &p, nil
}

// Returns a single random problem that matches any of the specified
// difficulties and has any of the specified tags.
func (ds *dataStore) GetRandomProblemByTagsAndDifficulties(
	tags []int, difficulties []models.Difficulty,
) (*models.Problem, error) {
	if len(difficulties) == 0 {
		return ds.GetRandomProblemByTags(tags)
	}

	if len(tags) == 0 {
		return ds.GetRandomProblemByDifficulties(difficulties)
	}

	args := []interface{}{pq.Array(difficulties)}

	tagSubquery := ``
	if len(tags) > 0 {
		args = append(args, pq.Array(tags))
		tagSubquery = `
		AND EXISTS (
		SELECT 1 
		FROM problem_tags pt 
		WHERE pt.problem_id = p.id 
			AND p.is_paid = false
			AND pt.tag_id = ANY($2)
		)`
	}

	query := fmt.Sprintf(`
	SELECT p.id, p.name, p.slug, p.difficulty
	FROM problems p
	WHERE p.is_paid = false
		AND p.difficulty = ANY($1)
	%s
	ORDER BY RANDOM() LIMIT 1`,
		tagSubquery)

	var p models.Problem
	if err := ds.db.QueryRow(query, args...).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblemDuel: %w", err)
	}
	return &p, nil
}

// Stores a new match record in the database.
func (ds *dataStore) StoreMatch(match *models.Session) error {
	query := `
	INSERT INTO matches (id, problem_id, is_rated, status, winner_id, start_time, end_time)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := ds.db.Exec(query, match.ID, match.Problem.ID, match.IsRated,
		match.Status, match.Winner, match.StartTime, match.EndTime)
	if err != nil {
		return fmt.Errorf("StoreMatch: %w", err)
	}

	return nil
}

// Returns full match information for a specified session, or nil if not found.
func (ds *dataStore) GetMatch(matchID uuid.UUID) (*models.Session, error) {
	// TODO: Investigate if triple query or single query with ARRAY_AGG is better
	const matchQ = `
	SELECT 
	  m.id, 
	  m.problem_id, 
	  p.name, 
	  p.slug, 
	  p.difficulty, 
	  m.is_rated, 
	  m.status, 
	  m.winner_id, 
	  m.start_time, 
	  m.end_time
	FROM matches m
	JOIN problems p ON p.id = m.problem_id
	WHERE m.id = $1`

	var (
		id        string
		probID    int
		probName  string
		probSlug  string
		probDiff  string
		isRated   bool
		statusStr string
		winnerID  int64
		startTime time.Time
		endTime   time.Time
	)
	err := ds.db.QueryRow(matchQ, matchID.String()).
		Scan(&id, &probID, &probName, &probSlug, &probDiff, &isRated, &statusStr, &winnerID, &startTime, &endTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetMatch: querying match: %w", err)
	}

	parsedStatus, err := models.ParseMatchStatus(statusStr)
	if err != nil {
		return nil, fmt.Errorf("GetMatch: parse status: %w", err)
	}
	parsedDiff, err := models.ParseDifficulty(probDiff)
	if err != nil {
		return nil, fmt.Errorf("GetMatch: parse difficulty: %w", err)
	}

	const playersQ = `
	SELECT player_id
	FROM match_players
	WHERE match_id = $1`

	rows, err := ds.db.Query(playersQ, matchID.String())
	if err != nil {
		return nil, fmt.Errorf("GetMatch: querying players: %w", err)
	}
	defer rows.Close()

	var players []int64
	for rows.Next() {
		var pid int64
		if err := rows.Scan(&pid); err != nil {
			return nil, fmt.Errorf("GetMatch: scanning player: %w", err)
		}
		players = append(players, pid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetMatch: players rows error: %w", err)
	}

	subs, err := ds.GetMatchSubmissions(matchID)
	if err != nil {
		return nil, fmt.Errorf("GetMatch: fetching submissions: %w", err)
	}

	return &models.Session{
		ID:          id,
		Problem:     models.Problem{ID: probID, Name: probName, Slug: probSlug, Difficulty: parsedDiff},
		IsRated:     isRated,
		Status:      parsedStatus,
		Winner:      winnerID,
		StartTime:   startTime,
		EndTime:     endTime,
		Players:     players,
		Submissions: subs,
	}, nil
}

// Returns match information (EXCEPT SUBMISSIONS, MUST GET SEPARATELY)
func (ds *dataStore) GetPlayerMatches(userID int64) ([]models.Session, error) {
	query := `
	SELECT 
		m.id, 
		m.problem_id, 
		p.name, 
		p.slug, 
		p.difficulty, 
		m.is_rated, 
		m.status, 
		m.winner_id, 
		m.start_time, 
		m.end_time,
		ARRAY_AGG(mp2.player_id) AS player_ids
	FROM match_players mp
	JOIN matches m ON mp.match_id = m.id
	JOIN problems p ON m.problem_id = p.id
	JOIN match_players mp2 ON mp2.match_id = m.id
	WHERE mp.player_id = $1
	GROUP BY m.id, m.problem_id, p.name, p.slug, p.difficulty, m.is_rated, m.status, m.winner_id, m.start_time, m.end_time
	ORDER BY m.start_time DESC`

	rows, err := ds.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetPlayerMatches: %w", err)
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var id string
		var probID int
		var probName string
		var probSlug string
		var probDifficulty string
		var status string
		var isRated bool
		var winnerID int64
		var startTime time.Time
		var endTime time.Time
		var playerIDs pq.Int64Array

		err = rows.Scan(&id, &probID, &probName, &probSlug, &probDifficulty,
			&isRated, &status, &winnerID, &startTime, &endTime, &playerIDs)
		if err != nil {
			return nil, fmt.Errorf("GetPlayerMatches scan: %w", err)
		}

		parsedStatus, err := models.ParseMatchStatus(status)
		if err != nil {
			return nil, fmt.Errorf("GetPlayerMatches parse: %w", err)
		}

		parsedDifficulty, err := models.ParseDifficulty(probDifficulty)
		if err != nil {
			return nil, fmt.Errorf("GetPlayerMatches parse: %w", err)
		}

		sesh := models.Session{
			ID:          id,
			Status:      parsedStatus,
			IsRated:     isRated,
			Problem:     models.Problem{ID: probID, Name: probName, Slug: probSlug, Difficulty: parsedDifficulty},
			Players:     playerIDs,
			Submissions: nil, // Do not populate submissions
			Winner:      winnerID,
			StartTime:   startTime,
			EndTime:     endTime,
		}
		sessions = append(sessions, sesh)
	}
	return sessions, nil
}

// Returns all submissions for a given match.
func (ds *dataStore) GetMatchSubmissions(matchID uuid.UUID) ([]models.PlayerSubmission, error) {
	query := `
	SELECT submission_id, player_id, passed_test_cases, total_test_cases, 
		status, runtime, memory, lang, submitted_at
	FROM submissions
	WHERE match_id = $1`

	rows, err := ds.db.Query(query, matchID.String())
	if err != nil {
		return nil, fmt.Errorf("GetMatchSubmissions: %w", err)
	}
	defer rows.Close()

	var submissions []models.PlayerSubmission
	for rows.Next() {
		var id int
		var playerID int64
		var passedTestCases int
		var totalTestCases int
		var status string
		var runtime int
		var memory int
		var lang string
		var submittedAt time.Time

		err = rows.Scan(&id, &playerID, &passedTestCases, &totalTestCases, &status, &runtime, &memory, &lang, &submittedAt)
		if err != nil {
			return nil, fmt.Errorf("GetMatchSubmissions scan: %w", err)
		}

		parsedStatus, err := models.ParseSubmissionStatus(status)
		if err != nil {
			return nil, fmt.Errorf("GetMatchSubmissions parse: %w", err)
		}

		parsedLang, err := models.ParseLang(lang)
		if err != nil {
			return nil, fmt.Errorf("GetMatchSubmissions parse: %w", err)
		}

		submission := models.PlayerSubmission{
			ID:              id,
			PlayerID:        playerID,
			PassedTestCases: passedTestCases,
			TotalTestCases:  totalTestCases,
			Status:          parsedStatus,
			Runtime:         runtime,
			Memory:          memory,
			Lang:            parsedLang,
			Time:            submittedAt,
		}
		submissions = append(submissions, submission)
	}
	return submissions, nil
}
