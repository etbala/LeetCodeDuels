package tests

import (
	"context"
	"database/sql"
	"fmt"
	"leetcodeduels/config"
	"leetcodeduels/server"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
)

var ts *httptest.Server
var pool *dockertest.Pool
var pgResource, redisResource *dockertest.Resource

func TestMain(m *testing.M) {
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to Docker: %s", err)
	}

	// Start Postgres
	pgResource, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13-alpine",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_DB=testdb",
		},
	})
	if err != nil {
		log.Fatalf("could not start postgres: %s", err)
	}

	// Exponential retry to wait for Postgres to be ready
	err = pool.Retry(func() error {
		db, err := sql.Open("postgres", fmt.Sprintf(
			"host=localhost port=%s user=postgres password=secret dbname=testdb sslmode=disable",
			pgResource.GetPort("5432/tcp"),
		))
		if err != nil {
			return err
		}
		return db.Ping()
	})
	if err != nil {
		log.Fatalf("could not connect to postgres: %s", err)
	}

	// Start Redis
	redisResource, err = pool.Run("redis", "latest", nil)
	if err != nil {
		log.Fatalf("could not start redis: %s", err)
	}
	if err = pool.Retry(func() error {
		rdb := redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", redisResource.GetPort("6379/tcp")),
		})
		return rdb.Ping(context.Background()).Err()
	}); err != nil {
		log.Fatalf("could not connect to redis: %s", err)
	}

	os.Setenv("DB_URL", fmt.Sprintf(
		"postgres://postgres:secret@localhost:%s/testdb?sslmode=disable",
		pgResource.GetPort("5432/tcp"),
	))
	os.Setenv("RDB_URL", fmt.Sprintf(
		"redis://localhost:%s",
		redisResource.GetPort("6379/tcp"),
	))
	os.Setenv("PORT", "8765")
	os.Setenv("JWT_SECRET", "0")

	// Migrations (Create Tables)
	cfg, _ := config.LoadConfig()
	mustMigrate(cfg.DB_URL, "../migrations")

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("could not init server: %s", err)
	}
	ts = httptest.NewServer(srv.Handler) // uses the Handler you wired up
	defer ts.Close()

	// 6) Run tests
	code := m.Run()

	// 7) Teardown Docker resources
	if err := pool.Purge(pgResource); err != nil {
		log.Printf("could not purge postgres: %s", err)
	}
	if err := pool.Purge(redisResource); err != nil {
		log.Printf("could not purge redis: %s", err)
	}

	os.Exit(code)
}

func mustMigrate(dbURL, migrationsPath string) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("open db for migrations: %v", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver,
	)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}
	// ErrNoChange is fine if already up to date
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrate up: %v", err)
	}
}
