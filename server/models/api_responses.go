// Response structs for API endpoints
package models

import "time"

type TokenExchangeResponse struct {
	Token string           `json:"token"`
	User  UserInfoResponse `json:"user"`
}

type UserInfoResponse struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	LCUsername    string `json:"lc_username"`
	AvatarURL     string `json:"avatar_url"`
	Rating        int    `json:"rating"`
}

type UserStatusResponse struct {
	Online bool   `json:"online"`
	InGame bool   `json:"in_game"`
	GameID string `json:"game_id,omitempty"` // If InGame, include current game ID
}

type InviteNotification struct {
	FromUser     UserInfoResponse `json:"from_user"`
	MatchDetails MatchDetails     `json:"matchDetails"`
	CreatedAt    time.Time        `json:"createdAt"`
}

type NotificationsResponse struct {
	Invites []InviteNotification `json:"invites"`
}

type QueueSizeResponse struct {
	Size int `json:"size"`
}
