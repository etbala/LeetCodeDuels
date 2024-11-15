package presence

// Connection Manager

import (
	"sync"
	"time"
)

type UserStatus struct {
	IsOnline     bool
	InGame       bool
	LastActivity time.Time
}

type PresenceManager struct {
	sync.RWMutex
	Users map[int64]*UserStatus // Map of user ID to their status
}

var (
	instance *PresenceManager
	once     sync.Once
)

func GetPresenceManager() *PresenceManager {
	once.Do(func() {
		instance = &PresenceManager{
			Users: make(map[int64]*UserStatus),
		}
	})
	return instance
}

func (pm *PresenceManager) SetUserOnline(userID int64) {
	pm.Lock()
	pm.Users[userID] = &UserStatus{
		IsOnline:     true,
		InGame:       false,
		LastActivity: time.Now(),
	}
	pm.Unlock()
}

func (pm *PresenceManager) SetUserOffline(userID int64) {
	pm.Lock()
	pm.Users[userID].IsOnline = false
	pm.Users[userID].LastActivity = time.Now()
	pm.Unlock()
}

func (pm *PresenceManager) SetUserInGame(userID int64, inGame bool) {
	pm.Lock()
	if status, exists := pm.Users[userID]; exists {
		status.InGame = inGame
		status.LastActivity = time.Now()
	}
	pm.Unlock()
}

func (pm *PresenceManager) GetUserStatus(userID int64) *UserStatus {
	pm.RLock()
	status, exists := pm.Users[userID]
	pm.RUnlock()

	if !exists {
		return nil
	}
	return status
}
