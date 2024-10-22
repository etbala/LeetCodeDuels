package auth

/*
Singleton that store randomly generated states when logging
in via OAuth, along with their associated expiration times.
*/

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

type stateEntry struct {
	state     string
	expiresAt time.Time
}

type StateStore struct {
	states map[string]stateEntry
	mutex  sync.Mutex
}

var instance *StateStore
var once sync.Once

func GetStateStore() *StateStore {
	once.Do(func() {
		instance = &StateStore{
			states: make(map[string]stateEntry),
		}
	})
	return instance
}

func (s *StateStore) StartCleanupRoutine() {
	go s.cleanupExpiredStates()
}

// TODO: More clear way to change state expiration
//		 time and cleanup interval

func (s *StateStore) GenerateRandomState() (string, error) {
	// Create a secure random state value (128 bits of entropy)
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.URLEncoding.EncodeToString(b)
	expirationTime := time.Now().Add(5 * time.Minute)

	s.mutex.Lock()
	s.states[state] = stateEntry{
		state:     state,
		expiresAt: expirationTime,
	}
	s.mutex.Unlock()

	return state, nil
}

func (s *StateStore) ValidateState(state string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	entry, exists := s.states[state]
	if !exists {
		return errors.New("state not found or already used")
	}

	if time.Now().After(entry.expiresAt) {
		delete(s.states, state)
		return errors.New("state expired")
	}

	delete(s.states, state)
	return nil
}

func (s *StateStore) cleanupExpiredStates() {
	for {
		time.Sleep(60 * time.Minute)

		s.mutex.Lock()
		for state, entry := range s.states {
			if time.Now().After(entry.expiresAt) {
				delete(s.states, state)
			}
		}
		s.mutex.Unlock()
	}
}
