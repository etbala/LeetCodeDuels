package models

import "time"

type User struct {
	ID               int64
	Username         string
	LeetCodeUsername string
	AccessToken      string
	CreatedAt        time.Time
	UpdatedAt        time.Time // (Last Logged In)
	Rating           int
}
