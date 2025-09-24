// All structs for responses (API outputs) live here
package handlers

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
}
