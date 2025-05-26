package server

import (
	"leetcodeduels/api"
	"leetcodeduels/auth"
	"leetcodeduels/config"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"leetcodeduels/ws"
	"net/http"

	"github.com/rs/cors"
)

func New(cfg *config.Config, port string) (*http.Server, error) {
	err := store.InitDataStore(cfg.DB_URL)
	if err != nil {
		return nil, err
	}

	err = auth.InitStateStore(cfg.RDB_URL)
	if err != nil {
		return nil, err
	}
	defer auth.StateStore.Close()

	err = ws.InitConnManager(cfg.RDB_URL)
	if err != nil {
		return nil, err
	}
	defer ws.ConnManager.Close()

	err = services.InitInviteManager(cfg.RDB_URL)
	if err != nil {
		return nil, err
	}
	defer services.InviteManager.Close()

	err = services.InitGameManager(cfg.RDB_URL)
	if err != nil {
		return nil, err
	}
	defer services.GameManager.Close()

	ws.InitPubSub()

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
	return srv, nil
}
