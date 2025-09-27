// Request structs for API endpoints
package models

type UpdateUserRequest struct {
	Username         string `json:"username,omitempty"`
	LeetCodeUsername string `json:"leet_code_username,omitempty"`
}
