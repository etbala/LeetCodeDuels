package https

import (
	"encoding/json"
	"leetcodeduels/api/auth"
	"leetcodeduels/api/game"
	"leetcodeduels/internal/ws"
	"leetcodeduels/pkg/connections"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Define WebSocket upgrader with appropriate settings
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for now; adjust in production
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleMessage routes incoming messages to appropriate handlers
func handleMessage(playerID int64, msg ws.Message, cm *connections.ConnectionManager) {
	switch msg.Type {
	case ws.MessageTypeSubmission:
		var submissionPayload ws.SubmissionPayload
		err := json.Unmarshal(msg.Payload, &submissionPayload)
		if err != nil {
			sendError(playerID, "INVALID_PAYLOAD", "Invalid submission payload")
			return
		}

		// Process the submission
		processSubmission(playerID, submissionPayload)

	case ws.MessageTypeHeartbeat:
		// Update last activity time
		cm.UpdateLastActivity(playerID)
	default:
		sendError(playerID, "UNKNOWN_TYPE", "Unknown message type")
	}
}

// sendError sends an error message to the specified player
func sendError(playerID int64, code, message string) {
	errorPayload := ws.ErrorPayload{
		Code:    code,
		Message: message,
	}
	errorMessage := ws.Message{
		Type:    ws.MessageTypeError,
		Payload: ws.MarshalPayload(errorPayload),
	}

	messageBytes, err := json.Marshal(errorMessage)
	if err != nil {
		log.Printf("Failed to marshal error message: %v", err)
		return
	}

	cm := connections.GetConnectionManager()
	err = cm.SendMessageToUser(playerID, websocket.TextMessage, messageBytes)
	if err != nil {
		log.Printf("Failed to send error message to player %d: %v", playerID, err)
	}
}

// processSubmission handles a player's code submission
func processSubmission(playerID int64, payload ws.SubmissionPayload) {
	status, err := game.ParseSubmissionStatus(payload.Status)
	if err != nil {
		sendError(playerID, "INVALID_STATUS", "Invalid submission status")
		return
	}

	// Create a PlayerSubmission object
	submission := game.PlayerSubmission{
		ID:              payload.ID,
		PlayerID:        playerID,
		PassedTestCases: payload.PassedTestCases,
		TotalTestCases:  payload.TotalTestCases,
		Status:          status,
		Runtime:         payload.Runtime,
		Memory:          payload.Memory,
		Time:            time.Now(),
	}

	gm := game.GetGameManager()

	// Add the submission to the game session
	err = gm.AddSubmission(playerID, submission)
	if err != nil {
		sendError(playerID, "ADD_SUBMISSION_ERROR", err.Error())
		return
	}
}

// wsHandler handles WebSocket connection requests
func WsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract token from query params
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		// Alternatively, extract from headers
		tokenString = r.Header.Get("Authorization")
		if tokenString == "" {
			log.Println("Unauthorized: No token provided")
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}
		// If using "Bearer " prefix in headers
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}
	}

	// Validate the token and extract claims
	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	// Upgrade the HTTP connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		http.Error(w, "Error: Could not upgrade to websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// Get the ConnectionManager instance
	cm := connections.GetConnectionManager()

	// Register the connection with ConnectionManager
	cm.AddConnection(claims.UserID, conn)
	defer cm.RemoveConnection(claims.UserID)

	// Listen for incoming messages
	for {
		var msg ws.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		handleMessage(claims.UserID, msg, cm)
	}
}
