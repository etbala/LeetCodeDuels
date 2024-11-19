package connections

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type UserStatus struct {
	InGame       bool
	LastActivity time.Time
}

// ConnectionManager manages user WebSocket connections and usernames.
type ConnectionManager struct {
	sync.RWMutex
	UserConnections map[int64]*websocket.Conn // Maps user ID to their WebSocket connection
	UserStatus      map[int64]*UserStatus     // Maps user ID to their current status
}

var (
	instance *ConnectionManager
	once     sync.Once
)

func GetConnectionManager() *ConnectionManager {
	once.Do(func() {
		instance = &ConnectionManager{
			UserConnections: make(map[int64]*websocket.Conn),
			UserStatus:      make(map[int64]*UserStatus),
		}
	})
	return instance
}

// AddConnection adds or updates a user's connection.
// Does not reset InGame status if user is already in a game.
func (cm *ConnectionManager) AddConnection(userID int64, conn *websocket.Conn) {
	cm.Lock()
	defer cm.Unlock()

	cm.UserConnections[userID] = conn

	// Initialize UserStatus if not already present
	if _, exists := cm.UserStatus[userID]; !exists {
		cm.UserStatus[userID] = &UserStatus{InGame: false}
	}
}

// RemoveConnection handles disconnection of a user.
// Sets connection to nil if the user is in-game; otherwise, removes the user entirely.
func (cm *ConnectionManager) RemoveConnection(userID int64) {
	cm.Lock()
	defer cm.Unlock()

	conn, exists := cm.UserConnections[userID]
	if exists && conn != nil {
		conn.Close()
	}

	if cm.IsUserInGame(userID) {
		// User is in-game, set their connection to nil but keep them in the map
		cm.UserConnections[userID] = nil
	} else {
		// User is not in-game, remove them entirely
		delete(cm.UserConnections, userID)
		delete(cm.UserStatus, userID)
	}
}

// GetConnection retrieves a user's WebSocket connection.
func (cm *ConnectionManager) GetConnection(userID int64) (*websocket.Conn, bool) {
	cm.RLock()
	conn, exists := cm.UserConnections[userID]
	cm.RUnlock()
	return conn, exists && conn != nil
}

// IsUserOnline checks if a user is online.
func (cm *ConnectionManager) IsUserOnline(userID int64) bool {
	cm.RLock()
	conn, exists := cm.UserConnections[userID]
	cm.RUnlock()
	return exists && conn != nil
}

// SetUserInGame updates a user's in-game status.
func (cm *ConnectionManager) SetUserInGame(userID int64, inGame bool) {
	cm.Lock()
	if status, exists := cm.UserStatus[userID]; exists {
		status.InGame = inGame
	} else {
		cm.UserStatus[userID] = &UserStatus{InGame: inGame}
	}
	cm.Unlock()
}

// IsUserInGame checks if a user is currently in a game.
func (cm *ConnectionManager) IsUserInGame(userID int64) bool {
	cm.RLock()
	status, exists := cm.UserStatus[userID]
	cm.RUnlock()
	return exists && status.InGame
}

func (cm *ConnectionManager) UpdateLastActivity(userID int64) {
	cm.Lock()
	defer cm.Unlock()
	_, exists := cm.UserStatus[userID]
	if !exists {
		return
	}

	cm.UserStatus[userID].LastActivity = time.Now()
}

// CheckConnections checks the health of all connections and removes inactive ones.
// For users in-game, sets their connections to nil instead of removing them.
func (cm *ConnectionManager) CheckConnections() {
	cm.Lock()
	defer cm.Unlock()

	for userID, conn := range cm.UserConnections {
		if conn != nil {
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				// Connection is inactive
				conn.Close()
				if cm.IsUserInGame(userID) {
					cm.UserConnections[userID] = nil
				} else {
					delete(cm.UserConnections, userID)
					delete(cm.UserStatus, userID)
				}
			}
		} else if conn == nil {
			// Cleanup closed connections that are no longer in game
			if cm.UserStatus[userID].InGame == false {
				delete(cm.UserConnections, userID)
				delete(cm.UserStatus, userID)
			}
		}
	}
}

func (cm *ConnectionManager) SendMessageToUser(userID int64, messageType int, message []byte) error {
	cm.RLock()
	conn, exists := cm.UserConnections[userID]
	cm.RUnlock()
	if !exists || conn == nil {
		return fmt.Errorf("user %d is not connected", userID)
	}

	cm.Lock()
	defer cm.Unlock()
	if err := conn.WriteMessage(messageType, message); err != nil {
		// Handle disconnection during send
		conn.Close()
		if cm.IsUserInGame(userID) {
			cm.UserConnections[userID] = nil
		} else {
			delete(cm.UserConnections, userID)
			delete(cm.UserStatus, userID)
		}
		return err
	}
	return nil
}

// BroadcastMessage sends a message to all online users.
func (cm *ConnectionManager) BroadcastMessage(messageType int, message []byte) {
	cm.RLock()
	userIDs := make([]int64, 0, len(cm.UserConnections))
	for userID := range cm.UserConnections {
		userIDs = append(userIDs, userID)
	}
	cm.RUnlock()

	for _, userID := range userIDs {
		_ = cm.SendMessageToUser(userID, messageType, message)
	}
}

// GetOnlineStatistics returns (# people online, # people in-game)
func (cm *ConnectionManager) GetOnlineStatistics() (int64, int64) {
	cm.RLock()

	var onlineCount, inGameCount int64

	for userID, conn := range cm.UserConnections {
		if conn != nil {
			onlineCount++
		}
		if status, exists := cm.UserStatus[userID]; exists && status.InGame {
			inGameCount++
		}
	}
	cm.RUnlock()

	return onlineCount, inGameCount
}
