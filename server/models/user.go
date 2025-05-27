package models

import "time"

type User struct {
	ID               int64     `json:"userID"`
	Username         string    `json:"username"`
	LeetCodeUsername string    `json:"lcUsername"`
	AccessToken      string    `json:"token"`
	CreatedAt        time.Time `json:"startDate"`
	UpdatedAt        time.Time `json:"lastOnline"` // (Last Logged In)
	Rating           int       `json:"rating"`
}
