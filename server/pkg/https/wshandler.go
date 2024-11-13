package https

import (
	"encoding/json"
	"leetcodeduels/api/auth"
	"leetcodeduels/api/game"
	"leetcodeduels/internal/ws"
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
func handleMessage(playerID int64, msg ws.Message, gm *game.GameManager) {
	switch msg.Type {
	case ws.MessageTypeSubmission:
		var submissionPayload ws.SubmissionPayload
		err := json.Unmarshal(msg.Payload, &submissionPayload)
		if err != nil {
			sendError(playerID, "INVALID_PAYLOAD", "Invalid submission payload", gm)
			return
		}

		// Process the submission
		processSubmission(playerID, submissionPayload, gm)

	case ws.MessageTypeHeartbeat:
		// Handle heartbeat messages if implemented
		// Optionally send a pong or update last active time
	default:
		sendError(playerID, "UNKNOWN_TYPE", "Unknown message type", gm)
	}
}

// sendError sends an error message to the specified player
func sendError(playerID int64, code, message string, gm *game.GameManager) {
	errorPayload := ws.ErrorPayload{
		Code:    code,
		Message: message,
	}
	errorMessage := ws.Message{
		Type:    ws.MessageTypeError,
		Payload: ws.MarshalPayload(errorPayload),
	}
	gm.SendMessageToPlayer(playerID, errorMessage)
}

// processSubmission handles a player's code submission
func processSubmission(playerID int64, payload ws.SubmissionPayload, gm *game.GameManager) {
	status, err := game.ParseSubmissionStatus(payload.Status)
	if err != nil {
		// TODO: Handle invalid status
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

	// Add the submission to the game session
	gm.AddSubmission(playerID, submission)

	// Optionally, notify the opponent
	sessionID, err := gm.GetPlayersSessionID(playerID)
	if err != nil {
		sendError(playerID, "NOT_IN_SESSION", "Player not in any session", gm)
		return
	}

	opponentID, err := gm.GetOpponentID(sessionID, playerID)
	if err != nil {
		sendError(playerID, "OPPONENT_NOT_FOUND", "Opponent not found", gm)
		return
	}

	opponentSubmissionPayload := ws.OpponentSubmissionPayload{
		PlayerID: playerID,
		Status:   string(submission.Status),
	}

	opponentMessage := ws.Message{
		Type:    ws.MessageTypeOpponentSubmission,
		Payload: ws.MarshalPayload(opponentSubmissionPayload),
	}

	gm.SendMessageToPlayer(opponentID, opponentMessage)
}

// wsHandler handles WebSocket connection requests
func WsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract token from query parameters or headers
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
		return
	}
	defer conn.Close()

	// Register the connection with GameManager
	gm := game.GetGameManager()
	gm.AddPlayerConnection(claims.UserID, conn)
	defer gm.RemovePlayerConnection(claims.UserID)

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

		handleMessage(claims.UserID, msg, gm)
	}
}
