package handlers

import (
	"fmt"
	"leetcodeduels/auth"
	"leetcodeduels/ws"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WSConnect(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.ExtractTokenString(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "could not upgrade to ws", http.StatusBadRequest)
		return
	}

	cm, err := ws.GetConnectionManager("")
	if err != nil {
		http.Error(w, "could not access connection manager", http.StatusInternalServerError)
		return
	}

	userID := claims.UserID
	connID := fmt.Sprintf("%p", conn)
	oldConnID, err := cm.AddConnection(userID, connID)
	if err != nil {
		conn.Close()
		return
	}

	if oldConnID != "" {
		ws.PublishDisconnect(cm, userID, oldConnID)
	}

	sendCh := make(chan []byte, 256)
	go ws.ReadLoop(cm, userID, connID, conn, sendCh)
	go ws.WriteLoop(cm, userID, connID, conn, sendCh)
}
