package server

import (
	"context"
	"leetcodeduels/api"
	"leetcodeduels/config"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"leetcodeduels/ws"
	"net/http"
	"time"

	"github.com/rs/cors"
)

func New(cfg *config.Config) (*http.Server, error) {
	err := store.InitDataStore(cfg.DB_URL)
	if err != nil {
		return nil, err
	}

	err = services.InitInviteManager(cfg.RDB_URL)
	if err != nil {
		return nil, err
	}

	err = services.InitGameManager(cfg.RDB_URL)
	if err != nil {
		return nil, err
	}

	err = ws.InitConnManager(cfg.RDB_URL)
	if err != nil {
		return nil, err
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://leetcode.com", "http://127.0.0.1"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	router := api.SetupRoutes(services.Middleware)
	handler := c.Handler(router)

	srv := &http.Server{
		Addr:    ":" + cfg.PORT,
		Handler: handler,
	}
	return srv, nil
}

func Cleanup(srv *http.Server) error {
	// shut down HTTP server first
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	services.InviteManager.Close()
	services.GameManager.Close()
	ws.ConnManager.Close()

	return nil
}
