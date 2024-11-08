package matchmaking

import (
	"leetcodeduels/internal/enums"
	"time"
)

type PlayerInfo struct {
	ID           int64
	Difficulties []enums.Difficulty
	Tags         []int
	JoinedAt     time.Time
	ForceMatch   bool
}

type Match struct {
	Player1ID int64
	Player2ID int64
}
