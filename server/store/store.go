package store

import (
	"database/sql"
	"fmt"
	"leetcodeduels/models"
	"strings"
	"time"

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
	query := `SELECT github_id, username, lc_username, access_token, created_at, updated_at, rating
			FROM github_oauth_users WHERE github_id = $1`
	row := ds.db.QueryRow(query, githubID)
	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.LeetCodeUsername, &u.AccessToken, &u.CreatedAt, &u.UpdatedAt, &u.Rating); err != nil {
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
	err := ds.db.QueryRow(query, userID).Scan(&rating)
	if err != nil {
		return 0, fmt.Errorf("GetUserRating: %w", err)
	}
	return rating, nil
}

func (ds *dataStore) UpdateUsername(userID int64, newUsername string) error {
	query := `UPDATE github_oauth_users SET username = $1 WHERE github_id = $2`
	_, err := ds.db.Exec(query, newUsername, userID)
	if err != nil {
		return fmt.Errorf("UpdateUsername: %w", err)
	}
	return nil
}

func (ds *dataStore) UpdateLCUsername(userID int64, newLCUsername string) error {
	query := `UPDATE github_oauth_users SET lc_username = $1 WHERE github_id = $2`
	_, err := ds.db.Exec(query, newLCUsername, userID)
	if err != nil {
		return fmt.Errorf("UpdateLCUsername: %w", err)
	}
	return nil
}

func (ds *dataStore) UpdateUserRating(userID int64, newRating int) error {
	query := `UPDATE github_oauth_users SET rating = $1 WHERE github_id = $2`
	_, err := ds.db.Exec(query, newRating, userID)
	if err != nil {
		return fmt.Errorf("UpdateUserRating: %w", err)
	}
	return nil
}

func (ds *dataStore) DeleteUser(userID int64) error {
	query := `DELETE FROM github_oauth_users WHERE github_id = $1`
	_, err := ds.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("DeleteUser: %w", err)
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

// GetRandomProblemForDuel picks a random problem matching preferences of both players and a set of difficulties.
func (ds *dataStore) GetRandomProblemMatchmaking(
	player1Tags, player2Tags []int, difficulties []models.Difficulty,
) (*models.Problem, error) {
	// Build IN clauses dynamically
	if len(difficulties) == 0 {
		difficulties = []models.Difficulty{models.Easy, models.Medium, models.Hard}
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
		return nil, fmt.Errorf("GetRandomProblemMatchmaking: %w", err)
	}
	return &p, nil
}

func (ds *dataStore) GetRandomProblemDuel(
	tags []int, difficulties []models.Difficulty,
) (*models.Problem, error) {
	if len(difficulties) == 0 {
		difficulties = []models.Difficulty{models.Easy, models.Medium, models.Hard}
	}

	args := []interface{}{pq.Array(difficulties)}

	tagSubquery := ``
	if len(tags) > 0 {
		args = append(args, pq.Array(tags))
		tagSubquery = `
		AND EXISTS (
			SELECT 1 
			FROM problem_tags pt 
			WHERE pt1.problem_id = p.id 
			  AND pt.tag_id = ANY($2)
		)`
	}

	query := fmt.Sprintf(`
	SELECT p.id, p.name, p.slug, p.difficulty
	FROM problems p
	WHERE p.difficulty = ANY($1)
	%s
	ORDER BY RANDOM() LIMIT 1`,
		tagSubquery)

	var p models.Problem
	if err := ds.db.QueryRow(query, args...).Scan(&p.ID, &p.Name, &p.Slug, &p.Difficulty); err != nil {
		return nil, fmt.Errorf("GetRandomProblemDuel: %w", err)
	}
	return &p, nil
}

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

// TODO: Investigate if triple query or single query with ARRAY_AGG is better
func (ds *dataStore) GetMatch(matchID string) (*models.Session, error) {
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
	err := ds.db.QueryRow(matchQ, matchID).
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

	rows, err := ds.db.Query(playersQ, matchID)
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

func (ds *dataStore) GetMatchSubmissions(matchID string) ([]models.PlayerSubmission, error) {
	query := `
	SELECT submission_id, player_id, passed_test_cases, total_test_cases, 
		status, runtime, memory, lang, submitted_at
	FROM submissions
	WHERE match_id = $1`

	rows, err := ds.db.Query(query, matchID)
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
		var time time.Time

		err = rows.Scan(&id, &playerID, &passedTestCases, &totalTestCases, &status, &runtime, &memory, &lang, &time)
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
			Time:            time,
		}
		submissions = append(submissions, submission)
	}
	return submissions, nil
}
