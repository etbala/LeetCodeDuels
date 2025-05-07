package cache

import "time"

// SessionData is just an example—use whatever your game sessions actually look like.
type SessionData struct {
	Player1ID int64
	Player2ID int64
	StartedAt time.Time
	// …
}
