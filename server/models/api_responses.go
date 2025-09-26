// Response structs for API endpoints
package models

type TokenExchangeResponse struct {
	Token string                `json:"token"`
	User  UserClientInformation `json:"user"`
}

type UserClientInformation struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	LCUsername    string `json:"lc_username"`
	AvatarURL     string `json:"avatar_url"`
	Rating        int    `json:"rating"`
}
