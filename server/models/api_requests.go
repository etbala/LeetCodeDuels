// Request structs for API endpoints
package models

type UpdateUserRequest struct {
	Username         string `json:"username,omitempty"`
	LeetCodeUsername string `json:"lc_username,omitempty"`
}
