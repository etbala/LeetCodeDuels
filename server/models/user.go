package models

import "time"

type User struct {
	ID               int64     `json:"userID"`
	AccessToken      string    `json:"token"` // GitHub OAuth Token (SERVER SIDE ONLY)
	Username         string    `json:"username"`
	Discriminator    string    `json:"discriminator"`
	LeetCodeUsername string    `json:"lcUsername"`
	AvatarURL        string    `json:"avatarUrl"`
	CreatedAt        time.Time `json:"startDate"`
	UpdatedAt        time.Time `json:"lastOnline"`
	Rating           int       `json:"rating"`
}
