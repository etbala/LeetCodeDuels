package models

import "time"

type User struct {
	ID               int64     `json:"userID"`
	Username         string    `json:"username"`
	LeetCodeUsername string    `json:"leetcodeUsername"`
	AccessToken      string    `json:"accessToken"`
	CreatedAt        time.Time `json:"creationDate"`
	UpdatedAt        time.Time `json:"lastOnline"` // (Last Logged In)
	Rating           int       `json:"rating"`
}
