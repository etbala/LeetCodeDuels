package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"leetcodeduels/config"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

type DB struct {
	Postgres *sql.DB
	Redis    *redis.Client
}

// NewDB initializes both Postgres and Redis connections
func NewDB() (*DB, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Set up Postgres connection
	pgDB, err := sql.Open("postgres", cfg.DB_URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
	}

	// It's good practice to verify the connection
	if err = pgDB.Ping(); err != nil {
		pgDB.Close()
		return nil, fmt.Errorf("failed to ping Postgres: %w", err)
	}

	log.Println("Connected to Postgres successfully.")

	// Set up Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.DB_URL, // For example: "localhost:6379"
	})

	// Verify Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		pgDB.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Connected to Redis successfully.")

	return &DB{
		Postgres: pgDB,
		Redis:    rdb,
	}, nil
}

// Close closes both Postgres and Redis connections
func (db *DB) Close() error {
	var errs []error
	if err := db.Postgres.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close Postgres: %w", err))
	}
	if err := db.Redis.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close Redis: %w", err))
	}

	if len(errs) > 0 {
		// Combine errors if multiple occurred
		msg := "multiple errors occurred while closing DB connections:\n"
		for _, e := range errs {
			msg += "- " + e.Error() + "\n"
		}
		return fmt.Errorf(msg)
	}
	return nil
}

/*
	Redis for Session Information:

		Set TTL to expire after 24 hrs

		Key: "match:<match_id>"
		Value: {
			"player_one": 1,
			"player_two": 2,
			"problem_id": 42,
			"player_one_submissions": [],
			"player_two_submissions": [],
		}

		Key: "ws_connections:<match_id>"
		Value: [
			"conn1", "conn2"
		]

		Key: "user_session:<user_id>"
		Value: {
			"match_id": "12345"
		}

	Postgres for long term storage
	- Accounts
	- Problem Metadata
	- Match History (Aborted/Completed Matches)

*/
