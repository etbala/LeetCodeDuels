package main

import (
	"context"
	"flag"
	"leetcodeduels/api"
	"leetcodeduels/auth"
	"leetcodeduels/config"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"leetcodeduels/ws"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables from .env file (at root of server dir)
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	err = store.InitDataStore(cfg.DB_URL)
	if err != nil {
		log.Fatalf("Failed to Init DataStore: %v", err)
	}

	err = auth.InitStateStore(cfg.RDB_URL)
	if err != nil {
		log.Fatalf("Failed to Init StateStore: %v", err)
	}
	defer auth.StateStore.Close()

	err = ws.InitConnManager(cfg.RDB_URL)
	if err != nil {
		log.Fatalf("Failed to Init ConnManager: %v", err)
	}
	defer ws.ConnManager.Close()

	err = services.InitInviteManager(cfg.RDB_URL)
	if err != nil {
		log.Fatalf("Failed to Init InviteManager: %v", err)
	}
	defer services.InviteManager.Close()

	err = services.InitGameManager(cfg.RDB_URL)
	if err != nil {
		log.Fatalf("Failed to Init InviteManager: %v", err)
	}
	defer services.GameManager.Close()

	var port string
	flag.StringVar(&port, "port", "8080", "Server port to listen on")
	flag.Parse()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://leetcode.com", "http://127.0.0.1"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	router := api.SetupRoutes(auth.Middleware)
	handler := c.Handler(router)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s\n", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s", err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	// Shutdown the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down server...")

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
