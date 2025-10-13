package main

import (
	"context"
	"leetcodeduels/config"
	"leetcodeduels/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	// Load environment variables from .env file (at root of server dir)
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Info: .env file not found, relying on system environment variables.")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config loading failed: %v", err)
	}

	setLogLevel(cfg.LOG_LEVEL)

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Server initialization failed: %v", err)
	}

	// Start server in goroutine
	go func() {
		addr := ":" + cfg.PORT
		log.Printf("Listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests up to 5s to complete, then clean up everything
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Cleanup(srv)
	if err != nil {
		log.Fatalf("Server cleanup failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}

func setLogLevel(level string) {
	switch strings.ToLower(level) {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "info":
		fallthrough
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
