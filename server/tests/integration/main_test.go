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
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	ts            *httptest.Server
	pool          *dockertest.Pool
	pgResource    *dockertest.Resource
	redisResource *dockertest.Resource
)

func mustMigrate(dbURL string) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(fmt.Sprintf("migrate: open db: %v", err))
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(fmt.Sprintf("migrate: driver init: %v", err))
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://../migrations",
		"postgres",
		driver,
	)
	if err != nil {
		panic(fmt.Sprintf("migrate: new instance: %v", err))
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(fmt.Sprintf("migrate: up failed: %v", err))
	}
}

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
	os.Setenv("LOG_LEVEL", "error")               // only log errors during tests
	os.Setenv("SUBMISSION_VALIDATION", "disable") // don't query leetcode during tests

	// Migrations (Create Tables)
	cfg, _ := config.InitConfig()
	mustMigrate(cfg.DB_URL)

	// Start Server
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("could not init server: %s", err)
	}
	ts = httptest.NewServer(srv.Handler)
	defer ts.Close()

	// Run tests
	code := m.Run()

	// Close connections before killing databases
	server.Cleanup(srv)

	// Teardown Docker resources
	pool.Purge(pgResource)
	pool.Purge(redisResource)

	os.Exit(code)
}
