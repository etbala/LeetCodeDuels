package cache

import "time"

// SessionData is just an example—use whatever your game sessions actually look like.
type SessionData struct {
	Player1ID int64
	Player2ID int64
	StartedAt time.Time
	// …
}

// 5 Min Expire Time
type InviteData struct {
	InviterID int64
	InviteeID int64
	StartTime time.Time
}

type ConnectionData struct {
}

// 5 Min Expire Time
type StateData struct {
	UserID int64
	State  string
}
